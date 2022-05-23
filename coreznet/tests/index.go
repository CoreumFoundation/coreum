package tests

import (
	"fmt"

	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps/cored"
	"github.com/CoreumFoundation/coreum/coreznet/infra/testing"
	"github.com/CoreumFoundation/coreum/coreznet/tests/transfers"
)

// Tests returns testing environment and tests
func Tests(appF *apps.Factory) (infra.Mode, []*testing.T) {
	mode, node := coredNodes("coretest", 3, appF)
	return mode,
		[]*testing.T{
			testing.New(transfers.VerifyInitialBalance(node)),
			testing.New(transfers.TransferCore(node)),
		}
}

func coredNodes(chainID string, numOfNodes int, appF *apps.Factory) (infra.Mode, apps.Cored) {
	genesis := cored.NewGenesis(chainID)
	node0 := appF.Cored("cored-00", apps.CoredPorts{
		RPC:     10001,
		P2P:     10002,
		GRPC:    10003,
		GRPCWeb: 10004,
		PProf:   10005,
	}, genesis, nil)
	node0.AddWallet("1000000000000000core,1000000000000000stake")
	nodes := infra.Mode{
		node0,
	}
	for i := 1; i < numOfNodes; i++ {
		port := 10000 + 10*(i+1)
		node := appF.Cored(fmt.Sprintf("cored-%02d", i), apps.CoredPorts{
			RPC:     port + 1,
			P2P:     port + 2,
			GRPC:    port + 3,
			GRPCWeb: port + 4,
			PProf:   port + 5,
		}, genesis, &node0)
		node.AddWallet("1000000000000000core,1000000000000000stake")
		nodes = append(nodes, node)
	}
	return nodes, node0
}
