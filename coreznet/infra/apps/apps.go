package apps

import (
	"fmt"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/blockexplorer"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/blockexplorer/postgres"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
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
	var node0 *Cored
	for i := 0; i < numOfNodes; i++ {
		name := name + fmt.Sprintf("-%02d", i)
		portDelta := i * 100
		node := NewCored(name, f.config, genesis, f.spec.DescribeApp(CoredType, name), cored.Ports{
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
func (f *Factory) BlockExplorer(name string) infra.Mode {
	namePostgres := name + "-postgres"
	return infra.Mode{
		NewPostgres(namePostgres, f.spec.DescribeApp(PostgresType, namePostgres), blockexplorer.DefaultPorts.Postgres, postgres.LoadSchema),
		// FIXME (wojciech): more apps coming soon
	}
}
