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
		genesis.AddWallet(cored.AlicePrivKey.PubKey(), "1000000000000000core")
		genesis.AddWallet(cored.BobPrivKey.PubKey(), "1000000000000000core")
		genesis.AddWallet(cored.CharliePrivKey.PubKey(), "1000000000000000core")
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
			"alice":   cored.AlicePrivKey,
			"bob":     cored.BobPrivKey,
			"charlie": cored.CharliePrivKey,
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

// Type returns type of application
func (c Cored) Type() infra.AppType {
	return CoredType
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
if [ "$1" == "tx" ] || [ "$1" == "q" ] || [ "$1" == "query" ]; then
	OPTS="$OPTS --chain-id ""` + c.genesis.ChainID() + `"" --node ""tcp://` + net.JoinHostPort(ip.String(), strconv.Itoa(c.ports.RPC)) + `"""
fi
if [ "$1" == "tx" ] || [ "$1" == "keys" ]; then
	OPTS="$OPTS --keyring-backend ""test"""
fi

exec ` + c.executor.Bin() + ` --home "` + c.executor.Home() + `" "$@" $OPTS
`
	return ioutil.WriteFile(wrapperDir+"/"+c.Name(), []byte(client), 0o700)
}
