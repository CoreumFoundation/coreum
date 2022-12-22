package assetft

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// TestWhitelistUnwhitelistable checks whitelist functionality on unwhitelistable fungible tokens.
func TestWhitelistUnwhitelistable(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assetfttypes.MsgIssue{},
				&assetfttypes.MsgSetWhitelistedLimit{},
			},
		}))

	// Issue an unwhitelistable fungible token
	subunit := "uabcnotwhitelistable"
	unwhitelistableDenom := assetfttypes.BuildDenom(subunit, issuer)
	amount := sdk.NewInt(1000)
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotWhitelistable",
		Subunit:       "uabcnotwhitelistable",
		Description:   "ABC Description",
		InitialAmount: amount,
		Features:      []assetfttypes.TokenFeature{},
	}

	_, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	requireT.NoError(err)

	// try to whitelist unwhitelistable token
	whitelistMsg := &assetfttypes.MsgSetWhitelistedLimit{
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
	assertT.True(assetfttypes.ErrFeatureNotActive.Is(err))
}

// TestWhitelist checks whitelist functionality of fungible tokens.
func TestWhitelist(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	assertT := assert.New(t)
	clientCtx := chain.ClientContext

	ftClient := assetfttypes.NewQueryClient(clientCtx)
	bankClient := banktypes.NewQueryClient(clientCtx)

	issuer := chain.GenAccount()
	nonIssuer := chain.GenAccount()
	recipient := chain.GenAccount()
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assetfttypes.MsgIssue{},
				&assetfttypes.MsgSetWhitelistedLimit{},
				&assetfttypes.MsgSetWhitelistedLimit{},
				&assetfttypes.MsgSetWhitelistedLimit{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
			},
		}))
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, nonIssuer, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assetfttypes.MsgSetWhitelistedLimit{},
			},
		}))
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, recipient, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assetfttypes.MsgSetWhitelistedLimit{},
			},
		}))

	// Issue the new fungible token
	amount := sdk.NewInt(20000)
	subunit := "uabc"
	denom := assetfttypes.BuildDenom(subunit, issuer)
	msg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: amount,
		Features: []assetfttypes.TokenFeature{
			assetfttypes.TokenFeature_whitelist, //nolint:nosnakecase
		},
	}
	_, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	requireT.NoError(err)

	// try to pass non-issuer signature to whitelist msg
	whitelistMsg := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  nonIssuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(400)),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(nonIssuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	requireT.Error(err)
	assertT.True(sdkerrors.ErrUnauthorized.Is(err))

	// try to send to recipient before it is whitelisted (balance 0, whitelist limit 0)
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(10))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	assertT.True(assetfttypes.ErrWhitelistedLimitExceeded.Is(err))

	// whitelist 400 tokens
	whitelistMsg = &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(400)),
	}
	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	requireT.NoError(err)
	assertT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(whitelistMsg))

	// query whitelisted tokens
	whitelistedBalance, err := ftClient.WhitelistedBalance(ctx, &assetfttypes.QueryWhitelistedBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(400)), whitelistedBalance.Balance)

	whitelistedBalances, err := ftClient.WhitelistedBalances(ctx, &assetfttypes.QueryWhitelistedBalancesRequest{
		Account: recipient.String(),
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(400))), whitelistedBalances.Balances)

	// try to receive more than whitelisted (600) (possible 400)
	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(600))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	assertT.True(assetfttypes.ErrWhitelistedLimitExceeded.Is(err))

	// try to send whitelisted balance (400)
	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(400))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(400)).String(), balance.GetBalance().String())

	// try to send one more
	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	assertT.True(assetfttypes.ErrWhitelistedLimitExceeded.Is(err))

	// whitelist one more
	whitelistMsg = &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: recipient.String(),
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
	whitelistedBalance, err = ftClient.WhitelistedBalance(ctx, &assetfttypes.QueryWhitelistedBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(401)), whitelistedBalance.Balance)

	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	balance, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(401)).String(), balance.GetBalance().String())

	// Verify that issuer has no whitelisted balance
	whitelistedBalance, err = ftClient.WhitelistedBalance(ctx, &assetfttypes.QueryWhitelistedBalanceRequest{
		Account: issuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdk.ZeroInt()), whitelistedBalance.Balance)

	// Send something to issuer, it should succeed despite the fact that issuer is not whitelisted
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   issuer.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(10))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	balance, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: issuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(19609)).String(), balance.GetBalance().String())

	// Set whitelisted balance to 0 for recipient
	whitelistMsg = &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.ZeroInt()),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(whitelistMsg)),
		whitelistMsg,
	)
	requireT.NoError(err)

	// Transfer to recipient should fail now
	sendMsg = &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.OneInt())),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	assertT.True(assetfttypes.ErrWhitelistedLimitExceeded.Is(err))
}
