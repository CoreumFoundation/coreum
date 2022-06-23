package znet

import (
	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/tests"
)

// DevMode is the environment for developer
func DevMode(appF *apps.Factory) infra.Mode {
	var mode infra.Mode
	mode = append(mode, appF.CoredNetwork("coredev", 1)...)
	mode = append(mode, appF.BlockExplorer("explorer")...)
	return mode
}

// TestMode returns environment used for testing
func TestMode(appF *apps.Factory) infra.Mode {
	mode, _ := tests.Tests(appF)
	return mode
}
