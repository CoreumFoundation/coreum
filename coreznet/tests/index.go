package tests

import (
	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/infra/testing"
	"github.com/CoreumFoundation/coreum/coreznet/tests/transfers"
)

// Tests returns testing environment and tests
func Tests(appF *apps.Factory) (infra.Mode, []*testing.T) {
	mode := appF.CoredNetwork("coretest", 3)
	node := mode[0].(apps.Cored)
	return mode,
		[]*testing.T{
			testing.New(transfers.VerifyInitialBalance(node)),
			testing.New(transfers.TransferCore(node)),
		}
}
