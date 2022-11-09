package gov

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TestProposalWithDepositAndWeightedVotes - is a complex governance test which tests:
// 1. proposal submission without enough deposit
// 2. depositing missing amount to proposal created on the 1st step
// 3. voting using weighted votes
func TestProposalWithDepositAndWeightedVotes(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)

	missingDepositAmount := chain.NewCoin(sdk.NewInt(10))

	gov := chain.Governance

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := gov.ComputeProposerBalance(ctx)
	requireT.NoError(err)
	proposerBalance = proposerBalance.Sub(missingDepositAmount)
	requireT.NoError(chain.Faucet.FundAccounts(ctx, testing.FundedAccount{Address: proposer, Amount: proposerBalance}))

	// Create proposer depositor.
	depositor := chain.GenAccount()
	err = chain.Faucet.FundAccountsWithOptions(ctx, depositor, testing.BalancesOptions{
		Messages: []sdk.Msg{&govtypes.MsgDeposit{}},
		Amount:   missingDepositAmount.Amount,
	})
	requireT.NoError(err)

	govParams, err := gov.QueryGovParams(ctx)
	requireT.NoError(err)

	// Create proposal with deposit less than min deposit.
	initialDeposit := govParams.DepositParams.MinDeposit[0].Sub(missingDepositAmount)
	msg, err := govtypes.NewMsgSubmitProposal(
		govtypes.NewTextProposal("Test proposal with weighted votes", "-"),
		sdk.Coins{initialDeposit},
		proposer,
	)
	proposalID, err := gov.ProposeV2(ctx, msg)
	requireT.NoError(err)

	logger.Get(ctx).Info("proposal created", zap.Int("proposal_id", proposalID))

	// Verify that proposal is waiting for deposit.
	requireProposalStatusF := func(expectedStatus govtypes.ProposalStatus) {
		proposal, err := gov.GetProposal(ctx, uint64(proposalID))
		requireT.NoError(err)
		requireT.Equal(expectedStatus, proposal.Status)
	}
	requireProposalStatusF(govtypes.StatusDepositPeriod)

	// Deposit missing amount to proposal.
	msg2 := govtypes.NewMsgDeposit(depositor, uint64(proposalID), sdk.Coins{missingDepositAmount})
	result, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(depositor),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg2)),
		msg2,
	)
	requireT.NoError(err)
	logger.Get(ctx).Info("deposited more funds to proposal", zap.String("txHash", result.TxHash), zap.Int64("gas_used", result.GasUsed))

	// Verify that proposal voting has started.
	requireProposalStatusF(govtypes.StatusVotingPeriod)

	// Store proposer and depositor balances before voting has finished.
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	accBalanceF := func(address sdk.AccAddress) sdk.Coin {
		accBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
			Address: proposer.String(),
			Denom:   chain.NetworkConfig.Denom,
		})
		requireT.NoError(err)
		return *accBalance.Balance
	}
	proposerBalanceBeforeVoting := accBalanceF(proposer)
	depositorBalanceBeforeVoting := accBalanceF(depositor)

	// Vote by all staker accounts 70% - NoWithVeto 30% - Yes.
	err = gov.VoteAllWeighted(ctx,
		govtypes.WeightedVoteOptions{
			govtypes.WeightedVoteOption{
				Option: govtypes.OptionNoWithVeto,
				Weight: sdk.MustNewDecFromStr("0.7"),
			},
			govtypes.WeightedVoteOption{
				Option: govtypes.OptionYes,
				Weight: sdk.MustNewDecFromStr("0.3"),
			},
		},
		uint64(proposalID),
	)
	requireT.NoError(err)

	// Wait for proposal result.
	finalStatus, err := chain.Governance.WaitForVotingToFinalize(ctx, uint64(proposalID))
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusRejected, finalStatus)

	// Assert that proposer & depositor deposits were not credited back.
	proposerBalanceAfterVoting := accBalanceF(proposer)
	depositorBalanceAfterVoting := accBalanceF(depositor)
	requireT.Equal(proposerBalanceBeforeVoting, proposerBalanceAfterVoting)
	requireT.Equal(depositorBalanceBeforeVoting, depositorBalanceAfterVoting)
}
