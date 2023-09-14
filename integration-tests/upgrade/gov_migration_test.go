//go:build integrationtests

package upgrade

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
)

type govMigrationTest struct {
}

func (gmt *govMigrationTest) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	// Fund community pool.
	communityPoolFunder := chain.GenAccount()
	fundAmount := sdkmath.NewInt(1_000_000)
	msgFundCommunityPool := &distributiontypes.MsgFundCommunityPool{
		Amount:    sdk.NewCoins(chain.NewCoin(fundAmount)),
		Depositor: communityPoolFunder.String(),
	}

	chain.FundAccountWithOptions(ctx, t, communityPoolFunder, integrationtests.BalancesOptions{
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

	// Propose community pool spend but keep on deposit stage.
	proposer := chain.GenAccount()
	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)

	communityPoolRecipient := chain.GenAccount()

	chain.Faucet.FundAccounts(ctx, t, integrationtests.NewFundedAccount(proposer, proposerBalance))

	msgPoolSpend := &distributiontypes.MsgCommunityPoolSpend{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Recipient: communityPoolRecipient.String(),
		Amount:    sdk.NewCoins(chain.NewCoin(fundAmount)),
	}
	proposalMsg, err := chain.LegacyGovernance.NewMsgSubmitProposalV1Beta1(
		ctx,
		proposer,
		distributiontypes.MsgCommunityPoolSpend{
			Authority: "",
			Recipient: "",
			Amount:    nil,
		},
	)

	proposalMsg.InitialDeposit = proposalMsg.InitialDeposit[0]
	requireT.NoError(err)
	proposalID, err := chain.Governance.Propose(ctx, t, proposalMsg)
	requireT.NoError(err)

	//proposer := chain.GenAccount()
	//
	//paramChangeProposal := paramproposal.NewParameterChangeProposal("title", "description",
	//	[]paramproposal.ParamChange{
	//		paramproposal.NewParamChange(stakingtypes.ModuleName, string(stakingtypes.KeyMaxValidators), strconv.Itoa(int(targetMaxValidators))),
	//	},
	//	//[]paramproposal.ParamChange{
	//	//{
	//	//	Subspace: "foo",
	//	//	Key:      "bar",
	//	//	Value:    "baz",
	//	//},
	//})
	//
	//proposalMsg, err := chain.LegacyGovernance.NewMsgSubmitProposalV1Beta1(
	//	ctx,
	//	proposer,
	//	paramChangeProposal,
	//)
	//
	//chain.LegacyGovernance.Propose()
}

func (gmt *govMigrationTest) After(t *testing.T) {
}
