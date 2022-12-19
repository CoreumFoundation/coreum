//go:build integrationtests

package modules

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

// TestAssetBurnFungibleToken tests burn functionality of fungible tokens.
func TestAssetBurnFungibleToken(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	randomAddress := chain.GenAccount()
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgBurnFungibleToken{},
				&assettypes.MsgBurnFungibleToken{},
			},
		}))
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, randomAddress, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgBurnFungibleToken{},
			},
		}))

	// Issue an unburnable fungible token
	issueMsg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotBurnable",
		Subunit:       "uabcnotburnable",
		Precision:     6,
		Description:   "ABC Description",
		Recipient:     issuer.String(),
		InitialAmount: sdk.NewInt(1000),
		Features: []assettypes.FungibleTokenFeature{
			assettypes.FungibleTokenFeature_mint,   //nolint:nosnakecase
			assettypes.FungibleTokenFeature_freeze, //nolint:nosnakecase
		},
	}

	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assettypes.EventFungibleTokenIssued](res.Events)
	requireT.NoError(err)
	unburnable := fungibleTokenIssuedEvts[0].Denom

	// try to burn unburnable token
	burnMsg := &assettypes.MsgBurnFungibleToken{
		Sender: issuer.String(),
		Coin: sdk.Coin{
			Denom:  unburnable,
			Amount: sdk.NewInt(1000),
		},
	}

	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.True(assettypes.ErrFeatureNotActive.Is(err))

	// Issue a burnable fungible token
	issueMsg = &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "ABCBurnable",
		Subunit:       "uabcburnable",
		Precision:     6,
		Description:   "ABC Description",
		Recipient:     issuer.String(),
		InitialAmount: sdk.NewInt(1000),
		Features:      []assettypes.FungibleTokenFeature{assettypes.FungibleTokenFeature_burn}, //nolint:nosnakecase
	}

	res, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvts, err = event.FindTypedEvents[*assettypes.EventFungibleTokenIssued](res.Events)
	requireT.NoError(err)
	burnableDenom := fungibleTokenIssuedEvts[0].Denom

	// try to pass non-issuer signature to msg
	burnMsg = &assettypes.MsgBurnFungibleToken{
		Sender: randomAddress.String(),
		Coin:   sdk.NewCoin(burnableDenom, sdk.NewInt(1000)),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(randomAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.Error(err)
	assertT.True(sdkerrors.ErrUnauthorized.Is(err))

	// burn tokens and check balance and total supply
	oldSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: burnableDenom})
	requireT.NoError(err)
	burnCoin := sdk.NewCoin(burnableDenom, sdk.NewInt(600))

	burnMsg = &assettypes.MsgBurnFungibleToken{
		Sender: issuer.String(),
		Coin:   burnCoin,
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.NoError(err)

	balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{Address: issuer.String(), Denom: burnableDenom})
	requireT.NoError(err)
	assertT.EqualValues(sdk.NewCoin(burnableDenom, sdk.NewInt(1000)).Sub(burnCoin).String(), balance.GetBalance().String())

	newSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: burnableDenom})
	requireT.NoError(err)
	assertT.EqualValues(burnCoin, oldSupply.GetAmount().Sub(newSupply.GetAmount()))
}

// TestAssetFreezeUnfreezableFungibleToken checks freeze functionality on unfreezable fungible tokens.
//
//nolint:dupl
func TestAssetFreezeUnfreezableFungibleToken(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgFreezeFungibleToken{},
			},
		}))

	// Issue an unfreezable fungible token
	msg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotFreezable",
		Subunit:       "uabcnotfreezable",
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
	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assettypes.EventFungibleTokenIssued](res.Events)
	requireT.NoError(err)
	unfreezableDenom := fungibleTokenIssuedEvts[0].Denom

	// try to freeze unfreezable token
	freezeMsg := &assettypes.MsgFreezeFungibleToken{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(unfreezableDenom, sdk.NewInt(1000)),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	assertT.True(assettypes.ErrFeatureNotActive.Is(err))
}

