//go:build integrationtests

package modules

import (
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
)

// TestGovProposalWithDepositAndWeightedVotes - is a complex governance test which tests:
// 1. proposal submission without enough deposit,
// 2. depositing missing amount to proposal created on the 1st step,
// 3. voting using weighted votes.
func TestGovProposalWithDepositAndWeightedVotes(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)
	gov := chain.Governance
	missingDepositAmount := chain.NewCoin(sdkmath.NewInt(10))

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := gov.ComputeProposerBalance(ctx)
	requireT.NoError(err)
	proposerBalance = proposerBalance.Sub(missingDepositAmount)
	requireT.NoError(chain.Faucet.FundAccounts(ctx, integrationtests.FundedAccount{Address: proposer, Amount: proposerBalance}))

	// Create proposer depositor.
	depositor := chain.GenAccount()
	err = chain.Faucet.FundAccountsWithOptions(ctx, depositor, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&govtypesv1beta1.MsgDeposit{}},
		Amount:   missingDepositAmount.Amount,
	})
	requireT.NoError(err)

	// Create proposal with deposit less than min deposit.
	proposalMsg, err := gov.NewMsgSubmitProposal(ctx, proposer, govtypesv1beta1.NewTextProposal("Test proposal with weighted votes", strings.Repeat("Description", 20)))
	requireT.NoError(err)
	proposalMsg.InitialDeposit = proposalMsg.InitialDeposit.Sub(sdk.Coins{missingDepositAmount}...)
	proposalID, err := gov.Propose(ctx, proposalMsg)
	requireT.NoError(err)

	logger.Get(ctx).Info("proposal created", zap.Uint64("proposal_id", proposalID))

	// Verify that proposal is waiting for deposit.
	requirePropStatusFunc := func(expectedStatus govtypesv1beta1.ProposalStatus) {
		proposal, err := gov.GetProposal(ctx, proposalID)
		requireT.NoError(err)
		requireT.Equal(expectedStatus, proposal.Status)
	}
	requirePropStatusFunc(govtypesv1beta1.StatusDepositPeriod)

	// Deposit missing amount to proposal.
	depositMsg := govtypesv1beta1.NewMsgDeposit(depositor, proposalID, sdk.Coins{missingDepositAmount})
	result, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(depositor),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(depositMsg)),
		depositMsg,
	)
	requireT.NoError(err)
	require.Equal(t, chain.GasLimitByMsgs(depositMsg), uint64(result.GasUsed))

	logger.Get(ctx).Info("deposited more funds to proposal", zap.String("txHash", result.TxHash), zap.Int64("gas_used", result.GasUsed))

	// Verify that proposal voting has started.
	requirePropStatusFunc(govtypesv1beta1.StatusVotingPeriod)

	// Store proposer and depositor balances before voting has finished.
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	accBalanceFunc := func(prop sdk.AccAddress) sdk.Coin {
		accBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
			Address: prop.String(),
			Denom:   chain.NetworkConfig.Denom,
		})
		requireT.NoError(err)
		return *accBalance.Balance
	}
	proposerBalanceBeforeVoting := accBalanceFunc(proposer)
	depositorBalanceBeforeVoting := accBalanceFunc(depositor)

	// Vote by all staker accounts:
	// NoWithVeto 70% & No,Yes,Abstain 10% each.
	err = gov.VoteAllWeighted(ctx,
		govtypesv1beta1.WeightedVoteOptions{
			govtypesv1beta1.WeightedVoteOption{
				Option: govtypesv1beta1.OptionNoWithVeto,
				Weight: sdk.MustNewDecFromStr("0.7"),
			},
			govtypesv1beta1.WeightedVoteOption{
				Option: govtypesv1beta1.OptionNo,
				Weight: sdk.MustNewDecFromStr("0.1"),
			},
			govtypesv1beta1.WeightedVoteOption{
				Option: govtypesv1beta1.OptionYes,
				Weight: sdk.MustNewDecFromStr("0.1"),
			},
			govtypesv1beta1.WeightedVoteOption{
				Option: govtypesv1beta1.OptionAbstain,
				Weight: sdk.MustNewDecFromStr("0.1"),
			},
		},
		proposalID,
	)
	requireT.NoError(err)

	// Wait for proposal result.
	finalStatus, err := chain.Governance.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1beta1.StatusRejected, finalStatus)

	// Assert that proposer & depositor deposits were not credited back.
	proposerBalanceAfterVoting := accBalanceFunc(proposer)
	depositorBalanceAfterVoting := accBalanceFunc(depositor)
	requireT.Equal(proposerBalanceBeforeVoting, proposerBalanceAfterVoting)
	requireT.Equal(depositorBalanceBeforeVoting, depositorBalanceAfterVoting)
}
