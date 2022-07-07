package tests

import (
	"github.com/CoreumFoundation/crust/infra"
	"github.com/CoreumFoundation/crust/infra/apps"
	"github.com/CoreumFoundation/crust/infra/apps/cored"
	"github.com/CoreumFoundation/crust/infra/testing"
	"github.com/CoreumFoundation/crust/tests/auth"
	"github.com/CoreumFoundation/crust/tests/bank"
)

// TODO (ysv): check if we can adapt our tests to run standard go testing framework

// Tests returns testing environment and tests
func Tests(appF *apps.Factory) (infra.Mode, []*testing.T) {
	mode := appF.CoredNetwork("coretest", 3, 0)
	node := mode[0].(cored.Cored)
	return mode,
		[]*testing.T{
			testing.New(auth.TestUnexpectedSequenceNumber(node)),
			testing.New(bank.TestInitialBalance(node)),
			testing.New(bank.TestCoreTransfer(node)),
		}
}
