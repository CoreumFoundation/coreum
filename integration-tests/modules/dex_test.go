//go:build integrationtests

package modules

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	"github.com/CoreumFoundation/coreum/v4/pkg/client"
	"github.com/CoreumFoundation/coreum/v4/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

// TestAssetIssueAndQueryTokens checks that tokens query works as expected.
func TestCreateLimitOrder(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	clientCtx := chain.ClientContext

	issueFee := chain.QueryAssetFTParams(ctx, t).IssueFee.Amount

	issuer1 := chain.GenAccount()
	issuer2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer1, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&types.MsgCreateLimitOrder{},
		},
		Amount: issueFee,
	})
	chain.FundAccountWithOptions(ctx, t, issuer2, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&types.MsgCreateLimitOrder{},
		},
		Amount: issueFee,
	})

	msgIssue1 := &assetfttypes.MsgIssue{
		Issuer:             issuer1.String(),
		Symbol:             "AAA",
		Subunit:            "uaaa",
		Precision:          6,
		InitialAmount:      sdkmath.NewInt(1000),
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.NewDec(0),
	}
	msgIssue2 := &assetfttypes.MsgIssue{
		Issuer:             issuer2.String(),
		Symbol:             "BBB",
		Subunit:            "ubbb",
		Precision:          6,
		InitialAmount:      sdkmath.NewInt(1000),
		BurnRate:           sdk.NewDec(0),
		SendCommissionRate: sdk.NewDec(0),
	}

	_, err := client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(issuer1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgIssue1)),
		msgIssue1,
	)
	requireT.NoError(err)

	_, err = client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(issuer2),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgIssue2)),
		msgIssue2,
	)
	requireT.NoError(err)

	denom1 := assetfttypes.BuildDenom(msgIssue1.Subunit, issuer1)
	denom2 := assetfttypes.BuildDenom(msgIssue2.Subunit, issuer2)

	msgOrder1 := &types.MsgCreateLimitOrder{
		Owner:         issuer1.String(),
		OfferedAmount: sdk.NewInt64Coin(denom1, 10),
		SellPrice:     sdk.NewDecCoinFromDec(denom2, sdk.MustNewDecFromStr("0.5")),
	}
	msgOrder2 := &types.MsgCreateLimitOrder{
		Owner:         issuer2.String(),
		OfferedAmount: sdk.NewInt64Coin(denom2, 5),
		SellPrice:     sdk.NewInt64DecCoin(denom1, 2),
	}

	_, err = client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(issuer1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgOrder1)),
		msgOrder1,
	)
	requireT.NoError(err)

	_, err = client.BroadcastTx(
		ctx,
		clientCtx.WithFromAddress(issuer2),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgOrder2)),
		msgOrder2,
	)
	requireT.NoError(err)
}
