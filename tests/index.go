package tests

import (
	"github.com/CoreumFoundation/coreum/tests/bank"
	"github.com/CoreumFoundation/coreum/tests/testing"
)

// Tests returns testing environment and tests
func Tests() testing.TestSet {
	return testing.TestSet{
		SingleChain: []testing.SingleChainSignature{
			bank.TestInitialBalance,
		},
	}
}
