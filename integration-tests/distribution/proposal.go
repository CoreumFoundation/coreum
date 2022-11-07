package distribution

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
)

// TestSpendCommunityPoolProposal checks that SpendCommunityPoolProposal works correctly.
func TestSpendCommunityPoolProposal(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	distributionClient := distributiontypes.NewQueryClient(chain.ClientContext)

	// create new proposer
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)

	communityPoolRecipient := chain.GenAccount()

	err = chain.Faucet.FundAccounts(ctx, testing.NewFundedAccount(proposer, proposerBalance))
	requireT.NoError(err)

	// get the community pool balance
	communityPoolRes, err := distributionClient.CommunityPool(ctx, &distributiontypes.QueryCommunityPoolRequest{})
	requireT.NoError(err)

	requireT.Equal(1, len(communityPoolRes.Pool))
	poolDecCoin := communityPoolRes.Pool[0]
	poolIntCoin := sdk.NewCoin(poolDecCoin.Denom, poolDecCoin.Amount.TruncateInt())
	requireT.True(poolIntCoin.IsPositive())
	poolIntCoins := sdk.NewCoins(poolIntCoin)

	// create proposition to spend the community pool
	proposalID, err := chain.Governance.Propose(
		ctx,
		proposer,
		distributiontypes.NewCommunityPoolSpendProposal(
			"Spend community pool",
			"Spend community pool",
			communityPoolRecipient,
			poolIntCoins,
		),
	)
	requireT.NoError(err)
	logger.Get(ctx).Info("Proposal has been submitted", zap.Int("proposalID", proposalID))

	// verify that voting period started
	proposal, err := chain.Governance.GetProposal(ctx, uint64(proposalID))
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusVotingPeriod, proposal.Status)

	// vote yes from all vote accounts
	err = chain.Governance.VoteAll(ctx, govtypes.OptionYes, proposal.ProposalId)
	requireT.NoError(err)

	logger.Get(ctx).Info("Voters have voted successfully, waiting for voting period to be finished", zap.Time("votingEndTime", proposal.VotingEndTime))

	// wait for proposal result.
	requireT.NoError(chain.Governance.WaitForVotingToPass(ctx, uint64(proposalID)))

	// check that recipient has received the coins
	communityPoolRecipientBalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: communityPoolRecipient.String(),
	})
	requireT.NoError(err)
	requireT.Equal(poolIntCoins, communityPoolRecipientBalancesRes.Balances)
}
