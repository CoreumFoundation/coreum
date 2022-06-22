package apps

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
	"github.com/CoreumFoundation/coreum/coreznet/infra/targets"
	"github.com/CoreumFoundation/coreum/coreznet/pkg/retry"
	"github.com/CoreumFoundation/coreum/coreznet/pkg/rnd"
)

// CoredType is the type of cored application
const CoredType infra.AppType = "cored"

// NewCored creates new cored app
func NewCored(name string, config infra.Config, genesis *cored.Genesis, appInfo *infra.AppInfo, ports cored.Ports, rootNode *Cored) Cored {
	nodePublicKey, nodePrivateKey, err := ed25519.GenerateKey(rand.Reader)
	must.OK(err)
	validatorPublicKey, validatorPrivateKey, err := ed25519.GenerateKey(rand.Reader)
	must.OK(err)

	stakerPubKey, stakerPrivKey := cored.GenerateSecp256k1Key()

	genesis.AddWallet(stakerPubKey, "100000000000000000000000core")
	genesis.AddValidator(validatorPublicKey, stakerPrivKey, "100000000core")

	return Cored{
		name:                name,
		homeDir:             config.AppDir + "/" + name,
		config:              config,
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
	homeDir             string
	config              infra.Config
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

// NodeID returns node ID
func (c Cored) NodeID() string {
	return c.nodeID
}

// Ports returns ports used by the application
func (c Cored) Ports() cored.Ports {
	return c.ports
}

// ChainID returns ID of the chain
func (c Cored) ChainID() string {
	return c.genesis.ChainID()
}

// Info returns deployment info
func (c Cored) Info() infra.DeploymentInfo {
	return c.appInfo.Info()
}

// AddWallet adds wallet to genesis block and local keystore
func (c Cored) AddWallet(balances string) cored.Wallet {
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
	return cored.Wallet{Name: name, Key: privKey}
}

// Client creates new client for cored blockchain
func (c Cored) Client() cored.Client {
	return cored.NewClient(c.genesis.ChainID(), infra.JoinProtoIPPort("", c.Info().FromHostIP, c.Ports().RPC))
}

// HealthCheck checks if cored chain is ready to accept transactions
func (c Cored) HealthCheck(ctx context.Context) error {
	if c.appInfo.Info().Status != infra.AppStatusRunning {
		return retry.Retryable(errors.Errorf("cored chain hasn't started yet"))
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	statusURL := url.URL{Scheme: "http", Host: infra.JoinProtoIPPort("", c.Info().FromHostIP, c.ports.RPC), Path: "/status"}
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
		return retry.Retryable(errors.Errorf("health check failed, status code: %d, response: %s", resp.StatusCode, body))
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
		BinPath: c.config.BinDir + "/linux/cored",
		AppBase: infra.AppBase{
			Name: c.Name(),
			Info: c.appInfo,
			ArgsFunc: func() []string {
				args := []string{
					"start",
					"--home", targets.AppHomeDir,
					"--rpc.laddr", infra.JoinProtoIPPort("tcp", net.IPv4zero, c.ports.RPC),
					"--p2p.laddr", infra.JoinProtoIPPort("tcp", net.IPv4zero, c.ports.P2P),
					"--grpc.address", infra.JoinProtoIPPort("", net.IPv4zero, c.ports.GRPC),
					"--grpc-web.address", infra.JoinProtoIPPort("", net.IPv4zero, c.ports.GRPCWeb),
					"--rpc.pprof_laddr", infra.JoinProtoIPPort("", net.IPv4zero, c.ports.PProf),
				}
				if c.rootNode != nil {
					args = append(args,
						"--p2p.persistent_peers", c.rootNode.NodeID()+"@"+infra.JoinProtoIPPort("", c.rootNode.Info().FromContainerIP, c.rootNode.Ports().P2P),
					)
				}

				return args
			},
			Ports: portsToMap(c.ports),
			PreFunc: func() error {
				c.mu.RLock()
				defer c.mu.RUnlock()

				cored.NodeConfig{
					Name:           c.name,
					PrometheusPort: c.ports.Prometheus,
					NodeKey:        c.nodePrivateKey,
					ValidatorKey:   c.validatorPrivateKey,
				}.Save(c.homeDir)

				cored.AddKeysToStore(c.homeDir, c.walletKeys)

				c.genesis.Save(c.homeDir)
				return nil
			},
			PostFunc: func(ctx context.Context, deployment infra.DeploymentInfo) error {
				return c.saveClientWrapper(c.config.WrapperDir, deployment.FromHostIP)
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
	client := `#!/bin/bash
OPTS=""
if [ "$1" == "tx" ] || [ "$1" == "q" ] || [ "$1" == "query" ]; then
	OPTS="$OPTS --chain-id ""` + c.genesis.ChainID() + `"" --node ""` + infra.JoinProtoIPPort("tcp", ip, c.ports.RPC) + `"""
fi
if [ "$1" == "tx" ] || [ "$1" == "keys" ]; then
	OPTS="$OPTS --keyring-backend ""test"""
fi

exec "` + c.config.BinDir + `/cored" --home "` + c.homeDir + `" "$@" $OPTS
`
	return ioutil.WriteFile(wrapperDir+"/"+c.Name(), []byte(client), 0o700)
}