// TestAssetFreezeFungibleToken checks freeze functionality of fungible tokens.
func TestAssetFreezeFungibleToken(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	clientCtx := chain.ClientContext

	assetClient := assettypes.NewQueryClient(clientCtx)
	bankClient := banktypes.NewQueryClient(clientCtx)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	randomAddress := chain.GenAccount()
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgFreezeFungibleToken{},
				&assettypes.MsgFreezeFungibleToken{},
				&assettypes.MsgUnfreezeFungibleToken{},
				&assettypes.MsgUnfreezeFungibleToken{},
				&assettypes.MsgUnfreezeFungibleToken{},
			},
		}))
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, recipient, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
			},
		}))
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, randomAddress, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgFreezeFungibleToken{},
			},
		}))

	// Issue the new fungible token
	msg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
		Description:   "ABC Description",
		Recipient:     recipient.String(),
		InitialAmount: sdk.NewInt(1000),
		Features: []assettypes.FungibleTokenFeature{
			assettypes.FungibleTokenFeature_freeze, //nolint:nosnakecase
		},
	}

	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		msg,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assettypes.EventFungibleTokenIssued](res.Events)
	requireT.NoError(err)
	denom := fungibleTokenIssuedEvts[0].Denom

	// try to pass non-issuer signature to freeze msg
	freezeMsg := &assettypes.MsgFreezeFungibleToken{
		Sender:  randomAddress.String(),
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
	assertT.True(sdkerrors.ErrUnauthorized.Is(err))

	// freeze 400 tokens
	freezeMsg = &assettypes.MsgFreezeFungibleToken{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(400)),
	}
	res, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(freezeMsg)),
		freezeMsg,
	)
	requireT.NoError(err)
	assertT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(freezeMsg))

	fungibleTokenFreezeEvts, err := event.FindTypedEvents[*assettypes.EventFungibleTokenFrozenAmountChanged](res.Events)
	requireT.NoError(err)
	assertT.EqualValues(&assettypes.EventFungibleTokenFrozenAmountChanged{
		Account:        recipient.String(),
		PreviousAmount: sdk.NewCoin(denom, sdk.NewInt(0)),
		CurrentAmount:  sdk.NewCoin(denom, sdk.NewInt(400)),
	}, fungibleTokenFreezeEvts[0])

	// query frozen tokens
	frozenBalance, err := assetClient.FrozenBalance(ctx, &assettypes.QueryFrozenBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(400)), frozenBalance.Balance)

	frozenBalances, err := assetClient.FrozenBalances(ctx, &assettypes.QueryFrozenBalancesRequest{
		Account: recipient.String(),
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(400))), frozenBalances.Balances)

	// try to send more than available (650) (600 is available)
	recipient2 := chain.GenAccount()
	sendMsg := &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(650))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.Error(err)
	assertT.True(sdkerrors.ErrInsufficientFunds.Is(err))

	// try to send available tokens (600)
	sendMsg = &banktypes.MsgSend{
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
	requireT.NoError(err)
	balance1, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(400)).String(), balance1.GetBalance().String())

	// unfreeze 200 tokens and try send 250 tokens
	unfreezeMsg := &assettypes.MsgUnfreezeFungibleToken{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(200)),
	}
	res, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(unfreezeMsg)),
		unfreezeMsg,
	)
	requireT.NoError(err)
	assertT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(unfreezeMsg))

	fungibleTokenFreezeEvts, err = event.FindTypedEvents[*assettypes.EventFungibleTokenFrozenAmountChanged](res.Events)
	requireT.NoError(err)
	assertT.EqualValues(&assettypes.EventFungibleTokenFrozenAmountChanged{
		Account:        recipient.String(),
		PreviousAmount: sdk.NewCoin(denom, sdk.NewInt(400)),
		CurrentAmount:  sdk.NewCoin(denom, sdk.NewInt(200)),
	}, fungibleTokenFreezeEvts[0])

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
	assertT.True(sdkerrors.ErrInsufficientFunds.Is(err))

	// send available tokens (200)
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

	// unfreeze 400 tokens (frozen balance is 200), it should give error
	unfreezeMsg = &assettypes.MsgUnfreezeFungibleToken{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(400)),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(unfreezeMsg)),
		unfreezeMsg,
	)
	requireT.True(assettypes.ErrNotEnoughBalance.Is(err))

	// unfreeze 200 tokens and observer current frozen amount is zero
	unfreezeMsg = &assettypes.MsgUnfreezeFungibleToken{
		Sender:  issuer.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(200)),
	}
	res, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(unfreezeMsg)),
		unfreezeMsg,
	)
	requireT.NoError(err)
	assertT.EqualValues(res.GasUsed, chain.GasLimitByMsgs(unfreezeMsg))

	fungibleTokenFreezeEvts, err = event.FindTypedEvents[*assettypes.EventFungibleTokenFrozenAmountChanged](res.Events)
	requireT.NoError(err)
	assertT.EqualValues(&assettypes.EventFungibleTokenFrozenAmountChanged{
		Account:        recipient.String(),
		PreviousAmount: sdk.NewCoin(denom, sdk.NewInt(200)),
		CurrentAmount:  sdk.NewCoin(denom, sdk.NewInt(0)),
	}, fungibleTokenFreezeEvts[0])
}

