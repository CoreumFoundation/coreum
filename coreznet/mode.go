package coreznet

import (
	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
	"github.com/CoreumFoundation/coreum/coreznet/tests"
)

// DevMode is the environment for developer
func DevMode(af *apps.Factory) infra.Mode {
	genesis := cored.NewGenesis("coredev")
	return infra.Mode{
		af.Cored("cored-node", cored.Ports{
			RPC:     26657,
			P2P:     26656,
			GRPC:    9090,
			GRPCWeb: 9091,
			PProf:   6060,
		}, genesis, nil),
	}
}

// TestMode returns environment used for testing
func TestMode(af *apps.Factory) infra.Mode {
	env, _ := tests.Tests(af)
	return env
}
