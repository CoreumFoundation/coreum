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
	node0 := appF.Cored("cored-00", cored.DefaultPorts, genesis, nil)
	nodes := infra.Mode{
		node0,
	}
	for i := 1; i < numOfNodes; i++ {
		port := 10000 + 10*i
		node := appF.Cored(fmt.Sprintf("cored-%02d", i), cored.Ports{
			RPC:        port + 1,
			P2P:        port + 2,
			GRPC:       port + 3,
			GRPCWeb:    port + 4,
			PProf:      port + 5,
			Prometheus: port + 6,
		}, genesis, &node0)
		nodes = append(nodes, node)
	}
	return nodes, node0
}
