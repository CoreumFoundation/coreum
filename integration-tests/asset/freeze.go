package asset

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

// TestFreezeFungibleToken checks freeze functionality of fungible tokens.
//
//nolint:funlen // this is a single test scenario and breaking it down is not beneficial
func TestFreezeFungibleToken(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	assertT := assert.New(t)
	clientCtx := chain.ClientContext

	assetClient := assettypes.NewQueryClient(clientCtx)
	bankClient := banktypes.NewQueryClient(clientCtx)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	randomAddress := chain.GenAccount()
	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(issuer, chain.NewCoin(sdk.NewInt(1000_000))),
		testing.NewFundedAccount(recipient, chain.NewCoin(sdk.NewInt(1000_000))),
		testing.NewFundedAccount(randomAddress, chain.NewCoin(sdk.NewInt(100_000))),
	))

	// Issue the new fungible token
	msg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "BTC",
		Description:   "BTC Description",
		Recipient:     recipient.String(),
		InitialAmount: sdk.NewInt(1000),
		Options: []assettypes.FungibleTokenOption{
			assettypes.FungibleTokenOption_Freezable, //nolint:nosnakecase
		},
	}

	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	requireT.NoError(err)
	evt := testing.FindTypedEvent(t, &assettypes.EventFungibleTokenIssued{}, res.Events)
	fungibleTokenIssuedEvt, ok := evt.(*assettypes.EventFungibleTokenIssued)
	requireT.True(ok)
	denom := fungibleTokenIssuedEvt.Denom

	// try to pass wrong signature to freeze msg
	freezeMsg := &assettypes.MsgFreezeFungibleToken{
		Issuer:  randomAddress.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(1000)),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(randomAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.Error(err)
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrUnauthorized))

	// freeze 500 tokens
	freezeMsg = &assettypes.MsgFreezeFungibleToken{
		Issuer:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(500)),
	}
	res, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.NoError(err)
	assertT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(freezeMsg))

	// query frozen tokens
	frozenBalance, err := assetClient.FrozenBalance(ctx, &assettypes.QueryFrozenBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(500)), frozenBalance.Coin)

	frozenBalances, err := assetClient.FrozenBalances(ctx, &assettypes.QueryFrozenBalancesRequest{
		Account: recipient.String(),
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(500))), frozenBalances.Coins)

	// try to send more than allowed (600)
	recipient2 := chain.GenAccount()
	sendMsg := &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(600))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.Error(err)
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrInsufficientFunds))

	// try to send allowed tokens (500)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(500))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	balance1, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(balance1.GetBalance().String(), sdk.NewCoin(denom, sdk.NewInt(500)).String())

	// unfreeze 200 tokens and try send 250 tokens
	unFreezeMsg := &assettypes.MsgUnfreezeFungibleToken{
		Issuer:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(200)),
	}
	res, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(unFreezeMsg)),
		unFreezeMsg,
	)
	requireT.NoError(err)
	assertT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(unFreezeMsg))

	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(250))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.Error(err)
	assertT.True(sdkerrors.IsOf(err, sdkerrors.ErrInsufficientFunds))

	// send allowed tokens (200)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(200))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
}
