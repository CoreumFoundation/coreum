package bdjuno

import (
	"bytes"
	"io/ioutil"
	"text/template"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/postgres"
	"github.com/CoreumFoundation/coreum/coreznet/infra/targets"
)

const (
	// AppType is the type of bdjuno application
	AppType infra.AppType = "bdjuno"

	// DefaultPort is the default port bdjuno listens on for client connections
	DefaultPort = 3030
)

// New creates new bdjuno app
func New(name string, config infra.Config, appInfo *infra.AppInfo, port int, configTemplate string, cored cored.Cored, postgres postgres.Postgres) BDJuno {
	return BDJuno{
		name:           name,
		homeDir:        config.AppDir + "/" + name,
		appInfo:        appInfo,
		port:           port,
		configTemplate: configTemplate,
		cored:          cored,
		postgres:       postgres,
	}
}

// BDJuno represents bdjuno
type BDJuno struct {
	name           string
	homeDir        string
	appInfo        *infra.AppInfo
	port           int
	configTemplate string
	cored          cored.Cored
	postgres       postgres.Postgres
}

// Type returns type of application
func (j BDJuno) Type() infra.AppType {
	return AppType
}

// Name returns name of app
func (j BDJuno) Name() string {
	return j.name
}

// Port returns port used by hasura to accept client connections
func (j BDJuno) Port() int {
	return j.port
}

// Info returns deployment info
func (j BDJuno) Info() infra.DeploymentInfo {
	return j.appInfo.Info()
}

// Deployment returns deployment of hasura
func (j BDJuno) Deployment() infra.Deployment {
	return infra.Container{
		Image: "gcr.io/coreum-devnet-1/bdjuno:0.44.0",
		AppBase: infra.AppBase{
			Name: j.Name(),
			Info: j.appInfo,
			ArgsFunc: func() []string {
				return []string{
					"bdjuno", "start",
					"--home", targets.AppHomeDir,
				}
			},
			Ports: map[string]int{
				"actions": j.port,
			},
			Requires: infra.Prerequisites{
				Timeout: 20 * time.Second,
				Dependencies: []infra.HealthCheckCapable{
					j.cored,
					infra.IsRunning(j.postgres),
				},
			},
			PrepareFunc: func() error {
				return ioutil.WriteFile(j.homeDir+"/config.yaml", j.prepareConfig(), 0o644)
			},
		},
	}
}

func (j BDJuno) prepareConfig() []byte {
	configBuf := &bytes.Buffer{}
	must.OK(template.Must(template.New("config").Parse(j.configTemplate)).Execute(configBuf, struct {
		Port  int
		Cored struct {
			Host          string
			PortRPC       int
			PortGRPC      int
			AddressPrefix string
		}
		Postgres struct {
			Host string
			Port int
			User string
			DB   string
		}
	}{
		Port: j.port,
		Cored: struct {
			Host          string
			PortRPC       int
			PortGRPC      int
			AddressPrefix string
		}{
			Host:          j.cored.Info().FromContainerIP.String(),
			PortRPC:       j.cored.Ports().RPC,
			PortGRPC:      j.cored.Ports().GRPC,
			AddressPrefix: sdk.GetConfig().GetBech32AccountAddrPrefix(),
		},
		Postgres: struct {
			Host string
			Port int
			User string
			DB   string
		}{
			Host: j.postgres.Info().FromContainerIP.String(),
			Port: j.postgres.Port(),
			User: postgres.User,
			DB:   postgres.DB,
		},
	}))
	return configBuf.Bytes()
}
