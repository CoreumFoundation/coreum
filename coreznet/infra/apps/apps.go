package apps

import (
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

// Cored creates new cored app
func (f *Factory) Cored(name string, ports cored.Ports, genesis *cored.Genesis, rootNode *Cored) Cored {
	return NewCored(name, f.config, genesis, cored.NewExecutor(genesis.ChainID(), f.config.BinDir+"/cored", f.config.AppDir+"/"+name),
		f.spec.DescribeApp(CoredType, name), ports, rootNode)
}
