package auth

import "github.com/CoreumFoundation/coreum/integration-tests/testing"

// SingleChainTests returns single chain tests of the auth module
func SingleChainTests() []testing.SingleChainSignature {
	return []testing.SingleChainSignature{
		TestUnexpectedSequenceNumber,
		TestTooLowGasPrice,
		TestNoFee,
		TestGasLimitHigherThanMaxBlockGas,
		TestGasLimitEqualToMaxBlockGas,
		TestMultisig,
	}
}
