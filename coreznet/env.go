package coreznet

import (
	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/tests"
)

// DevSet is the environment for developer
func DevSet(af *apps.Factory) infra.Set {
	return infra.Set{
		af.Cored("cored-node"),
	}
}

// FullSet is the environment with all apps
func FullSet(af *apps.Factory) infra.Set {
	return infra.Set{
		af.Cored("cored-a"),
		af.Cored("cored-b"),
	}
}

// TestsSet returns environment used for testing
func TestsSet(af *apps.Factory) infra.Set {
	env, _ := tests.Tests(af)
	return env
}
