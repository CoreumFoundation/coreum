package apps

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
	"github.com/CoreumFoundation/coreum/coreznet/pkg/retry"
	"github.com/CoreumFoundation/coreum/coreznet/pkg/rnd"
)

// CoredType is the type of cored application
const CoredType infra.AppType = "cored"

// well-known keys to create predictable wallets so manual operation is easier
var (
	// alice's address: cosmos1x645ym2yz4gckqjtpwr8yddqzkkzdpktxr3clr
	alicePrivKey = cored.Secp256k1PrivateKey{0x9b, 0xc9, 0xd0, 0x15, 0x11, 0x2, 0x94, 0x9, 0x92, 0xfd, 0x2b, 0xad, 0xbe, 0x36, 0x63, 0x1f, 0xed, 0x30, 0x10, 0xd, 0x6e, 0x24, 0xb1, 0xc2, 0x58, 0xb4, 0xfd, 0xe4, 0xaf, 0xdd, 0xf2, 0x40}
	// bob's address: cosmos1cjs7qela0trw2qyyfxw5e5e7cvwzprkju5d2su
	bobPrivKey = cored.Secp256k1PrivateKey{0x87, 0x70, 0x0, 0x22, 0xa3, 0x24, 0x81, 0x59, 0x8d, 0xb8, 0x27, 0x57, 0xdb, 0x97, 0xe6, 0x9b, 0xed, 0x11, 0xb6, 0x17, 0x3, 0xcc, 0x44, 0xe0, 0x2a, 0xd3, 0x1e, 0x95, 0x36, 0xcf, 0x2d, 0x7f}
	// charlie's address: cosmos1rd8wynz2987ey6pwmkuwfg9q8hf04xdyjqy2f4
	charliePrivKey = cored.Secp256k1PrivateKey{0x12, 0x9, 0x56, 0x3d, 0x40, 0x69, 0xf7, 0x57, 0xdd, 0x4c, 0x69, 0x17, 0x92, 0x7, 0xf0, 0xe6, 0x62, 0xa1, 0xcb, 0x8c, 0xfe, 0x8, 0x61, 0x68, 0x4c, 0x5e, 0xbc, 0x6b, 0x34, 0xa9, 0x5f, 0x7}
)

// NewCored creates new cored app
func NewCored(name string, config infra.Config, genesis *cored.Genesis, executor cored.Executor, appInfo *infra.AppInfo, ports cored.Ports, rootNode *Cored) Cored {
	nodePublicKey, nodePrivateKey, err := ed25519.GenerateKey(rand.Reader)
	must.OK(err)
	validatorPublicKey, validatorPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	must.OK(err)

	stakerPubKey, stakerPrivKey := cored.GenerateSecp256k1Key()

	genesis.AddWallet(stakerPubKey, "100000000000000000000000core,10000000000000000000000000stake")
	genesis.AddValidator(validatorPublicKey, stakerPrivKey, "100000000stake")

	if rootNode == nil {
		genesis.AddWallet(alicePrivKey.PubKey(), "1000000000000000core")
		genesis.AddWallet(bobPrivKey.PubKey(), "1000000000000000core")
		genesis.AddWallet(charliePrivKey.PubKey(), "1000000000000000core")
	}

	return Cored{
		name:                name,
		config:              config,
		executor:            executor,
		nodeID:              cored.NodeID(nodePublicKey),
		nodePrivateKey:      nodePrivateKey,
		validatorPrivateKey: validatorPrivateKey,
		genesis:             genesis,
		appInfo:             appInfo,
		ports:               ports,
		rootNode:            rootNode,
		mu:                  &sync.RWMutex{},
		walletKeys: map[string]cored.Secp256k1PrivateKey{
			"staker":  stakerPrivKey,
			"alice":   alicePrivKey,
			"bob":     bobPrivKey,
			"charlie": charliePrivKey,
		},
	}
}

// Cored represents cored
type Cored struct {
	name                string
	config              infra.Config
	executor            cored.Executor
	nodeID              string
	nodePrivateKey      ed25519.PrivateKey
	validatorPrivateKey ed25519.PrivateKey
	genesis             *cored.Genesis
	appInfo             *infra.AppInfo
	ports               cored.Ports
	rootNode            *Cored

	mu         *sync.RWMutex
	walletKeys map[string]cored.Secp256k1PrivateKey
}

// Name returns name of app
func (c Cored) Name() string {
	return c.name
}

// PeerAddress returns a string which might be passed to other node to connect to this peer
func (c Cored) PeerAddress() string {
	return fmt.Sprintf("%s@%s:%d", c.nodeID, c.IP(), c.ports.P2P)
}

// IP returns IP chain listens on
func (c Cored) IP() net.IP {
	return c.appInfo.IP()
}

// Status returns status of application
func (c Cored) Status() infra.AppStatus {
	return c.appInfo.Status()
}

