package asset

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

// TestFreezeFungibleToken checks freeze functionality of fungible tokens.
func TestGlobalFreezeFungibleToken(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	//assertT := assert.New(t)
	//clientCtx := chain.ClientContext

	//assetClient := assettypes.NewQueryClient(clientCtx)
	//bankClient := banktypes.NewQueryClient(clientCtx)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	randomAddress := chain.GenAccount()
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgIssueFungibleToken{},
				&assettypes.MsgFreezeFungibleToken{},
				&assettypes.MsgFreezeFungibleToken{},
				&assettypes.MsgUnfreezeFungibleToken{},
				&assettypes.MsgUnfreezeFungibleToken{},
				&assettypes.MsgUnfreezeFungibleToken{},
			},
		}),
		chain.Faucet.FundAccountsWithOptions(ctx, recipient, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
				&banktypes.MsgSend{},
			},
		}),
		chain.Faucet.FundAccountsWithOptions(ctx, randomAddress, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgFreezeFungibleToken{},
			},
		}),
	)

	// Issue the new fungible token
	msg := &assettypes.MsgIssueFungibleToken{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
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
	fungibleTokenIssuedEvt, err := event.FindTypedEvent[*assettypes.EventFungibleTokenIssued](res.Events)
	requireT.NoError(err)
	denom := fungibleTokenIssuedEvt.Denom

	globFreezeMsg := &assettypes.MsgGlobalFreezeFungibleToken{
		Sender: issuer.String(),
		Denom:  denom,
	}
	res, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msg)),
		globFreezeMsg,
	)
	requireT.NoError(err)
	fmt.Println(res)
	fmt.Println("-------------------------------------------")
}
