package tests

import (
	"github.com/CoreumFoundation/coreum/integration-tests/auth"
	"github.com/CoreumFoundation/coreum/integration-tests/bank"
	"github.com/CoreumFoundation/coreum/integration-tests/feemodel"
	"github.com/CoreumFoundation/coreum/integration-tests/staking"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/integration-tests/wasm"
)

// Tests returns testing environment and tests
func Tests() []testing.TestSet {
	return []testing.TestSet{
		// FIXME (wojtek): uncomment once crust is ready
		// {
		// 	Name: "Upgrade",
		// 	SingleChain: []testing.SingleChainSignature{
		// 		upgrade.TestUpgrade,
		// 	},
		// },
		{
			Parallel: true,
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
				bank.TestTransferDeterministicGas,
				bank.TestTransferDeterministicGasTwoBankSends,
				bank.TestTransferGasEstimation,
				feemodel.TestQueryingMinGasPrice,
				feemodel.TestFeeModelProposalParamChange,
				staking.TestStakingProposalParamChange,
				staking.TestStaking,
				wasm.TestSimpleStateWasmContract,
				wasm.TestBankSendWasmContract,
				wasm.TestGasWasmBankSendAndBankSend,
			},
		},
	}
}
