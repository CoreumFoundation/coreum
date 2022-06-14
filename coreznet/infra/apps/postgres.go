package apps

import (
	"context"
	"net"
	"strconv"

	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/postgres"
	"github.com/CoreumFoundation/coreum/coreznet/pkg/retry"
)

// PostgresType is the type of postgres application
const PostgresType infra.AppType = "postgres"

// NewPostgres creates new postgres app
func NewPostgres(name string, appInfo *infra.AppInfo, port int) Postgres {
	return Postgres{
		name:    name,
		appInfo: appInfo,
		port:    port,
	}
}

// Postgres represents postgres
type Postgres struct {
	name    string
	appInfo *infra.AppInfo
	port    int
}

// Type returns type of application
func (p Postgres) Type() infra.AppType {
	return PostgresType
}

// Name returns name of app
func (p Postgres) Name() string {
	return p.name
}

// Port returns port used by postgres to accept client connections
func (p Postgres) Port() int {
	return p.port
}

// Info returns deployment info
func (p Postgres) Info() infra.DeploymentInfo {
	return p.appInfo.Info()
}

// HealthCheck checks if postgres is configured and running
func (p Postgres) HealthCheck(ctx context.Context) error {
	if p.appInfo.Info().Status != infra.AppStatusRunning {
		return retry.Retryable(errors.Errorf("postgres hasn't started yet"))
	}
	return nil
}

// Deployment returns deployment of postgres
func (p Postgres) Deployment() infra.Deployment {
	return infra.Container{
		Image: "postgres",
		Tag:   "latest",
		EnvVars: []infra.EnvVar{
			{
				Name:  "POSTGRES_USER",
				Value: postgres.User,
			},
			{
				Name:  "POSTGRES_DB",
				Value: postgres.DB,
			},

			// This allows to log in using any existing user (even superuser) without providing a password.
			// This is local, temporary development setup so security doesn't matter.
			{
				Name:  "POSTGRES_HOST_AUTH_METHOD",
				Value: "trust",
			},
		},
		AppBase: infra.AppBase{
			Name: p.Name(),
			Info: p.appInfo,
			ArgsFunc: func(bindIP net.IP, homeDir string, ipResolver infra.IPResolver) []string {
				args := []string{
					"-h", bindIP.String(),
					"-p", strconv.Itoa(p.port),
				}
				return args
			},
			Ports: map[string]int{
				"sql": p.port,
			},
			PostFunc: func(ctx context.Context, deployment infra.DeploymentInfo) error {
				// FIXME (wojciech): load initial sql here
				return nil
			},
		},
	}
}
