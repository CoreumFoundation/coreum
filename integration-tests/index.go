package tests

import (
	"github.com/CoreumFoundation/coreum/integration-tests/asset"
	"github.com/CoreumFoundation/coreum/integration-tests/auth"
	"github.com/CoreumFoundation/coreum/integration-tests/bank"
	"github.com/CoreumFoundation/coreum/integration-tests/feemodel"
	"github.com/CoreumFoundation/coreum/integration-tests/staking"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/integration-tests/wasm"
)

// Tests returns testing environment and tests
func Tests() testing.TestSet {
	testSet := testing.TestSet{
		SingleChain: []testing.SingleChainSignature{
			asset.TestIssueBasicFungibleToken,
			auth.TestGasLimits,
			auth.TestMultisig,
			auth.TestUnexpectedSequenceNumber,
			bank.TestCoreTransfer,
			bank.TestTransferDeterministicGas,
			bank.TestTransferFailsIfNotEnoughGasIsProvided,
			bank.TestTransferGasEstimation,
			bank.TestTransferDeterministicGasTwoBankSends,
			feemodel.TestQueryingMinGasPrice,
			feemodel.TestFeeModelProposalParamChange,
			staking.TestStakingProposalParamChange,
			staking.TestStaking,
			wasm.TestSimpleStateWasmContract,
			wasm.TestGasWasmBankSendAndBankSend,
			wasm.TestBankSendWasmContract,
		},
	}

	return testSet
}