// TestAssetGloballyFreezeFungibleToken checks global freeze functionality of fungible tokens.
func TestAssetGloballyFreezeFungibleToken(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgGloballyFreezeFungibleToken{},
				&banktypes.MsgSend{},
				&assettypes.MsgGloballyUnfreezeFungibleToken{},
				&banktypes.MsgSend{},
			},
		}))

	// Issue the new fungible token
	issueMsg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "FREEZE",
		Subunit:       "freeze",
		Precision:     6,
		Description:   "FREEZE Description",
		Recipient:     issuer.String(),
		InitialAmount: sdk.NewInt(1000),
		Features: []assettypes.FungibleTokenFeature{
			assettypes.FungibleTokenFeature_freeze, //nolint:nosnakecase
		},
	}
	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assettypes.EventFungibleTokenIssued](res.Events)
	requireT.NoError(err)
	denom := fungibleTokenIssuedEvts[0].Denom

	// Globally freeze FT.
	globFreezeMsg := &assettypes.MsgGloballyFreezeFungibleToken{
		Sender: issuer.String(),
		Denom:  denom,
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(globFreezeMsg)),
		globFreezeMsg,
	)
	requireT.NoError(err)

	// Try to send FT.
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(50))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.Error(err)
	assertT.True(assettypes.ErrGloballyFrozen.Is(err))

	// Globally unfreeze FT.
	globUnfreezeMsg := &assettypes.MsgGloballyUnfreezeFungibleToken{
		Sender: issuer.String(),
		Denom:  denom,
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(globUnfreezeMsg)),
		globUnfreezeMsg,
	)
	requireT.NoError(err)

	// Try to send FT.
	sendMsg2 := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(55))),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg2)),
		sendMsg2,
	)
	requireT.NoError(err)
}

// TestAssetIssueBasicFungibleToken checks that fungible token is issued.
func TestAssetIssueBasicFungibleToken(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)
	clientCtx := chain.ClientContext

	assetClient := assettypes.NewQueryClient(clientCtx)
	bankClient := banktypes.NewQueryClient(clientCtx)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	requireT.NoError(chain.Faucet.FundAccountsWithOptions(ctx, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&assettypes.MsgIssueFungibleToken{}},
	}))

	// Issue the new fungible token
	msg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "WBTC",
		Subunit:       "wsatoshi",
		Precision:     8,
		Description:   "Wrapped BTC",
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
		Denom:         assettypes.BuildFungibleTokenDenom(msg.Subunit, issuer),
		Issuer:        msg.Issuer,
		Symbol:        msg.Symbol,
		Precision:     msg.Precision,
		Subunit:       msg.Subunit,
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
		Subunit:     "wsatoshi",
		Precision:   8,
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