// AddWallet adds wallet to genesis block and local keystore
func (c Cored) AddWallet(balances string) (cored.Wallet, cored.Secp256k1PrivateKey) {
	pubKey, privKey := cored.GenerateSecp256k1Key()
	c.genesis.AddWallet(pubKey, balances)

	c.mu.Lock()
	defer c.mu.Unlock()

	var name string
	for {
		name = rnd.GetRandomName()
		if c.walletKeys[name] == nil {
			break
		}
	}

	c.walletKeys[name] = privKey
	return cored.Wallet{Name: name, Address: privKey.Address()}, privKey
}

// Client creates new client for cored blockchain
func (c Cored) Client() *cored.Client {
	return cored.NewClient(c.executor, c.IP(), c.ports.RPC)
}

// HealthCheck checks if cored chain is empty
func (c Cored) HealthCheck(ctx context.Context) error {
	if c.appInfo.Status() != infra.AppStatusRunning {
		return retry.Retryable(fmt.Errorf("cored chain hasn't started yet"))
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	statusURL := url.URL{Scheme: "http", Host: net.JoinHostPort(c.IP().String(), strconv.Itoa(c.ports.RPC)), Path: "/status"}
	req := must.HTTPRequest(http.NewRequestWithContext(ctx, http.MethodGet, statusURL.String(), nil))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return retry.Retryable(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return retry.Retryable(err)
	}

	if resp.StatusCode != http.StatusOK {
		return retry.Retryable(fmt.Errorf("health check failed, status code: %d, response: %s", resp.StatusCode, body))
	}

	data := struct {
		Result struct {
			SyncInfo struct {
				LatestBlockHash string `json:"latest_block_hash"` // nolint: tagliatelle
			} `json:"sync_info"` // nolint: tagliatelle
		} `json:"result"`
	}{}

	if err := json.Unmarshal(body, &data); err != nil {
		return retry.Retryable(err)
	}

	if data.Result.SyncInfo.LatestBlockHash == "" {
		return retry.Retryable(errors.New("genesis block hasn't been mined yet"))
	}

	return nil
}

// Deployment returns deployment of cored
func (c Cored) Deployment() infra.Deployment {
	deployment := infra.Binary{
		BinPathFunc: func(targetOS string) string {
			return c.config.BinDir + "/" + targetOS + "/cored"
		},
		AppBase: infra.AppBase{
			Name: c.Name(),
			Info: c.appInfo,
			ArgsFunc: func(ip net.IP, homeDir string) []string {
				args := []string{
					"start",
					"--home", homeDir,
					"--rpc.laddr", "tcp://" + net.JoinHostPort(ip.String(), strconv.Itoa(c.ports.RPC)),
					"--p2p.laddr", "tcp://" + net.JoinHostPort(ip.String(), strconv.Itoa(c.ports.P2P)),
					"--grpc.address", net.JoinHostPort(ip.String(), strconv.Itoa(c.ports.GRPC)),
					"--grpc-web.address", net.JoinHostPort(ip.String(), strconv.Itoa(c.ports.GRPCWeb)),
					"--rpc.pprof_laddr", net.JoinHostPort(ip.String(), strconv.Itoa(c.ports.PProf)),
				}
				if c.rootNode != nil {
					args = append(args,
						"--p2p.persistent_peers", c.rootNode.PeerAddress(),
					)
				}

				return args
			},
			Ports: portsToMap(c.ports),
			PreFunc: func(ctx context.Context) error {
				c.mu.RLock()
				defer c.mu.RUnlock()

				cored.SaveIdentityFiles(c.executor.Home(), c.nodePrivateKey, c.validatorPrivateKey)

				cored.AddKeysToStore(c.executor.Home(), c.walletKeys)

				c.genesis.Save(c.executor.Home())
				return nil
			},
			PostFunc: func(ctx context.Context, deployment infra.DeploymentInfo) error {
				return c.saveClientWrapper(c.config.WrapperDir, deployment.IP)
			},
		},
	}
	if c.rootNode != nil {
		deployment.Requires = infra.Prerequisites{
			Timeout: 20 * time.Second,
			Dependencies: []infra.HealthCheckCapable{
				infra.IsRunning(*c.rootNode),
			},
		}
	}
	return deployment
}

func (c Cored) saveClientWrapper(wrapperDir string, ip net.IP) error {
	client := `#!/bin/sh
OPTS=""
if [ "$1" == "tx" ] || [ "$1" == "q" ]; then
	OPTS="$OPTS --chain-id ""` + c.genesis.ChainID() + `"" --node ""tcp://` + net.JoinHostPort(ip.String(), strconv.Itoa(c.ports.RPC)) + `"""
fi
if [ "$1" == "tx" ] || [ "$1" == "keys" ]; then
	OPTS="$OPTS --keyring-backend ""test"""
fi

exec ` + c.executor.Bin() + ` --home "` + c.executor.Home() + `" "$@" $OPTS
`
	return ioutil.WriteFile(wrapperDir+"/"+c.Name(), []byte(client), 0o700)
}
