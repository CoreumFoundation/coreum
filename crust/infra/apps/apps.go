package apps

import (
	"fmt"

	"github.com/CoreumFoundation/coreum/crust/infra"
	"github.com/CoreumFoundation/coreum/crust/infra/apps/bdjuno"
	"github.com/CoreumFoundation/coreum/crust/infra/apps/blockexplorer"
	"github.com/CoreumFoundation/coreum/crust/infra/apps/cored"
	"github.com/CoreumFoundation/coreum/crust/infra/apps/hasura"
	"github.com/CoreumFoundation/coreum/crust/infra/apps/postgres"
)

// NewFactory creates new app factory
func NewFactory(config infra.Config, spec *infra.Spec) *Factory {
	return &Factory{
		config: config,
		spec:   spec,
	}
}

// Factory produces apps from config
type Factory struct {
	config infra.Config
	spec   *infra.Spec
}

// CoredNetwork creates new network of cored nodes
func (f *Factory) CoredNetwork(name string, numOfNodes int) infra.Mode {
	genesis := cored.NewGenesis(name)
	nodes := make(infra.Mode, 0, numOfNodes)
	var node0 *cored.Cored
	for i := 0; i < numOfNodes; i++ {
		name := name + fmt.Sprintf("-%02d", i)
		portDelta := i * 100
		node := cored.New(name, f.config, genesis, f.spec.DescribeApp(cored.AppType, name), cored.Ports{
			RPC:        cored.DefaultPorts.RPC + portDelta,
			P2P:        cored.DefaultPorts.P2P + portDelta,
			GRPC:       cored.DefaultPorts.GRPC + portDelta,
			GRPCWeb:    cored.DefaultPorts.GRPCWeb + portDelta,
			PProf:      cored.DefaultPorts.PProf + portDelta,
			Prometheus: cored.DefaultPorts.Prometheus + portDelta,
		}, node0)
		if node0 == nil {
			node0 = &node
		}
		nodes = append(nodes, node)
	}
	return nodes
}

// BlockExplorer returns set of applications required to run block explorer
func (f *Factory) BlockExplorer(name string, coredApp cored.Cored) infra.Mode {
	namePostgres := name + "-postgres"
	nameHasura := name + "-hasura"
	nameBDJuno := name + "-bdjuno"

	postgresApp := postgres.New(namePostgres, f.spec.DescribeApp(postgres.AppType, namePostgres), blockexplorer.DefaultPorts.Postgres, blockexplorer.LoadPostgresSchema)
	hasuraApp := hasura.New(nameHasura, f.spec.DescribeApp(hasura.AppType, nameHasura), blockexplorer.DefaultPorts.Hasura, blockexplorer.HasuraMetadataTemplate, postgresApp)
	bdjunoApp := bdjuno.New(nameBDJuno, f.config, f.spec.DescribeApp(bdjuno.AppType, nameBDJuno), blockexplorer.DefaultPorts.BDJuno, blockexplorer.BDJunoConfigTemplate, coredApp, postgresApp)
	return infra.Mode{
		postgresApp,
		hasuraApp,
		bdjunoApp,
		// FIXME (wojciech): more apps coming soon
	}
}
