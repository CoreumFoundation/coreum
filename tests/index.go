package tests

import (
	"github.com/CoreumFoundation/crust/infra"
	"github.com/CoreumFoundation/crust/infra/apps/cored"
	"github.com/CoreumFoundation/crust/infra/testing"

	"github.com/CoreumFoundation/coreum/tests/auth"
	"github.com/CoreumFoundation/coreum/tests/bank"
)

// TODO (ysv): check if we can adapt our tests to run standard go testing framework

// Tests returns integration tests
func Tests(mode infra.Mode) []*testing.T {
	// FIXME (wojciech): Find a better name of getting `cored` instance from `mode`
	node := mode[0].(cored.Cored)

	return []*testing.T{
		testing.New(auth.TestUnexpectedSequenceNumber(node)),
		testing.New(bank.TestInitialBalance(node)),
		testing.New(bank.TestCoreTransfer(node)),
	}
}
