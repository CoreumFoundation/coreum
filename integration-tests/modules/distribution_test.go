//go:build integrationtests

package modules

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	customparamstypes "github.com/CoreumFoundation/coreum/x/customparams/types"
)

// TestDistributionSpendCommunityPoolProposal checks that FundCommunityPool and SpendCommunityPoolProposal work correctly.
func TestDistributionSpendCommunityPoolProposal(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	distributionClient := distributiontypes.NewQueryClient(chain.ClientContext)

	// *** Check the MsgFundCommunityPool ***

	communityPoolFunder := chain.GenAccount()
	fundAmount := sdk.NewInt(1_000)
	msgFundCommunityPool := &distributiontypes.MsgFundCommunityPool{
		Amount:    sdk.NewCoins(chain.NewCoin(fundAmount)),
		Depositor: communityPoolFunder.String(),
	}

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, communityPoolFunder, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			msgFundCommunityPool,
		},
		Amount: fundAmount,
	}))

	// capture the pool amount now to check it later
	poolBeforeFunding := getCommunityPoolCoin(ctx, requireT, distributionClient)

	txResult, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(communityPoolFunder),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgFundCommunityPool)),
		msgFundCommunityPool,
	)
	requireT.NoError(err)
	// validate the deterministic gas
	requireT.Equal(chain.GasLimitByMsgs(msgFundCommunityPool), uint64(txResult.GasUsed))

	poolAfterFunding := getCommunityPoolCoin(ctx, requireT, distributionClient)

	// check that after funding we have more than before + funding amount
	requireT.True(poolAfterFunding.Sub(poolBeforeFunding).IsGTE(chain.NewCoin(fundAmount)))

	// *** Check the CommunityPoolSpendProposal ***

	// create new proposer
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)

	communityPoolRecipient := chain.GenAccount()

	err = chain.Faucet.FundAccounts(ctx, integrationtests.NewFundedAccount(proposer, proposerBalance))
	requireT.NoError(err)

	poolCoin := getCommunityPoolCoin(ctx, requireT, distributionClient)

	proposalMsg, err := chain.Governance.NewMsgSubmitProposal(ctx, proposer, distributiontypes.NewCommunityPoolSpendProposal(
		"Spend community pool",
		"Spend community pool",
		communityPoolRecipient,
		sdk.NewCoins(poolCoin),
	))
	requireT.NoError(err)
	proposalID, err := chain.Governance.Propose(ctx, proposalMsg)
	requireT.NoError(err)

	requireT.NoError(err)
	logger.Get(ctx).Info("Proposal has been submitted", zap.Uint64("proposalID", proposalID))

	// verify that voting period started
	proposal, err := chain.Governance.GetProposal(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusVotingPeriod, proposal.Status)

	// vote yes from all vote accounts
	err = chain.Governance.VoteAll(ctx, govtypes.OptionYes, proposal.ProposalId)
	requireT.NoError(err)

	logger.Get(ctx).Info("Voters have voted successfully, waiting for voting period to be finished", zap.Time("votingEndTime", proposal.VotingEndTime))

	// wait for proposal result.
	finalStatus, err := chain.Governance.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusPassed, finalStatus)

	// check that recipient has received the coins
	communityPoolRecipientBalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: communityPoolRecipient.String(),
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoins(poolCoin), communityPoolRecipientBalancesRes.Balances)
}

