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
	"sync"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
	"github.com/CoreumFoundation/coreum/coreznet/pkg/retry"
	"github.com/CoreumFoundation/coreum/coreznet/pkg/rnd"
)

// CoredPorts defines ports used by cored application
type CoredPorts struct {
	RPC     int `json:"rpc"`
	P2P     int `json:"p2p"`
	GRPC    int `json:"grpc"`
	GRPCWeb int `json:"grpcWeb"`
	PProf   int `json:"pprof"`
}

// NewCored creates new cored app
func NewCored(name string, config infra.Config, genesis *cored.Genesis, executor cored.Executor, appInfo *infra.AppInfo, ports CoredPorts) Cored {
	nodePublicKey, nodePrivateKey, err := ed25519.GenerateKey(rand.Reader)
	must.OK(err)
	validatorPublicKey, validatorPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	must.OK(err)

	c := Cored{
		name:                name,
		config:   config,
		executor:            executor,
		nodeID:              cored.NodeID(nodePublicKey),
		nodePrivateKey:      nodePrivateKey,
		validatorPrivateKey: validatorPrivateKey,
		genesis:             genesis,
		appInfo:             appInfo,
		ports:               ports,
		walletKeys:          map[string]cored.Secp256k1PrivateKey{},
		mu:                  &sync.RWMutex{},
	}

	_, stakerPrivateKey := c.AddWallet("500000000000000000000000core,990000000000000000000000000stake")
	genesis.AddValidator(validatorPublicKey, stakerPrivateKey)

	return c
}

// Cored represents cored
type Cored struct {
	name                string
	config   infra.Config
	executor            cored.Executor
	nodeID              string
	nodePrivateKey      ed25519.PrivateKey
	validatorPrivateKey ed25519.PrivateKey
	genesis             *cored.Genesis
	appInfo             *infra.AppInfo
	ports               CoredPorts

	mu         *sync.RWMutex
	walletKeys map[string]cored.Secp256k1PrivateKey
}

// Name returns name of app
func (c Cored) Name() string {
	return c.name
}

// ID returns ID of the node
func (c Cored) ID() string {
	return c.nodeID
}

// IP returns IP chain listens on
func (c Cored) IP() net.IP {
	return c.appInfo.IP()
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

	req := must.HTTPRequest(http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s:%d/status", c.IP(), c.ports.RPC), nil))
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
	return infra.Binary{
		BinPathFunc: func(targetOS string) string {
			return c.config.BinDir + "/" + targetOS + "/cored"
		},
		AppBase: infra.AppBase{
			Name: c.Name(),
			Info: c.appInfo,
			Args: []string{
				"start",
				"--home", "{{ .HomeDir }}",
				"--rpc.laddr", fmt.Sprintf("tcp://{{ .IP }}:%d", c.ports.RPC),
				"--p2p.laddr", fmt.Sprintf("tcp://{{ .IP }}:%d", c.ports.P2P),
				"--grpc.address", fmt.Sprintf("{{ .IP }}:%d", c.ports.GRPC),
				"--grpc-web.address", fmt.Sprintf("{{ .IP }}:%d", c.ports.GRPCWeb),
				"--rpc.pprof_laddr", fmt.Sprintf("{{ .IP }}:%d", c.ports.PProf),
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
}

func (c Cored) saveClientWrapper(wrapperDir string, ip net.IP) error {
	client := `#!/bin/sh
OPTS=""
if [ "$1" == "tx" ] || [ "$1" == "q" ]; then
	OPTS="$OPTS --chain-id ""` + c.genesis.ChainID() + `"" --node ""tcp://` + ip.String() + ":" + fmt.Sprintf("%d", c.ports.RPC) + `"""
fi
if [ "$1" == "tx" ] || [ "$1" == "keys" ]; then
	OPTS="$OPTS --keyring-backend ""test"""
fi

exec ` + c.executor.Bin() + ` --home "` + c.executor.Home() + `" "$@" $OPTS
`
	return ioutil.WriteFile(wrapperDir+"/"+c.Name(), []byte(client), 0o700)
}
