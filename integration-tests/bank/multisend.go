package bank

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

// TestMultiSend tests MultiSend message
func TestMultiSend(ctx context.Context, t testing.T, chain testing.Chain) {
	sender := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	amount := sdk.NewInt(1000)

	issueMsgs := []sdk.Msg{
		&assettypes.MsgIssueFungibleToken{
			Issuer:        sender.String(),
			Symbol:        "TOK1",
			Subunit:       "tok1",
			Description:   "TOK1 Description",
			Recipient:     sender.String(),
			InitialAmount: amount,
		},
		&assettypes.MsgIssueFungibleToken{
			Issuer:        sender.String(),
			Symbol:        "TOK2",
			Subunit:       "tok2",
			Description:   "TOK2 Description",
			Recipient:     sender.String(),
			InitialAmount: amount,
		},
	}

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, testing.BalancesOptions{
		Messages: append([]sdk.Msg{&banktypes.MsgMultiSend{Outputs: []banktypes.Output{
			{Coins: make(sdk.Coins, 2)},
			{Coins: make(sdk.Coins, 2)},
		}}}, issueMsgs...),
	}))

	// Issue fungible tokens
	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsgs...)),
		issueMsgs...,
	)
	require.NoError(t, err)

	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assettypes.EventFungibleTokenIssued](res.Events)
	require.NoError(t, err)
	require.Equal(t, len(issueMsgs), len(fungibleTokenIssuedEvts))

	denom1 := fungibleTokenIssuedEvts[0].Denom
	denom2 := fungibleTokenIssuedEvts[1].Denom

	msg := &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{
			{
				Address: sender.String(),
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin(denom1, 1000),
					sdk.NewInt64Coin(denom2, 1000),
				),
			},
		},
		Outputs: []banktypes.Output{
			{
				Address: recipient1.String(),
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin(denom1, 600),
					sdk.NewInt64Coin(denom2, 400),
				),
			},
			{
				Address: recipient2.String(),
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin(denom1, 400),
					sdk.NewInt64Coin(denom2, 600),
				),
			},
		},
	}

	clientCtx := chain.ClientContext.WithFromAddress(sender)
	bankMultiSendGas := chain.GasLimitByMsgs(msg)
	res, err = tx.BroadcastTx(
		ctx,
		clientCtx,
		chain.TxFactory().
			WithMemo(maxMemo). // memo is set to max length here to charge as much gas as possible
			WithGas(bankMultiSendGas),
		msg)
	require.NoError(t, err)
	require.Equal(t, bankMultiSendGas, uint64(res.GasUsed))

	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	qres, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: sender.String()})
	require.NoError(t, err)
	require.Empty(t, qres.Balances)

	qres, err = bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: recipient1.String()})
	require.NoError(t, err)
	require.Len(t, qres.Balances, 2)
	require.EqualValues(t, 600, qres.Balances.AmountOf(denom1).Uint64())
	require.EqualValues(t, 400, qres.Balances.AmountOf(denom2).Uint64())

	qres, err = bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: recipient2.String()})
	require.NoError(t, err)
	require.Len(t, qres.Balances, 2)
	require.EqualValues(t, 400, qres.Balances.AmountOf(denom1).Uint64())
	require.EqualValues(t, 600, qres.Balances.AmountOf(denom2).Uint64())
}
