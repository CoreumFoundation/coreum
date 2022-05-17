package apps

import (
	"context"
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
func NewCored(config infra.Config, executor *cored.Executor, appInfo *infra.AppInfo, ports CoredPorts) *Cored {
	return &Cored{
		config:   config,
		executor:   executor,
		genesis:    cored.NewGenesis(executor),
		appInfo:    appInfo,
		ports:      ports,
		mu:         &sync.RWMutex{},
	}
}

// Cored represents cored
type Cored struct {
	config   infra.Config
	executor   *cored.Executor
	genesis    *cored.Genesis
	appInfo    *infra.AppInfo
	ports      CoredPorts

	mu *sync.RWMutex
	ip net.IP
}

// ChainID returns chain ID
func (c Cored) ChainID() string {
	return c.executor.Name()
}

// Name returns name of app
func (c Cored) Name() string {
	return c.executor.Name()
}

// IP returns IP chain listens on
func (c Cored) IP() net.IP {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.ip
}

// Genesis returns configurator of genesis block
func (c Cored) Genesis() *cored.Genesis {
	return c.genesis
}

// Client creates new client for cored blockchain
func (c Cored) Client() *cored.Client {
	return cored.NewClient(c.executor, c.IP(), c.ports.RPC)
}

// HealthCheck checks if cored chain is empty
func (c Cored) HealthCheck(ctx context.Context) error {
	if c.IP() == nil {
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
			Name: c.executor.Name(),
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
				return c.executor.PrepareNode(ctx, c.genesis)
			},
			PostFunc: func(ctx context.Context, deployment infra.DeploymentInfo) error {
				c.mu.Lock()
				c.ip = deployment.IP
				c.mu.Unlock()

				return c.saveClientWrapper(c.config.WrapperDir)
			},
		},
	}
}

func (c Cored) saveClientWrapper(wrapperDir string) error {
	client := `#!/bin/sh
OPTS=""
if [ "$1" == "tx" ] || [ "$1" == "q" ]; then
	OPTS="$OPTS --chain-id ""` + c.executor.Name() + `"" --node ""tcp://` + c.IP().String() + ":" + fmt.Sprintf("%d", c.ports.RPC) + `"""
fi
if [ "$1" == "tx" ] || [ "$1" == "keys" ]; then
	OPTS="$OPTS --keyring-backend ""test"""
fi

exec ` + c.executor.Bin() + ` --home "` + c.executor.Home() + `" "$@" $OPTS
`
	return ioutil.WriteFile(wrapperDir+"/"+c.executor.Name(), []byte(client), 0o700)
}
