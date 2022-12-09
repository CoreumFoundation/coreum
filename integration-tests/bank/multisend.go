package bank

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdksigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
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

// TestMultiSendFromMultipleAccounts tests MultiSend message form multiple accounts.
func TestMultiSendFromMultipleAccounts(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)

	sender1 := chain.GenAccount()
	sender1KeyInfo, err := chain.ClientContext.Keyring().KeyByAddress(sender1)
	requireT.NoError(err)

	sender2 := chain.GenAccount()
	sender2KeyInfo, err := chain.ClientContext.Keyring().KeyByAddress(sender2)
	requireT.NoError(err)

	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	amount := sdk.NewInt(1000)

	issue1Msg := &assettypes.MsgIssueFungibleToken{
		Issuer:        sender1.String(),
		Symbol:        "TOK1",
		Subunit:       "tok1",
		Description:   "TOK1 Description",
		Recipient:     sender1.String(),
		InitialAmount: amount,
	}

	issue2Msg := &assettypes.MsgIssueFungibleToken{
		Issuer:        sender2.String(),
		Symbol:        "TOK2",
		Subunit:       "tok2",
		Description:   "TOK2 Description",
		Recipient:     sender2.String(),
		InitialAmount: amount,
	}

	denom1 := assettypes.BuildFungibleTokenDenom(issue1Msg.Subunit, sender1)
	denom2 := assettypes.BuildFungibleTokenDenom(issue2Msg.Subunit, sender2)

	// define the message to send from multiple accounts to multiple
	multiSendMsg := &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{
			{
				Address: sender1.String(),
				Coins: sdk.NewCoins(
					sdk.NewInt64Coin(denom1, 1000),
				),
			},
			{
				Address: sender2.String(),
				Coins: sdk.NewCoins(
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

	// fund accounts
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, sender1, testing.BalancesOptions{
		Messages: []sdk.Msg{
			multiSendMsg,
			issue1Msg,
		},
	}))
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, sender2, testing.BalancesOptions{
		Messages: []sdk.Msg{issue2Msg},
	}))

	// issue first fungible token
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issue1Msg)),
		issue1Msg,
	)
	requireT.NoError(err)
	// issue second fungible token
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender2),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issue2Msg)),
		issue2Msg,
	)
	requireT.NoError(err)

	// create MultiSend tx message and sign it from 2 accounts
	sender1AccInfo, err := tx.GetAccountInfo(ctx, chain.ClientContext, sender1)
	requireT.NoError(err)

	// set sender1 params for the signature
	txF := chain.TxFactory().
		WithAccountNumber(sender1AccInfo.GetAccountNumber()).
		WithSequence(sender1AccInfo.GetSequence()).
		WithGas(chain.GasLimitByMsgs(multiSendMsg)).
		WithSignMode(sdksigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON) //nolint:nosnakecase // the sdk constant

	txBuilder, err := txF.BuildUnsignedTx(multiSendMsg)
	requireT.NoError(err)

	// sign from sender1
	err = tx.Sign(txF, sender1KeyInfo.GetName(), txBuilder, false)
	requireT.NoError(err)

	sender2AccInfo, err := tx.GetAccountInfo(ctx, chain.ClientContext, sender2)
	requireT.NoError(err)

	// set sender2 params for the signature
	txF = chain.TxFactory().
		WithAccountNumber(sender2AccInfo.GetAccountNumber()).
		WithSequence(sender2AccInfo.GetSequence()).
		WithGas(chain.GasLimitByMsgs(multiSendMsg)).
		WithSignMode(sdksigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON) //nolint:nosnakecase // the sdk constant

	// sign from sender2
	err = tx.Sign(txF, sender2KeyInfo.GetName(), txBuilder, false)
	requireT.NoError(err)

	// encode tx and broadcast
	encodedMultiSendTx, err := chain.ClientContext.TxConfig().TxEncoder()(txBuilder.GetTx())
	requireT.NoError(err)
	_, err = tx.BroadcastRawTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender1),
		encodedMultiSendTx)
	requireT.NoError(err)

	// check the received balances
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	sender1AllBalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: sender1.String()})
	requireT.NoError(err)
	requireT.Empty(sender1AllBalancesRes.Balances)

	sender2AllBalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: sender2.String()})
	requireT.NoError(err)
	requireT.Empty(sender2AllBalancesRes.Balances)

	recipient1AllBalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: recipient1.String()})
	requireT.NoError(err)
	requireT.Len(recipient1AllBalancesRes.Balances, 2)
	requireT.EqualValues(600, recipient1AllBalancesRes.Balances.AmountOf(denom1).Uint64())
	requireT.EqualValues(400, recipient1AllBalancesRes.Balances.AmountOf(denom2).Uint64())

	recipient2AllBalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: recipient2.String()})
	requireT.NoError(err)
	requireT.Len(recipient2AllBalancesRes.Balances, 2)
	requireT.EqualValues(400, recipient2AllBalancesRes.Balances.AmountOf(denom1).Uint64())
	requireT.EqualValues(600, recipient2AllBalancesRes.Balances.AmountOf(denom2).Uint64())
}
