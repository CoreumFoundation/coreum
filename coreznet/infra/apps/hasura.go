package apps

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/postgres"
)

// HasuraType is the type of hasura application
const HasuraType infra.AppType = "hasura"

// NewHasura creates new hasura app
func NewHasura(name string, appInfo *infra.AppInfo, port int, postgres Postgres) Hasura {
	return Hasura{
		name:     name,
		appInfo:  appInfo,
		port:     port,
		postgres: postgres,
	}
}

// Hasura represents hasura
type Hasura struct {
	name     string
	appInfo  *infra.AppInfo
	port     int
	postgres Postgres
}

// Type returns type of application
func (h Hasura) Type() infra.AppType {
	return HasuraType
}

// Name returns name of app
func (h Hasura) Name() string {
	return h.name
}

// Port returns port used by hasura to accept client connections
func (h Hasura) Port() int {
	return h.port
}

// Info returns deployment info
func (h Hasura) Info() infra.DeploymentInfo {
	return h.appInfo.Info()
}

// Deployment returns deployment of hasura
func (h Hasura) Deployment() infra.Deployment {
	return infra.Container{
		Image: "hasura/graphql-engine",
		Tag:   "latest",
		AppBase: infra.AppBase{
			Name: h.Name(),
			Info: h.appInfo,
			ArgsFunc: func(bindIP net.IP, homeDir string, ipResolver infra.IPResolver) []string {
				return []string{
					"graphql-engine",
					"--host", ipResolver.IPOf(h.postgres).String(),
					"--port", strconv.Itoa(h.postgres.Port()),
					"--user", postgres.User,
					"--dbname", postgres.DB,
					"serve",
					"--server-host", bindIP.String(),
					"--server-port", strconv.Itoa(h.port),
					"--enable-console",
					"--dev-mode",
					"--enabled-log-types", "startup,http-log,webhook-log,websocket-log,query-log",
				}
			},
			Ports: map[string]int{
				"server": h.port,
			},
			Requires: infra.Prerequisites{
				Timeout: 20 * time.Second,
				Dependencies: []infra.HealthCheckCapable{
					infra.IsRunning(h.postgres),
				},
			},
			PostFunc: func(ctx context.Context, deployment infra.DeploymentInfo) error {
				// FIXME (wojciech): Load metadata
				return nil
			},
		},
	}
}