// TestAssetMintFungibleToken tests mint functionality of fungible tokens.
func TestAssetMintFungibleToken(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	randomAddress := chain.GenAccount()
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgMintFungibleToken{},
				&assettypes.MsgMintFungibleToken{},
			},
		}))
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, randomAddress, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgMintFungibleToken{},
			},
		}))

	// Issue an unmintable fungible token
	issueMsg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotMintable",
		Subunit:       "uabcnotmintable",
		Precision:     6,
		Description:   "ABC Description",
		Recipient:     issuer.String(),
		InitialAmount: sdk.NewInt(1000),
		Features: []assettypes.FungibleTokenFeature{
			assettypes.FungibleTokenFeature_burn,   //nolint:nosnakecase
			assettypes.FungibleTokenFeature_freeze, //nolint:nosnakecase
		},
	}

	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assettypes.EventFungibleTokenIssued](res.Events)
	requireT.NoError(err)
	unmintableDenom := fungibleTokenIssuedEvts[0].Denom

	// try to mint unmintable token
	mintMsg := &assettypes.MsgMintFungibleToken{
		Sender: issuer.String(),
		Coin: sdk.Coin{
			Denom:  unmintableDenom,
			Amount: sdk.NewInt(1000),
		},
	}

	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.True(assettypes.ErrFeatureNotActive.Is(err))

	// Issue a mintable fungible token
	issueMsg = &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "ABCMintable",
		Subunit:       "uabcmintable",
		Precision:     6,
		Description:   "ABC Description",
		Recipient:     issuer.String(),
		InitialAmount: sdk.NewInt(1000),
		Features:      []assettypes.FungibleTokenFeature{assettypes.FungibleTokenFeature_mint}, //nolint:nosnakecase
	}

	res, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvts, err = event.FindTypedEvents[*assettypes.EventFungibleTokenIssued](res.Events)
	requireT.NoError(err)
	mintableDenom := fungibleTokenIssuedEvts[0].Denom

	// try to pass non-issuer signature to msg
	mintMsg = &assettypes.MsgMintFungibleToken{
		Sender: randomAddress.String(),
		Coin:   sdk.NewCoin(mintableDenom, sdk.NewInt(1000)),
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(randomAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.Error(err)
	assertT.True(sdkerrors.ErrUnauthorized.Is(err))

	// mint tokens and check balance and total supply
	oldSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: mintableDenom})
	requireT.NoError(err)
	mintCoin := sdk.NewCoin(mintableDenom, sdk.NewInt(1600))
	mintMsg = &assettypes.MsgMintFungibleToken{
		Sender: issuer.String(),
		Coin:   mintCoin,
	}
	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.NoError(err)

	balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{Address: issuer.String(), Denom: mintableDenom})
	requireT.NoError(err)
	assertT.EqualValues(mintCoin.Add(sdk.NewCoin(mintableDenom, sdk.NewInt(1000))).String(), balance.GetBalance().String())

	newSupply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: mintableDenom})
	requireT.NoError(err)
	assertT.EqualValues(mintCoin, newSupply.GetAmount().Sub(oldSupply.GetAmount()))
}

// TestAssetWhitelistUnwhitelistableFungibleToken checks whitelist functionality on unwhitelistable fungible tokens.
//
//nolint:dupl
func TestAssetWhitelistUnwhitelistableFungibleToken(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgSetWhitelistedLimitFungibleToken{},
			},
		}))

	// Issue an unwhitelistable fungible token
	msg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotWhitelistable",
		Subunit:       "uabcnotwhitelistable",
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
	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assettypes.EventFungibleTokenIssued](res.Events)
	requireT.NoError(err)
	unwhitelistableDenom := fungibleTokenIssuedEvts[0].Denom

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

// TestAssetWhitelistFungibleToken checks whitelist functionality of fungible tokens.
func TestAssetWhitelistFungibleToken(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)
	assertT := assert.New(t)
	clientCtx := chain.ClientContext

	assetClient := assettypes.NewQueryClient(clientCtx)
	bankClient := banktypes.NewQueryClient(clientCtx)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	randomAccount := chain.GenAccount()
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgSetWhitelistedLimitFungibleToken{},
				&assettypes.MsgSetWhitelistedLimitFungibleToken{},
				&assettypes.MsgSetWhitelistedLimitFungibleToken{},
				&banktypes.MsgSend{},
			},
		}))
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, recipient, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgSetWhitelistedLimitFungibleToken{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
			},
		}))
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, randomAccount, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgSetWhitelistedLimitFungibleToken{},
			},
		}))

	// Issue the new fungible token
	msg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     6,
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
	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assettypes.EventFungibleTokenIssued](res.Events)
	requireT.NoError(err)
	denom := fungibleTokenIssuedEvts[0].Denom

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
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(400)).String(), balance.GetBalance().String())

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
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(401)).String(), balance.GetBalance().String())

	// Verify that issuer has no whitelisted balance
	whitelistedBalance, err = assetClient.WhitelistedBalance(ctx, &assettypes.QueryWhitelistedBalanceRequest{
		Account: issuer.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdk.ZeroInt()), whitelistedBalance.Balance)

	// Send something to issuer, it should succeed despite the fact that issuer is not whitelisted
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient.String(),
		ToAddress:   issuer.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(19599))),
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
	requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(19599)).String(), balance.GetBalance().String())

	// Ensure that recipient holds 0
	balance, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(denom, sdk.ZeroInt()).String(), balance.GetBalance().String())

	// Set whitelisted balance to 0 for recipient
	whitelistMsg = &assettypes.MsgSetWhitelistedLimitFungibleToken{
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
	assertT.True(assettypes.ErrWhitelistedLimitExceeded.Is(err))
}
