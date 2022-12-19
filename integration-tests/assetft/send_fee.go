package assetft

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// TestSendFee tests send fee functionality of fungible tokens.
func TestSendFee(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	issuer := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&types.MsgIssue{},
				&banktypes.MsgSend{},
			},
		}),
		chain.Faucet.FundAccountsWithOptions(ctx, recipient1, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&banktypes.MsgSend{},
			},
		}),
		chain.Faucet.FundAccountsWithOptions(ctx, recipient2, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&banktypes.MsgSend{},
			},
		}),
	)

	// Issue an fungible token
	issueMsg := &types.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     6,
		InitialAmount: sdk.NewInt(1000),
		Description:   "ABC Description",
		Features:      []types.TokenFeature{},
		BurnRate:      sdk.NewDec(0),
		SendFee:       sdk.MustNewDecFromStr("0.10"),
	}

	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	tokenIssuedEvents, err := event.FindTypedEvents[*types.EventTokenIssued](res.Events)
	requireT.NoError(err)
	denom := tokenIssuedEvents[0].Denom

	// send from issuer to recipient1 (send fee must not apply)
	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   recipient1.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(400))),
	}

	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:     600,
		&recipient1: 400,
	})

	// send from recipient1 to recipient2 (send fee must apply)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient1.String(),
		ToAddress:   recipient2.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100))),
	}

	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:     610,
		&recipient1: 290,
		&recipient2: 100,
	})

	// send from recipient2 to issuer (send fee must not apply)
	sendMsg = &banktypes.MsgSend{
		FromAddress: recipient2.String(),
		ToAddress:   issuer.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100))),
	}

	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient2),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:     710,
		&recipient1: 290,
		&recipient2: 0,
	})

	// multi send from recipient1 to issuer and recipient2
	// (send fee must apply to both transfers. will be fixed later to apply to one transfer)
	multiSendMsg := &banktypes.MsgMultiSend{
		Inputs: []banktypes.Input{
			{Address: recipient1.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(200)))},
		},
		Outputs: []banktypes.Output{
			{Address: issuer.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100)))},
			{Address: recipient2.String(), Coins: sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(100)))},
		},
	}

	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, recipient1, testing.BalancesOptions{
			Messages: []sdk.Msg{
				multiSendMsg,
			}}),
	)

	_, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(multiSendMsg)),
		multiSendMsg,
	)
	requireT.NoError(err)
	assertCoinDistribution(ctx, chain.ClientContext, t, denom, map[*sdk.AccAddress]int64{
		&issuer:     830,
		&recipient1: 70,
		&recipient2: 100,
	})
}
