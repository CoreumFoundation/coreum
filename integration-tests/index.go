package tests

import (
	"github.com/CoreumFoundation/coreum/integration-tests/asset"
	"github.com/CoreumFoundation/coreum/integration-tests/auth"
	"github.com/CoreumFoundation/coreum/integration-tests/bank"
	"github.com/CoreumFoundation/coreum/integration-tests/distribution"
	"github.com/CoreumFoundation/coreum/integration-tests/feemodel"
	"github.com/CoreumFoundation/coreum/integration-tests/gov"
	"github.com/CoreumFoundation/coreum/integration-tests/staking"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/integration-tests/upgrade"
	"github.com/CoreumFoundation/coreum/integration-tests/wasm"
)

// Tests returns testing environment and tests
func Tests() []testing.TestSet {
	return []testing.TestSet{
		{
			Name:     "Upgrade",
			Parallel: false,
			SingleChain: []testing.SingleChainSignature{
				upgrade.TestUpgrade,
			},
		},
		{
			Name:     "Main",
			Parallel: true,
			SingleChain: []testing.SingleChainSignature{
				asset.TestIssueBasicFungibleToken,
				asset.TestFreezeFungibleToken,
				asset.TestFreezeUnfreezableFungibleToken,
				asset.TestMintFungibleToken,
				asset.TestBurnFungibleToken,
				auth.TestUnexpectedSequenceNumber,
				auth.TestFeeLimits,
				auth.TestMultisig,
				bank.TestCoreTransfer,
				bank.TestTransferFailsIfNotEnoughGasIsProvided,
				bank.TestTransferDeterministicGas,
				bank.TestTransferDeterministicGasTwoBankSends,
				bank.TestTransferGasEstimation,
				distribution.TestWithdrawRewardWithDeterministicGas,
				distribution.TestSpendCommunityPoolProposal,
				feemodel.TestQueryingMinGasPrice,
				feemodel.TestFeeModelProposalParamChange,
				staking.TestStakingProposalParamChange,
				staking.TestValidatorCrudAndStaking,
				wasm.TestPinningAndUnpinningSmartContractUsingGovernance,
				wasm.TestBankSendWasmContract,
				wasm.TestGasWasmBankSendAndBankSend,
				gov.TestProposalWithDepositAndWeightedVotes,
				wasm.TestIssueFungibleTokenInWASMContract,
			},
		},
	}
}
