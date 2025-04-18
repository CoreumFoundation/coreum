//go:build integrationtests

package modules

import (
	"context"
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v6/integration-tests"
	"github.com/CoreumFoundation/coreum/v6/pkg/client"
	"github.com/CoreumFoundation/coreum/v6/testutil/integration"
)

// TestDistributionSpendCommunityPoolProposal checks that FundCommunityPool and SpendCommunityPoolProposal
// work correctly.
func TestDistributionSpendCommunityPoolProposal(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	distributionClient := distributiontypes.NewQueryClient(chain.ClientContext)

	// *** Check the MsgFundCommunityPool ***

	communityPoolFunder := chain.GenAccount()
	fundAmount := sdkmath.NewInt(1_000)
	msgFundCommunityPool := &distributiontypes.MsgFundCommunityPool{
		Amount:    sdk.NewCoins(chain.NewCoin(fundAmount)),
		Depositor: communityPoolFunder.String(),
	}

	chain.FundAccountWithOptions(ctx, t, communityPoolFunder, integration.BalancesOptions{
		Messages: []sdk.Msg{
			msgFundCommunityPool,
		},
		Amount: fundAmount,
	})

	// capture the pool amount now to check it later
	poolBeforeFunding := getCommunityPoolCoin(ctx, requireT, distributionClient)

	txResult, err := client.BroadcastTx(
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
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx, false)
	requireT.NoError(err)

	communityPoolRecipient := chain.GenAccount()

	chain.Faucet.FundAccounts(ctx, t, integration.NewFundedAccount(proposer, proposerBalance))
	poolCoin := getCommunityPoolCoin(ctx, requireT, distributionClient)

	msgPoolSpend := &distributiontypes.MsgCommunityPoolSpend{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Recipient: communityPoolRecipient.String(),
		Amount:    sdk.NewCoins(poolCoin),
	}
	proposalMsg, err := chain.Governance.NewMsgSubmitProposal(
		ctx,
		proposer,
		[]sdk.Msg{msgPoolSpend},
		"Spend community pool",
		"Spend community pool",
		"Spend community pool",
		false,
	)
	requireT.NoError(err)
	proposalID, err := chain.Governance.Propose(ctx, t, proposalMsg)
	requireT.NoError(err)

	requireT.NoError(err)
	t.Logf("Proposal has been submitted, proposalID:%d", proposalID)

	// verify that voting period started
	proposal, err := chain.Governance.GetProposal(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusVotingPeriod, proposal.Status)

	// vote yes from all vote accounts
	err = chain.Governance.VoteAll(ctx, govtypesv1.OptionYes, proposal.Id)
	requireT.NoError(err)

	t.Logf(
		"Voters have voted successfully, waiting for voting period to be finished, votingEndTime:%s",
		proposal.VotingEndTime,
	)

	// wait for proposal result.
	finalStatus, err := chain.Governance.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusPassed, finalStatus)

	// check that recipient has received the coins
	communityPoolRecipientBalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: communityPoolRecipient.String(),
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoins(poolCoin), communityPoolRecipientBalancesRes.Balances)
}

func getCommunityPoolCoin(
	ctx context.Context,
	requireT *require.Assertions,
	distributionClient distributiontypes.QueryClient,
) sdk.Coin {
	communityPoolRes, err := distributionClient.CommunityPool(ctx, &distributiontypes.QueryCommunityPoolRequest{})
	requireT.NoError(err)

	requireT.Len(communityPoolRes.Pool, 1)
	poolDecCoin := communityPoolRes.Pool[0]
	poolIntCoin := sdk.NewCoin(poolDecCoin.Denom, poolDecCoin.Amount.TruncateInt())
	requireT.True(poolIntCoin.IsPositive())

	return poolIntCoin
}
