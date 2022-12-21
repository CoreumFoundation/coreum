package assetft

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// TestIssueBasic checks that fungible token is issued.
func TestIssueBasic(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	clientCtx := chain.ClientContext

	ftClient := assetfttypes.NewQueryClient(clientCtx)
	bankClient := banktypes.NewQueryClient(clientCtx)

	issuer := chain.GenAccount()
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
		Messages: []sdk.Msg{&assetfttypes.MsgIssue{}},
	}))

	// Issue the new fungible token
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "WBTC",
		Subunit:       "wsatoshi",
		Precision:     8,
		Description:   "Wrapped BTC",
		InitialAmount: sdk.NewInt(777),
		BurnRate:      sdk.NewDec(0),
	}

	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	require.NoError(t, err)
	assert.Equal(t, chain.GasLimitByMsgs(&assetfttypes.MsgIssue{}), uint64(res.GasUsed))
	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assetfttypes.EventTokenIssued](res.Events)

	require.NoError(t, err)
	require.Equal(t, assetfttypes.EventTokenIssued{
		Denom:         assetfttypes.BuildDenom(msg.Subunit, issuer),
		Issuer:        msg.Issuer,
		Symbol:        msg.Symbol,
		Precision:     msg.Precision,
		Subunit:       msg.Subunit,
		Description:   msg.Description,
		InitialAmount: msg.InitialAmount,
		Features:      []assetfttypes.TokenFeature{},
		BurnRate:      msg.BurnRate,
	}, *fungibleTokenIssuedEvts[0])

	denom := fungibleTokenIssuedEvts[0].Denom

	// query for the token to check what is stored
	gotToken, err := ftClient.Token(ctx, &assetfttypes.QueryTokenRequest{
		Denom: denom,
	})
	requireT.NoError(err)

	requireT.Equal(assetfttypes.FT{
		Denom:       denom,
		Issuer:      msg.Issuer,
		Symbol:      msg.Symbol,
		Subunit:     "wsatoshi",
		Precision:   8,
		Description: msg.Description,
		BurnRate:    msg.BurnRate,
	}, gotToken.Token)

	// query balance
	balanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: issuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, msg.InitialAmount).String(), balanceRes.Balance.String())
}
