//go:build integrationtests

package upgrade

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
)

var fundAmount = sdkmath.NewInt(1_000_000)

type govMigrationTest struct {
	onDepositProposalId    uint64
	proposer               sdk.AccAddress
	communityPoolRecipient sdk.AccAddress
}

func (gmt *govMigrationTest) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	// Fund community pool.
	communityPoolFunder := chain.GenAccount()
	msgFundCommunityPool := &distributiontypes.MsgFundCommunityPool{
		Amount:    sdk.NewCoins(chain.NewCoin(fundAmount)),
		Depositor: communityPoolFunder.String(),
	}

	chain.FundAccountWithOptions(ctx, t, communityPoolFunder,
		integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				msgFundCommunityPool,
			},
			Amount: fundAmount,
		})
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(communityPoolFunder),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgFundCommunityPool)),
		msgFundCommunityPool,
	)
	requireT.NoError(err)

	// Propose community pool spend but keep proposal in deposit status.
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)
	chain.Faucet.FundAccounts(ctx, t, integrationtests.NewFundedAccount(proposer, proposerBalance))

	communityPoolRecipient := chain.GenAccount()
	proposalMsg, err := chain.LegacyGovernance.NewMsgSubmitProposalV1Beta1(
		ctx,
		proposer,
		&distributiontypes.CommunityPoolSpendProposal{
			Title:       "Community pool spend created before upgrade",
			Description: "Community pool spend created before upgrade",
			Recipient:   communityPoolRecipient.String(),
			Amount:      sdk.NewCoins(chain.NewCoin(fundAmount)),
		},
	)

	// Subtract 10udevcore from initial deposit amount, so proposal stays on deposit status.
	proposalMsg.InitialDeposit = proposalMsg.InitialDeposit.Sub(chain.NewCoin(sdkmath.NewInt(10)))
	requireT.NoError(err)
	proposalID, err := chain.LegacyGovernance.Propose(ctx, t, proposalMsg)
	requireT.NoError(err)

	proposal, err := chain.LegacyGovernance.GetProposal(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1beta1.StatusDepositPeriod, proposal.Status)

	gmt.onDepositProposalId = proposalID
	gmt.proposer = proposer
	gmt.communityPoolRecipient = communityPoolRecipient
}

func (gmt *govMigrationTest) After(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	proposal, err := chain.Governance.GetProposal(ctx, gmt.onDepositProposalId)
	requireT.NoError(err)
	requireT.Equal(govtypesv1beta1.StatusDepositPeriod, proposal.Status)
	requireT.Equal(gmt.proposer.String(), proposal.Proposer)

	depositor := chain.GenAccount()
	requireT.NoError(err)

	missingDepositAmount := chain.NewCoin(sdkmath.NewInt(10))
	chain.FundAccountWithOptions(ctx, t, depositor, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&govtypesv1.MsgDeposit{}},
		Amount:   missingDepositAmount.Amount,
	})

	depositMsg := govtypesv1.NewMsgDeposit(depositor, gmt.onDepositProposalId, sdk.NewCoins(missingDepositAmount))
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(depositor),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(depositMsg)),
		depositMsg,
	)
	requireT.NoError(err)

	proposal, err = chain.Governance.GetProposal(ctx, gmt.onDepositProposalId)
	requireT.NoError(err)
	requireT.Equal(govtypesv1beta1.StatusVotingPeriod, proposal.Status)

	requireT.NoError(chain.Governance.VoteAll(ctx, govtypesv1.OptionYes, gmt.onDepositProposalId))

	proposalStatus, err := chain.Governance.WaitForVotingToFinalize(ctx, gmt.onDepositProposalId)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusPassed, proposalStatus)

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: gmt.communityPoolRecipient.String(),
		Denom:   chain.ChainSettings.Denom,
	})
	requireT.NoError(err)
	requireT.True(balance.Balance.Amount.Equal(fundAmount))
}
