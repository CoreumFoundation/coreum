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

// TestBurnFungibleToken tests burn functionality of fungible tokens.
func TestBurnFungibleToken(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	randomAddress := chain.GenAccount()
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgBurnFungibleToken{},
				&assettypes.MsgBurnFungibleToken{},
			},
		}),
		chain.Faucet.FundAccountsWithOptions(ctx, randomAddress, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgBurnFungibleToken{},
			},
		}),
	)

	// Issue an unburnable fungible token
	issueMsg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotBurnable",
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
	fungibleTokenIssuedEvt, err := event.FindTypedEvent[*assettypes.EventFungibleTokenIssued](res.Events)
	requireT.NoError(err)
	unburnable := fungibleTokenIssuedEvt.Denom

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
	fungibleTokenIssuedEvt, err = event.FindTypedEvent[*assettypes.EventFungibleTokenIssued](res.Events)
	requireT.NoError(err)
	burnableDenom := fungibleTokenIssuedEvt.Denom

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
