package tests

import (
	"github.com/CoreumFoundation/coreum/integration-tests/auth"
	"github.com/CoreumFoundation/coreum/integration-tests/bank"
	"github.com/CoreumFoundation/coreum/integration-tests/feemodel"
	"github.com/CoreumFoundation/coreum/integration-tests/gov"
	"github.com/CoreumFoundation/coreum/integration-tests/staking"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/integration-tests/wasm"
)

// Tests returns testing environment and tests
func Tests() testing.TestSet {
	var testSet testing.TestSet

	// Add gov module tests
	testSet.SingleChain = append(testSet.SingleChain, gov.SingleChainTests()...)

	// Add auth module tests
	testSet.SingleChain = append(testSet.SingleChain, auth.SingleChainTests()...)

	// Add bank module tests
	testSet.SingleChain = append(testSet.SingleChain, bank.SingleChainTests()...)

	// Add wasm module tests
	testSet.SingleChain = append(testSet.SingleChain, wasm.SingleChainTests()...)

	// Add fee model tests
	testSet.SingleChain = append(testSet.SingleChain, feemodel.SingleChainTests()...)

	// Add staking module tests
	testSet.SingleChain = append(testSet.SingleChain, staking.SingleChainTests()...)

	return testSet
}
