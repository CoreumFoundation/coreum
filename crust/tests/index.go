package tests

import (
	"github.com/CoreumFoundation/coreum/crust/infra"
	"github.com/CoreumFoundation/coreum/crust/infra/apps"
	"github.com/CoreumFoundation/coreum/crust/infra/apps/cored"
	"github.com/CoreumFoundation/coreum/crust/infra/testing"
	"github.com/CoreumFoundation/coreum/crust/tests/auth"
	"github.com/CoreumFoundation/coreum/crust/tests/bank"
)

// TODO (ysv): check if we can adapt our tests to run standard go testing framework

// Mode returns mode used by integration tests
func Mode(appF *apps.Factory) infra.Mode {
	return appF.CoredNetwork("coretest", 3)
}

// Tests returns testing environment and tests
func Tests(mode infra.Mode) []*testing.T {
	node := mode[0].(cored.Cored)
	return []*testing.T{
		testing.New(auth.TestUnexpectedSequenceNumber(node)),
		testing.New(bank.TestInitialBalance(node)),
		testing.New(bank.TestCoreTransfer(node)),
	}
}
