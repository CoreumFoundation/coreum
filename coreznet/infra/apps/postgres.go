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

// SchemaLoader is the function receiving sql client and loading schema there
type SchemaLoader func(ctx context.Context, db *pgx.Conn) error

// NewPostgres creates new postgres app
func NewPostgres(name string, appInfo *infra.AppInfo, port int, schemaLoader SchemaLoader) Postgres {
	return Postgres{
		name:         name,
		appInfo:      appInfo,
		port:         port,
		schemaLoader: schemaLoader,
	}
}

// Postgres represents postgres
type Postgres struct {
	name         string
	appInfo      *infra.AppInfo
	port         int
	schemaLoader SchemaLoader
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
		Image: "postgres",
		Tag:   "alpine",
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
				if p.schemaLoader == nil || p.Info().Status != infra.AppStatusNotDeployed {
					return nil
				}

				log := logger.Get(ctx)

				var db *pgx.Conn

				connStr := "postgres://" + postgres.User + "@" + net.JoinHostPort(deployment.FromHostIP.String(), strconv.Itoa(p.port)) + "/" + postgres.DB

				log.Info("Connecting to the database server", zap.String("connectionString", connStr))

				retryCtx, cancel1 := context.WithTimeout(ctx, 20*time.Second)
				defer cancel1()
				err := retry.Do(retryCtx, 2*time.Second, func() error {
					connCtx, cancel := context.WithTimeout(retryCtx, time.Second)
					defer cancel()

					var err error
					db, err = pgx.Connect(connCtx, connStr)
					return retry.Retryable(errors.WithStack(err))
				})
				if err != nil {
					return err
				}
				defer db.Close(ctx)

				log.Info("Waiting for database to be created", zap.String("db", postgres.DB))

				retryCtx, cancel2 := context.WithTimeout(ctx, 20*time.Second)
				defer cancel2()
				err = retry.Do(retryCtx, time.Second, func() error {
					queryCtx, cancel := context.WithTimeout(retryCtx, time.Second)
					defer cancel()

					row := db.QueryRow(queryCtx, "select 1 as result from pg_database where datname=$1", postgres.DB)
					var dummy int
					if err := row.Scan(&dummy); err != nil {
						if errors.Is(err, pgx.ErrNoRows) {
							return retry.Retryable(errors.New("database hasn't been created yet"))
						}
						return retry.Retryable(errors.Wrap(err, "verifying database readiness failed"))
					}
					return nil
				})
				if err != nil {
					return err
				}

				log.Info("Loading schema into the database")

				if err := p.schemaLoader(ctx, db); err != nil {
					return errors.Wrap(err, "loading schema failed")
				}

				log.Info("Database ready")
				return nil
			},
		},
	}
}
