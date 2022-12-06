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
	"github.com/CoreumFoundation/coreum/testutil/event"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

// TestWhitelistUnwhitelistableFungibleToken checks whitelist functionality on unwhitelistable fungible tokens.
//
//nolint:dupl // code duplication is detected between whitelisting and freezing but trying to fix this is not really helpful
func TestWhitelistUnwhitelistableFungibleToken(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgSetWhitelistedLimitFungibleToken{},
			},
		}),
	)

	// Issue an unwhitelistable fungible token
	msg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotWhitelistable",
		Description:   "ABC Description",
		Recipient:     recipient.String(),
		InitialAmount: sdk.NewInt(1000),
		Features:      []assettypes.FungibleTokenFeature{},
	}

	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvt, err := event.FindTypedEvent[*assettypes.EventFungibleTokenIssued](res.Events)
	requireT.NoError(err)
	unwhitelistableDenom := fungibleTokenIssuedEvt.Denom

	// try to whitelist unwhitelistable token
	whitelistMsg := &assettypes.MsgSetWhitelistedLimitFungibleToken{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(unwhitelistableDenom, sdk.NewInt(1000)),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	assertT.True(assettypes.ErrFeatureNotActive.Is(err))
}

// TestWhitelistFungibleToken checks whitelist functionality of fungible tokens.
func TestWhitelistFungibleToken(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	assertT := assert.New(t)
	clientCtx := chain.ClientContext

	assetClient := assettypes.NewQueryClient(clientCtx)
	bankClient := banktypes.NewQueryClient(clientCtx)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	randomAccount := chain.GenAccount()
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgSetWhitelistedLimitFungibleToken{},
				&assettypes.MsgSetWhitelistedLimitFungibleToken{},
			},
		}),
		chain.Faucet.FundAccountsWithOptions(ctx, recipient, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgSetWhitelistedLimitFungibleToken{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
			},
		}),
		chain.Faucet.FundAccountsWithOptions(ctx, randomAccount, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgSetWhitelistedLimitFungibleToken{},
			},
		}),
	)

	// Issue the new fungible token
	msg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Description:   "ABC Description",
		Recipient:     recipient.String(),
		InitialAmount: sdk.NewInt(20000),
		Features: []assettypes.FungibleTokenFeature{
			assettypes.FungibleTokenFeature_whitelist, //nolint:nosnakecase
		},
	}

	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvt, err := event.FindTypedEvent[*assettypes.EventFungibleTokenIssued](res.Events)
	requireT.NoError(err)
	denom := fungibleTokenIssuedEvt.Denom

	// try to pass non-issuer signature to whitelist msg
	whitelistMsg := &assettypes.MsgSetWhitelistedLimitFungibleToken{
		Sender:  recipient.String(),
		Account: randomAccount.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(400)),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	requireT.Error(err)
	assertT.True(sdkerrors.ErrUnauthorized.Is(err))

	// whitelist 400 tokens
	whitelistMsg = &assettypes.MsgSetWhitelistedLimitFungibleToken{
		Sender:  issuer.String(),
		Account: randomAccount.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(400)),
	}
	res, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	requireT.NoError(err)
	assertT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(whitelistMsg))

	// query whitelisted tokens
	whitelistedBalance, err := assetClient.WhitelistedBalance(ctx, &assettypes.QueryWhitelistedBalanceRequest{
		Account: randomAccount.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(400)), whitelistedBalance.Balance)

	whitelistedBalances, err := assetClient.WhitelistedBalances(ctx, &assettypes.QueryWhitelistedBalancesRequest{
		Account: randomAccount.String(),
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(400))), whitelistedBalances.Balances)

	// try to receive more than whitelisted (600) (possible 400)
	sendMsg := &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   randomAccount.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(600))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	assertT.True(assettypes.ErrWhitelistedLimitExceeded.Is(err))

	// try to send whitelisted balance (400)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   randomAccount.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(400))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: randomAccount.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(balance.GetBalance().String(), sdk.NewCoin(denom, sdk.NewInt(400)).String())

	// try to send one more
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   randomAccount.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	assertT.True(assettypes.ErrWhitelistedLimitExceeded.Is(err))

	// whitelist one more
	whitelistMsg = &assettypes.MsgSetWhitelistedLimitFungibleToken{
		Sender:  issuer.String(),
		Account: randomAccount.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(401)),
	}
	res, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	requireT.NoError(err)
	assertT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(whitelistMsg))

	// query whitelisted tokens
	whitelistedBalance, err = assetClient.WhitelistedBalance(ctx, &assettypes.QueryWhitelistedBalanceRequest{
		Account: randomAccount.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(401)), whitelistedBalance.Balance)

	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   randomAccount.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	balance, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: randomAccount.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(balance.GetBalance().String(), sdk.NewCoin(denom, sdk.NewInt(401)).String())
}
