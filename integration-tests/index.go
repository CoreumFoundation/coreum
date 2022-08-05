package tests

import (
	"github.com/CoreumFoundation/coreum/integration-tests/bank"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
)

// Tests returns testing environment and tests
func Tests() testing.TestSet {
	return testing.TestSet{
		SingleChain: []testing.SingleChainSignature{
			bank.TestInitialBalance,
		},
	}
}
