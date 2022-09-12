package tests

import (
	"github.com/CoreumFoundation/coreum/integration-tests/auth"
	"github.com/CoreumFoundation/coreum/integration-tests/bank"
	"github.com/CoreumFoundation/coreum/integration-tests/feemodel"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/integration-tests/wasm"
)

// Tests returns testing environment and tests
func Tests() testing.TestSet {
	testSet := testing.TestSet{
		SingleChain: []testing.SingleChainSignature{
			auth.TestUnexpectedSequenceNumber,
			auth.TestTooLowGasPrice,
			auth.TestNoFee,
			auth.TestGasLimitHigherThanMaxBlockGas,
			auth.TestGasLimitEqualToMaxBlockGas,
			auth.TestMultisig,
			bank.TestInitialBalance,
			bank.TestCoreTransfer,
			bank.TestTransferFailsIfNotEnoughGasIsProvided,
			wasm.TestSimpleStateWasmContract,
			wasm.TestBankSendWasmContract,
			feemodel.TestQueryingMinGasPrice,
		},
	}

	// The idea is to run 200 transfer transactions to be sure that none of them uses more gas than we assumed.
	// To make each faster the same test is started 10 times, each broadcasting 20 transactions, to make use of parallelism
	// implemented inside testing framework. Test itself is written serially to not fight for resources with other tests.
	// In the future, once we have more tests running in parallel, we will replace 10 tests running 20 transactions each
	// with a single one running 200 of them.
	for i := 0; i < 10; i++ {
		testSet.SingleChain = append(testSet.SingleChain, bank.TestTransferMaximumGas(20))
	}

	return testSet
}
