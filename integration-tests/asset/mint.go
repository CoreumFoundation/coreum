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

// TestMintFungibleToken tests mint functionality of fungible tokens.
func TestMintFungibleToken(ctx context.Context, t testing.T, chain testing.Chain) {
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
				&assettypes.MsgMintFungibleToken{},
				&assettypes.MsgMintFungibleToken{},
			},
		}),
		chain.Faucet.FundAccountsWithOptions(ctx, randomAddress, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgMintFungibleToken{},
			},
		}),
	)

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
