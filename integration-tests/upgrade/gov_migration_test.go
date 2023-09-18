//go:build integrationtests

package upgrade

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
)

var (
	fundAmount           = sdkmath.NewInt(1_000_000)
	missingDepositAmount = sdkmath.NewInt(10)
)

type govMigrationTest struct {
	onDepositProposalID    uint64
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
	proposerBalance, err := chain.LegacyGovernance.ComputeProposerBalance(ctx)
	requireT.NoError(err)
	chain.Faucet.FundAccounts(ctx, t, integrationtests.NewFundedAccount(proposer, proposerBalance))

	communityPoolRecipient := chain.GenAccount()
	proposalMsg, err := chain.LegacyGovernance.NewMsgSubmitProposalV1Beta1(
		ctx,
		proposer,
		&distributiontypes.CommunityPoolSpendProposal{ //nolint:staticcheck
			Title:       "Community pool spend created before upgrade",
			Description: "Community pool spend created before upgrade",
			Recipient:   communityPoolRecipient.String(),
			Amount:      sdk.NewCoins(chain.NewCoin(fundAmount)),
		},
	)

	// Subtract 10udevcore from initial deposit amount, so proposal stays on deposit status.
	proposalMsg.InitialDeposit = proposalMsg.InitialDeposit.Sub(chain.NewCoin(missingDepositAmount))
	requireT.NoError(err)
	proposalID, err := chain.LegacyGovernance.Propose(ctx, t, proposalMsg)
	requireT.NoError(err)

	proposal, err := chain.LegacyGovernance.GetProposal(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1beta1.StatusDepositPeriod, proposal.Status)

	gmt.onDepositProposalID = proposalID
	gmt.proposer = proposer
	gmt.communityPoolRecipient = communityPoolRecipient
}

func (gmt *govMigrationTest) After(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	proposal, err := chain.Governance.GetProposal(ctx, gmt.onDepositProposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusDepositPeriod, proposal.Status)
	// Proposer could be set as optional step during the upgrade, but we decided to not implement it
	// since proposal fails anyway.
	requireT.Equal("", proposal.Proposer)

	depositor := chain.GenAccount()
	requireT.NoError(err)

	chain.FundAccountWithOptions(ctx, t, depositor, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&govtypesv1.MsgDeposit{}},
		Amount:   missingDepositAmount,
	})

	depositMsg := govtypesv1.NewMsgDeposit(depositor, gmt.onDepositProposalID, sdk.NewCoins(chain.NewCoin(missingDepositAmount)))
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(depositor),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(depositMsg)),
		depositMsg,
	)
	requireT.NoError(err)

	proposal, err = chain.Governance.GetProposal(ctx, gmt.onDepositProposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusVotingPeriod, proposal.Status)

	requireT.NoError(chain.Governance.VoteAll(ctx, govtypesv1.OptionYes, gmt.onDepositProposalID))

	proposalStatus, err := chain.Governance.WaitForVotingToFinalize(ctx, gmt.onDepositProposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusFailed, proposalStatus)
	// Logs produced inside cored for such a proposal:
	// "proposal tallied module=x/gov proposal=1 results="passed, but msg 0 (/cosmos.gov.v1.MsgExecLegacyContent) failed on execution: distribution: no handler exists for proposal type"
}
