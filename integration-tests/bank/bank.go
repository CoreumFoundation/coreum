package bank

import "github.com/CoreumFoundation/coreum/integration-tests/testing"

// SingleChainTests returns single chain tests of the bank module
func SingleChainTests() []testing.SingleChainSignature {
	return []testing.SingleChainSignature{
		TestInitialBalance,
		TestCoreTransfer,
		TestTransferFailsIfNotEnoughGasIsProvided,
	}
}
