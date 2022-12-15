package asset

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

// TestGloballyFreezeFungibleToken checks global freeze functionality of fungible tokens.
func TestGloballyFreezeFungibleToken(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	assertT := assert.New(t)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
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
