package blockexplorer

import (
	"github.com/CoreumFoundation/coreum/crust/infra/apps/bdjuno"
	"github.com/CoreumFoundation/coreum/crust/infra/apps/hasura"
	"github.com/CoreumFoundation/coreum/crust/infra/apps/postgres"
)

// DefaultPorts are the default ports applications building block explorer listen on
var DefaultPorts = Ports{
	Postgres: postgres.DefaultPort,
	Hasura:   hasura.DefaultPort,
	BDJuno:   bdjuno.DefaultPort,
}
