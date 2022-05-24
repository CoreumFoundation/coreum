package tests

import (
	"github.com/CoreumFoundation/coreum/coreznet/infra"
	"github.com/CoreumFoundation/coreum/coreznet/infra/apps"
	"github.com/CoreumFoundation/coreum/coreznet/infra/testing"
	"github.com/CoreumFoundation/coreum/coreznet/tests/transfers"
)

// Tests returns testing environment and tests
func Tests(appF *apps.Factory) (infra.Mode, []*testing.T) {
	chain := appF.Cored("cored")
	return infra.Mode{
			chain,
		},
		[]*testing.T{
			testing.New(transfers.VerifyInitialBalance(chain)),
			testing.New(transfers.TransferCore(chain)),
		}
}
