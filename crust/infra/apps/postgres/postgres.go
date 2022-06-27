package postgres

import (
	"context"
	"net"
	"strconv"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum/crust/infra"
	"github.com/CoreumFoundation/coreum/crust/pkg/retry"
)

const (
	// AppType is the type of postgres application
	AppType infra.AppType = "postgres"

	// DefaultPort is the default port postgres listens on for client connections
	DefaultPort = 5432

	// User contains the login of superuser
	User = "postgres"

	// DB is the name of database
	DB = "db"
)

// SchemaLoaderFunc is the function receiving sql client and loading schema there
type SchemaLoaderFunc func(ctx context.Context, db *pgx.Conn) error

// New creates new postgres app
func New(name string, appInfo *infra.AppInfo, port int, schemaLoaderFunc SchemaLoaderFunc) Postgres {
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
	return AppType
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

// HealthCheck checks if postgres is ready to accept connections
func (p Postgres) HealthCheck(ctx context.Context) error {
	if p.appInfo.Info().Status != infra.AppStatusRunning {
		return retry.Retryable(errors.Errorf("postgres hasn't started yet"))
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	connStr := "postgres://" + User + "@" + infra.JoinNetAddr("", p.appInfo.Info().HostFromHost, p.port) + "/" + DB
	db, err := pgx.Connect(ctx, connStr)
	if err != nil {
		return retry.Retryable(errors.WithStack(err))
	}

	if err := db.Ping(ctx); err != nil {
		return errors.WithStack(err)
	}

	time.Sleep(10 * time.Second)

	return retry.Retryable(errors.WithStack(db.Close(ctx)))
}

// Deployment returns deployment of postgres
func (p Postgres) Deployment() infra.Deployment {
	return infra.Container{
		Image: "postgres:14.3-alpine",
		EnvVars: []infra.EnvVar{
			{
				Name:  "POSTGRES_USER",
				Value: User,
			},
			{
				Name:  "POSTGRES_DB",
				Value: DB,
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
				return []string{
					"-h", net.IPv4zero.String(),
					"-p", strconv.Itoa(p.port),
				}
			},
			Ports: map[string]int{
				"sql": p.port,
			},
			ConfigureFunc: func(ctx context.Context, deployment infra.DeploymentInfo) error {
				if p.schemaLoaderFunc == nil {
					return nil
				}

				log := logger.Get(ctx)

				db, err := p.dbConnection(ctx, deployment.HostFromHost)
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

func (p Postgres) dbConnection(ctx context.Context, hostname string) (*pgx.Conn, error) {
	connStr := "postgres://" + User + "@" + infra.JoinNetAddr("", hostname, p.port) + "/" + DB
	logger.Get(ctx).Info("Connecting to the database server", zap.String("connectionString", connStr))

	var db *pgx.Conn

	retryCtx, cancel := context.WithTimeout(ctx, 40*time.Second)
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
