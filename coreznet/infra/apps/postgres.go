package apps

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/postgres"
	"github.com/CoreumFoundation/coreum/coreznet/pkg/retry"
)

// PostgresType is the type of postgres application
const PostgresType infra.AppType = "postgres"

// SchemaLoaderFunc is the function receiving sql client and loading schema there
type SchemaLoaderFunc func(ctx context.Context, db *pgx.Conn) error

// NewPostgres creates new postgres app
func NewPostgres(name string, appInfo *infra.AppInfo, port int, schemaLoaderFunc SchemaLoaderFunc) Postgres {
	return Postgres{
		name:             name,
		appInfo:          appInfo,
		port:             port,
		schemaLoaderFunc: schemaLoaderFunc,
	}
}

// Postgres represents postgres
type Postgres struct {
	name             string
	appInfo          *infra.AppInfo
	port             int
	schemaLoaderFunc SchemaLoaderFunc
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

// Deployment returns deployment of postgres
func (p Postgres) Deployment() infra.Deployment {
	return infra.Container{
		Image: "postgres:14.3-alpine",
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
			ArgsFunc: func() []string {
				args := []string{
					"-h", net.IPv4zero.String(),
					"-p", strconv.Itoa(p.port),
				}
				return args
			},
			Ports: map[string]int{
				"sql": p.port,
			},
			PostFunc: func(ctx context.Context, deployment infra.DeploymentInfo) error {
				if p.schemaLoaderFunc == nil || p.Info().Status != infra.AppStatusNotDeployed {
					return nil
				}

				log := logger.Get(ctx)

				db, err := p.dbConnection(ctx, deployment.FromHostIP)
				if err != nil {
					return err
				}
				defer db.Close(ctx)

				log.Info("Loading schema into the database")

				if err := p.schemaLoaderFunc(ctx, db); err != nil {
					return errors.Wrap(err, "loading schema failed")
				}

				log.Info("Database ready")
				return nil
			},
		},
	}
}

func (p Postgres) dbConnection(ctx context.Context, serverIP net.IP) (*pgx.Conn, error) {
	connStr := "postgres://" + postgres.User + "@" + infra.JoinProtoIPPort("", serverIP, p.port) + "/" + postgres.DB
	logger.Get(ctx).Info("Connecting to the database server", zap.String("connectionString", connStr))

	var db *pgx.Conn

	retryCtx, cancel := context.WithTimeout(ctx, 20*time.Second)
	defer cancel()
	err := retry.Do(retryCtx, time.Second, func() error {
		connCtx, cancel := context.WithTimeout(ctx, time.Second)
		defer cancel()

		var err error
		db, err = pgx.Connect(connCtx, connStr)
		return retry.Retryable(errors.WithStack(err))
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}
