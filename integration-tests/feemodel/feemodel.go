package feemodel

import "github.com/CoreumFoundation/coreum/integration-tests/testing"

// SingleChainTests returns single chain tests of the fee model
func SingleChainTests() []testing.SingleChainSignature {
	return []testing.SingleChainSignature{
		TestQueryingMinGasPrice,
	}
}
