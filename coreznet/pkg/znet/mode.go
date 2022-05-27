package znet

import (
	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
	"github.com/CoreumFoundation/coreum/coreznet/tests"
)

// DevMode is the environment for developer
func DevMode(af *apps.Factory) infra.Mode {
	return infra.Mode{
		af.Cored("cored-node", cored.DefaultPorts, cored.NewGenesis("coredev"), nil),
	}
}

// TestMode returns environment used for testing
func TestMode(af *apps.Factory) infra.Mode {
	env, _ := tests.Tests(af)
	return env
}
