//go:build integrationtests

package modules

import (
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v2/integration-tests"
	"github.com/CoreumFoundation/coreum/v2/pkg/client"
)

// TestGovProposalWithDepositAndWeightedVotes - is a complex governance test which tests:
// 1. proposal submission without enough deposit,
// 2. depositing missing amount to proposal created on the 1st step,
// 3. voting using weighted votes.
func TestGovProposalWithDepositAndWeightedVotes(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	gov := chain.Governance
	missingDepositAmount := chain.NewCoin(sdk.NewInt(10))

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := gov.ComputeProposerBalance(ctx)
	requireT.NoError(err)
	proposerBalance = proposerBalance.Sub(missingDepositAmount)
	chain.Faucet.FundAccounts(ctx, t,
		integrationtests.FundedAccount{
			Address: proposer,
			Amount:  proposerBalance,
		},
	)

	// Create proposer depositor.
	depositor := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, depositor, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&govtypes.MsgDeposit{}},
		Amount:   missingDepositAmount.Amount,
	})

	// Create proposal with deposit less than min deposit.
	proposalMsg, err := gov.NewMsgSubmitProposal(ctx, proposer, govtypes.NewTextProposal("Test proposal with weighted votes", strings.Repeat("Description", 20)))
	requireT.NoError(err)
	proposalMsg.InitialDeposit = proposalMsg.InitialDeposit.Sub(sdk.Coins{missingDepositAmount})
	proposalID, err := gov.Propose(ctx, t, proposalMsg)
	requireT.NoError(err)

	t.Logf("Proposal created, proposalID: %d", proposalID)

	// Verify that proposal is waiting for deposit.
	requirePropStatusFunc := func(expectedStatus govtypes.ProposalStatus) {
		proposal, err := gov.GetProposal(ctx, proposalID)
		requireT.NoError(err)
		requireT.Equal(expectedStatus, proposal.Status)
	}
	requirePropStatusFunc(govtypes.StatusDepositPeriod)

	// Deposit missing amount to proposal.
	depositMsg := govtypes.NewMsgDeposit(depositor, proposalID, sdk.Coins{missingDepositAmount})
	result, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(depositor),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(depositMsg)),
		depositMsg,
	)
	requireT.NoError(err)
	require.Equal(t, chain.GasLimitByMsgs(depositMsg), uint64(result.GasUsed))

	t.Logf("Deposited more funds to proposal, txHash:%s, gasUsed:%d", result.TxHash, result.GasUsed)

	// Verify that proposal voting has started.
	requirePropStatusFunc(govtypes.StatusVotingPeriod)

	// Store proposer and depositor balances before voting has finished.
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	accBalanceFunc := func(prop sdk.AccAddress) sdk.Coin {
		accBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
			Address: prop.String(),
			Denom:   chain.ChainSettings.Denom,
		})
		requireT.NoError(err)
		return *accBalance.Balance
	}
	proposerBalanceBeforeVoting := accBalanceFunc(proposer)
	depositorBalanceBeforeVoting := accBalanceFunc(depositor)

	// Vote by all staker accounts:
	// NoWithVeto 70% & No,Yes,Abstain 10% each.
	err = gov.VoteAllWeighted(ctx,
		govtypes.WeightedVoteOptions{
			govtypes.WeightedVoteOption{
				Option: govtypes.OptionNoWithVeto,
				Weight: sdk.MustNewDecFromStr("0.7"),
			},
			govtypes.WeightedVoteOption{
				Option: govtypes.OptionNo,
				Weight: sdk.MustNewDecFromStr("0.1"),
			},
			govtypes.WeightedVoteOption{
				Option: govtypes.OptionYes,
				Weight: sdk.MustNewDecFromStr("0.1"),
			},
			govtypes.WeightedVoteOption{
				Option: govtypes.OptionAbstain,
				Weight: sdk.MustNewDecFromStr("0.1"),
			},
		},
		proposalID,
	)
	requireT.NoError(err)

	// Wait for proposal result.
	finalStatus, err := chain.Governance.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusRejected, finalStatus)

	// Assert that proposer & depositor deposits were not credited back.
	proposerBalanceAfterVoting := accBalanceFunc(proposer)
	depositorBalanceAfterVoting := accBalanceFunc(depositor)
	requireT.Equal(proposerBalanceBeforeVoting, proposerBalanceAfterVoting)
	requireT.Equal(depositorBalanceBeforeVoting, depositorBalanceAfterVoting)
}
