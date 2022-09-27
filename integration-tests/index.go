package tests

import (
	"github.com/CoreumFoundation/coreum/integration-tests/auth"
	"github.com/CoreumFoundation/coreum/integration-tests/bank"
	"github.com/CoreumFoundation/coreum/integration-tests/feemodel"
	"github.com/CoreumFoundation/coreum/integration-tests/gov"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/integration-tests/wasm"
)

// Tests returns testing environment and tests
func Tests() testing.TestSet {
	testSet := testing.TestSet{
		SingleChain: []testing.SingleChainSignature{
			gov.TestProposalParamChange,
			auth.TestUnexpectedSequenceNumber,
			auth.TestTooLowGasPrice,
			auth.TestNoFee,
			auth.TestGasLimitHigherThanMaxBlockGas,
			auth.TestGasLimitEqualToMaxBlockGas,
			auth.TestMultisig,
			bank.TestInitialBalance,
			bank.TestCoreTransfer,
			bank.TestTransferFailsIfNotEnoughGasIsProvided,
			bank.TestTransferDeterministicGas,
			bank.TestTransferGasEstimation,
			wasm.TestSimpleStateWasmContract,
			wasm.TestBankSendWasmContract,
			feemodel.TestQueryingMinGasPrice,
		},
	}
	return testSet
}
