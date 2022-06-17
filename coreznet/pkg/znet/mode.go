package znet

import (
	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/tests"
)

// DevMode is the environment for developer
func DevMode(appF *apps.Factory) infra.Mode {
	return appF.CoredNetwork("coredev", 1)
}

// TestMode returns environment used for testing
func TestMode(appF *apps.Factory) infra.Mode {
	env, _ := tests.Tests(appF)
	return env
}
