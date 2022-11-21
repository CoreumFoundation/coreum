package tests

import (
	"github.com/CoreumFoundation/coreum/integration-tests/staking"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
)

// Tests returns testing environment and tests
func Tests() []testing.TestSet {
	return []testing.TestSet{
		// FIXME uncomment
		//{
		//	Name:     "Upgrade",
		//	Parallel: false,
		//	SingleChain: []testing.SingleChainSignature{
		//		upgrade.TestUpgrade,
		//	},
		//},
		{
			Name:     "Main",
			Parallel: true,
			SingleChain: []testing.SingleChainSignature{
				//asset.TestIssueBasicFungibleToken,
				//asset.TestFreezeFungibleToken,
				//auth.TestUnexpectedSequenceNumber,
				//auth.TestFeeLimits,
				//auth.TestMultisig,
				//bank.TestCoreTransfer,
				//bank.TestTransferFailsIfNotEnoughGasIsProvided,
				//bank.TestTransferDeterministicGas,
				//bank.TestTransferDeterministicGasTwoBankSends,
				//bank.TestTransferGasEstimation,
				//distribution.TestWithdrawRewardWithDeterministicGas,
				//distribution.TestSpendCommunityPoolProposal,
				//feemodel.TestQueryingMinGasPrice,
				//feemodel.TestFeeModelProposalParamChange,
				//staking.TestStakingProposalParamChange,
				//staking.TestValidatorCRUDAndStaking,
				staking.TestValidatorMinParamsSelfDelegation,
				//wasm.TestPinningAndUnpinningSmartContractUsingGovernance,
				//wasm.TestBankSendWASMContract,
				//wasm.TestGasWASMBankSendAndBankSend,
				//gov.TestProposalWithDepositAndWeightedVotes,
				//wasm.TestIssueFungibleTokenInWASMContract,
			},
		},
	}
}
