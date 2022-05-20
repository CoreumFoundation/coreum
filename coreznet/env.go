package coreznet

import (
	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/tests"
)

// DevMode is the environment for developer
func DevMode(af *apps.Factory) infra.Mode {
	return infra.Mode{
		af.Cored("cored-node"),
	}
}

// FullMode is the environment with all apps
func FullMode(af *apps.Factory) infra.Mode {
	return infra.Mode{
		af.Cored("cored-a"),
		af.Cored("cored-b"),
	}
}

// TestMode returns environment used for testing
func TestMode(af *apps.Factory) infra.Mode {
	env, _ := tests.Tests(af)
	return env
}
