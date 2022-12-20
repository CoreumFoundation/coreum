package assetft

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// TestGloballyFreeze checks global freeze functionality of fungible tokens.
func TestGloballyFreeze(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	assertT := assert.New(t)

	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assetfttypes.MsgIssue{},
				&assetfttypes.MsgGloballyFreeze{},
				&banktypes.MsgSend{},
				&assetfttypes.MsgGloballyUnfreeze{},
				&banktypes.MsgSend{},
			},
		}))

	// Issue the new fungible token
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "FREEZE",
		Subunit:       "freeze",
		Precision:     6,
		Description:   "FREEZE Description",
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.TokenFeature{
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
	denom := fungibleTokenIssuedEvts[0].Denom

	// Globally freeze FT.
	globFreezeMsg := &assetfttypes.MsgGloballyFreeze{
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
	assertT.True(assetfttypes.ErrGloballyFrozen.Is(err))

	// Globally unfreeze FT.
	globUnfreezeMsg := &assetfttypes.MsgGloballyUnfreeze{
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
