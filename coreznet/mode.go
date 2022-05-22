package coreznet

import (
	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
	"github.com/CoreumFoundation/coreum/coreznet/tests"
)

// DevMode is the environment for developer
func DevMode(af *apps.Factory) infra.Mode {
	genesis := cored.NewGenesis("coreddev")
	return infra.Mode{
		af.Cored("cored-node", apps.CoredPorts{
			RPC:     26657,
			P2P:     26656,
			GRPC:    9090,
			GRPCWeb: 9091,
			PProf:   6060,
		}, genesis, nil),
	}
}

// FullMode is the environment with all apps
func FullMode(af *apps.Factory) infra.Mode {
	genesis := cored.NewGenesis("coreddev")
	coreA := af.Cored("cored-a", apps.CoredPorts{
		RPC:     16657,
		P2P:     16656,
		GRPC:    19090,
		GRPCWeb: 19091,
		PProf:   16060,
	}, genesis, nil)
	coreB := af.Cored("cored-b", apps.CoredPorts{
		RPC:     26657,
		P2P:     26656,
		GRPC:    29090,
		GRPCWeb: 29091,
		PProf:   26060,
	}, genesis, &coreA)
	return infra.Mode{
		coreA,
		coreB,
	}
}

// TestMode returns environment used for testing
func TestMode(af *apps.Factory) infra.Mode {
	env, _ := tests.Tests(af)
	return env
}
