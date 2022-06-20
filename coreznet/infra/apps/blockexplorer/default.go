package blockexplorer

import "github.com/CoreumFoundation/coreum/coreznet/infra/apps/postgres"

// DefaultPorts are the default ports applications building block explorer listen on
var DefaultPorts = Ports{
	Postgres: postgres.DefaultPort,
}
