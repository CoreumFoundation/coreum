package asset

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

// TestIssueBasicFungibleToken checks that fungible token is issued.
func TestIssueBasicFungibleToken(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	clientCtx := chain.ClientContext

	assetClient := assettypes.NewQueryClient(clientCtx)
	bankClient := banktypes.NewQueryClient(clientCtx)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
		Messages: []sdk.Msg{&assettypes.MsgIssueFungibleToken{}},
	}))

	// Issue the new fungible token
	msg := &assettypes.MsgIssueFungibleToken{
		Issuer:      issuer.String(),
		Symbol:      "BTC",
		Description: "BTC Description",
		// the custom receiver
		Recipient:     recipient.String(),
		InitialAmount: sdk.NewInt(777),
	}

	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	require.NoError(t, err)
	assert.Equal(t, chain.GasLimitByMsgs(&assettypes.MsgIssueFungibleToken{}), uint64(res.GasUsed))
	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assettypes.EventFungibleTokenIssued](res.Events)

	require.NoError(t, err)
	require.Equal(t, assettypes.EventFungibleTokenIssued{
		Denom:         assettypes.BuildFungibleTokenDenom(msg.Symbol, issuer),
		Issuer:        msg.Issuer,
		Symbol:        msg.Symbol,
		Description:   msg.Description,
		Recipient:     msg.Recipient,
		InitialAmount: msg.InitialAmount,
		Features:      []assettypes.FungibleTokenFeature{},
	}, *fungibleTokenIssuedEvts[0])

	denom := fungibleTokenIssuedEvts[0].Denom

	// query for the token to check what is stored
	gotToken, err := assetClient.FungibleToken(ctx, &assettypes.QueryFungibleTokenRequest{
		Denom: denom,
	})
	requireT.NoError(err)

	requireT.Equal(assettypes.FungibleToken{
		Denom:       denom,
		Issuer:      msg.Issuer,
		Symbol:      msg.Symbol,
		Description: msg.Description,
	}, gotToken.FungibleToken)

	// query balance
	// check the recipient balance
	recipientBalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, msg.InitialAmount).String(), recipientBalanceRes.Balance.String())

	// check the issuer balance
	issuerBalanceRes, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: issuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.True(issuerBalanceRes.Balance.Amount.IsZero())
}
