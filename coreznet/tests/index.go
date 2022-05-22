package tests

import (
	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
	"github.com/CoreumFoundation/coreum/coreznet/infra/testing"
	"github.com/CoreumFoundation/coreum/coreznet/tests/transfers"
)

// Tests returns testing environment and tests
func Tests(appF *apps.Factory) (infra.Mode, []*testing.T) {
	genesis := cored.NewGenesis("coredtest")
	chain := appF.Cored("cored", apps.CoredPorts{
		RPC:     26657,
		P2P:     26656,
		GRPC:    9090,
		GRPCWeb: 9091,
		PProf:   6060,
	}, genesis, nil)
	return infra.Mode{
			chain,
		},
		[]*testing.T{
			testing.New(transfers.VerifyInitialBalance(chain)),
			testing.New(transfers.TransferCore(chain)),
		}
}
