package apps

import (
	"fmt"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
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
		port := 10000 + 10*i
		name := name + fmt.Sprintf("-%02d", i)
		node := NewCored(name, f.config, genesis, f.spec.DescribeApp(CoredType, name), cored.Ports{
			RPC:        port + 1,
			P2P:        port + 2,
			GRPC:       port + 3,
			GRPCWeb:    port + 4,
			PProf:      port + 5,
			Prometheus: port + 6,
		}, node0)
		if node0 == nil {
			node0 = &node
		}
		nodes = append(nodes, node)
	}
	return nodes
}