// TestDistributionWithdrawRewardWithDeterministicGas checks that withdraw reward works correctly and gas is deterministic.
func TestDistributionWithdrawRewardWithDeterministicGas(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	delegator := chain.GenAccount()
	delegatorRewardRecipient := chain.GenAccount()

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	customParamsClient := customparamstypes.NewQueryClient(chain.ClientContext)

	requireT := require.New(t)
	// the amount of the delegation should be big enough to get at least some reward for the few blocks
	amountToDelegate := sdk.NewInt(1_000_000)
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, delegator, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&stakingtypes.MsgDelegate{},
			&distributiontypes.MsgWithdrawDelegatorReward{},
			&distributiontypes.MsgSetWithdrawAddress{},
			&distributiontypes.MsgWithdrawDelegatorReward{},
		},
		Amount: amountToDelegate,
	}))

	delegatedCoin := chain.NewCoin(amountToDelegate)

	// *** Create new validator to use it in the test and capture all required balances. ***
	customStakingParams, err := customParamsClient.StakingParams(ctx, &customparamstypes.QueryStakingParamsRequest{})
	require.NoError(t, err)
	validatorStakingAmount := customStakingParams.Params.MinSelfDelegation.Mul(sdk.NewInt(2)) // we multiply not to conflict with the tests which increases the min amount
	validatorStakerAddress, validatorAddress, deactivateValidator, err := integrationtests.CreateValidator(ctx, chain, validatorStakingAmount, validatorStakingAmount)
	require.NoError(t, err)
	defer func() {
		err := deactivateValidator()
		require.NoError(t, err)
	}()

	// delegate coins
	delegateMsg := &stakingtypes.MsgDelegate{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validatorAddress.String(),
		Amount:           delegatedCoin,
	}

	clientCtx := chain.ClientContext

	logger.Get(ctx).Info("Delegating some coins to validator to withdraw later")
	_, err = tx.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(delegateMsg)),
		delegateMsg,
	)
	requireT.NoError(err)

	// capture the normal staker balance
	delegatorBalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: delegator.String(),
		Denom:   delegatedCoin.Denom,
	})
	requireT.NoError(err)
	delegatorBalanceBeforeWithdrawal := delegatorBalanceRes.Balance

	// capture validator staker balance
	validatorStakerBalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: validatorStakerAddress.String(),
		Denom:   delegatedCoin.Denom,
	})
	requireT.NoError(err)
	validatorStakerBalanceBeforeWithdrawal := validatorStakerBalanceRes.Balance

	// await next 5 blocks
	requireT.NoError(tx.AwaitNextBlocks(ctx, clientCtx, 5))

	// *** Withdraw and check the delegator reward. ***

	// normal delegator
	logger.Get(ctx).Info("Withdrawing the delegator reward")
	// withdraw the normal reward
	withdrawRewardMsg := &distributiontypes.MsgWithdrawDelegatorReward{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validatorAddress.String(),
	}
	txResult, err := tx.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(withdrawRewardMsg)),
		withdrawRewardMsg,
	)
	requireT.NoError(err)
	// validate the deterministic gas
	requireT.Equal(chain.GasLimitByMsgs(withdrawRewardMsg), uint64(txResult.GasUsed))

	delegatorBalanceRes, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: delegator.String(),
		Denom:   delegatedCoin.Denom,
	})
	requireT.NoError(err)
	delegatorBalanceAfterWithdrawal := delegatorBalanceRes.Balance

	feeSpentOnWithdrawReward := chain.ComputeNeededBalanceFromOptions(integrationtests.BalancesOptions{
		Messages: []sdk.Msg{withdrawRewardMsg},
	})

	delegatorReward := delegatorBalanceAfterWithdrawal.Amount.Sub(delegatorBalanceBeforeWithdrawal.Amount.Sub(feeSpentOnWithdrawReward))
	requireT.True(delegatorReward.IsPositive())
	logger.Get(ctx).Info("Withdrawing of the delegator reward is done", zap.String("amount", delegatorReward.String()))

	// *** Change the reward owner and withdraw the delegator reward. ***

	// Change the reward owner and withdraw the reward
	logger.Get(ctx).Info("Changing the reward recipient and windowing the reward")
	// change withdraw address
	setWithdrawAddressMsg := &distributiontypes.MsgSetWithdrawAddress{
		DelegatorAddress: delegator.String(),
		WithdrawAddress:  delegatorRewardRecipient.String(),
	}
	txResult, err = tx.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(setWithdrawAddressMsg)),
		setWithdrawAddressMsg,
	)
	requireT.NoError(err)
	// validate the deterministic gas
	requireT.Equal(chain.GasLimitByMsgs(setWithdrawAddressMsg), uint64(txResult.GasUsed))
	// withdraw the reward second time
	txResult, err = tx.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(delegator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(withdrawRewardMsg)),
		withdrawRewardMsg,
	)
	requireT.NoError(err)
	// validate the deterministic gas
	requireT.Equal(chain.GasLimitByMsgs(withdrawRewardMsg), uint64(txResult.GasUsed))
	delegatorRewardRecipientBalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: delegatorRewardRecipient.String(),
		Denom:   delegatedCoin.Denom,
	})
	requireT.NoError(err)
	requireT.True(delegatorRewardRecipientBalanceRes.Balance.IsPositive())

	// *** Withdraw the validator commission. ***

	// validator commission
	logger.Get(ctx).Info("Withdrawing the validator commission")
	// withdraw the normal reward
	withdrawCommissionMsg := &distributiontypes.MsgWithdrawValidatorCommission{
		ValidatorAddress: validatorAddress.String(),
	}

	err = chain.Faucet.FundAccountsWithOptions(ctx, validatorStakerAddress, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{withdrawCommissionMsg},
	})
	requireT.NoError(err)

	txResult, err = tx.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(validatorStakerAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(withdrawCommissionMsg)),
		withdrawCommissionMsg,
	)
	requireT.NoError(err)
	// validate the deterministic gas
	requireT.Equal(chain.GasLimitByMsgs(withdrawCommissionMsg), uint64(txResult.GasUsed))

	validatorStakerBalanceRes, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: validatorStakerAddress.String(),
		Denom:   delegatedCoin.Denom,
	})
	requireT.NoError(err)
	validatorStakerBalanceAfterWithdrawal := validatorStakerBalanceRes.Balance

	feeSpentOnWithdrawCommission := chain.ComputeNeededBalanceFromOptions(integrationtests.BalancesOptions{
		Messages: []sdk.Msg{withdrawCommissionMsg},
	})

	validatorStakerCommissionReward := validatorStakerBalanceAfterWithdrawal.Amount.Sub(validatorStakerBalanceBeforeWithdrawal.Amount.Sub(feeSpentOnWithdrawCommission))
	requireT.True(validatorStakerCommissionReward.IsPositive())
	logger.Get(ctx).Info("Withdrawing of the validator commission is done", zap.String("amount", validatorStakerCommissionReward.String()))
}

func getCommunityPoolCoin(ctx context.Context, requireT *require.Assertions, distributionClient distributiontypes.QueryClient) sdk.Coin {
	communityPoolRes, err := distributionClient.CommunityPool(ctx, &distributiontypes.QueryCommunityPoolRequest{})
	requireT.NoError(err)

	requireT.Equal(1, len(communityPoolRes.Pool))
	poolDecCoin := communityPoolRes.Pool[0]
	poolIntCoin := sdk.NewCoin(poolDecCoin.Denom, poolDecCoin.Amount.TruncateInt())
	requireT.True(poolIntCoin.IsPositive())

	return poolIntCoin
}
