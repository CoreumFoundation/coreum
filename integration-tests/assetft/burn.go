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
	"github.com/CoreumFoundation/coreum/testutil/event"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// TestBurn tests burn functionality of fungible tokens.
func TestBurn(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	assertT := assert.New(t)
	issuer := chain.GenAccount()
	randomAddress := chain.GenAccount()
	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assetfttypes.MsgIssue{},
				&assetfttypes.MsgIssue{},
				&assetfttypes.MsgBurn{},
				&assetfttypes.MsgBurn{},
			},
		}))
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, randomAddress, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assetfttypes.MsgBurn{},
			},
		}))

	// Issue an unburnable fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCNotBurnable",
		Subunit:       "uabcnotburnable",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.TokenFeature{
			assetfttypes.TokenFeature_mint,   //nolint:nosnakecase
			assetfttypes.TokenFeature_freeze, //nolint:nosnakecase
		},
	}

	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvts, err := event.FindTypedEvents[*assetfttypes.EventTokenIssued](res.Events)
	requireT.NoError(err)
	unburnable := fungibleTokenIssuedEvts[0].Denom

	// try to burn unburnable token
	burnMsg := &assetfttypes.MsgBurn{
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
	requireT.True(assetfttypes.ErrFeatureNotActive.Is(err))

	// Issue a burnable fungible token
	issueMsg = &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABCBurnable",
		Subunit:       "uabcburnable",
		Precision:     6,
		Description:   "ABC Description",
		InitialAmount: sdk.NewInt(1000),
		Features:      []assetfttypes.TokenFeature{assetfttypes.TokenFeature_burn}, //nolint:nosnakecase
	}

	res, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	fungibleTokenIssuedEvts, err = event.FindTypedEvents[*assetfttypes.EventTokenIssued](res.Events)
	requireT.NoError(err)
	burnableDenom := fungibleTokenIssuedEvts[0].Denom

	// try to pass non-issuer signature to msg
	burnMsg = &assetfttypes.MsgBurn{
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

	burnMsg = &assetfttypes.MsgBurn{
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
