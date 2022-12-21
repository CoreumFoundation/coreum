package assetft

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// TestBurnRate tests burn rate functionality of fungible tokens.
func TestBurnRate(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	issuer := chain.GenAccount()
	recipient1 := chain.GenAccount()
	recipient2 := chain.GenAccount()

	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assetfttypes.MsgIssue{},
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
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "abc",
		Precision:     6,
		InitialAmount: sdk.NewInt(1000),
		Description:   "ABC Description",
		Features:      []assetfttypes.TokenFeature{},
		BurnRate:      sdk.MustNewDecFromStr("0.10")}

	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)

	requireT.NoError(err)
	tokenIssuedEvents, err := event.FindTypedEvents[*assetfttypes.EventTokenIssued](res.Events)
	requireT.NoError(err)
	denom := tokenIssuedEvents[0].Denom

	// send from issuer to recipient1 (burn must not apply)
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

	// send from recipient1 to recipient2 (burn must apply)
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
		&issuer:     600,
		&recipient1: 290,
		&recipient2: 100,
	})

	// send from recipient2 to issuer (burn must not apply)
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
		&issuer:     700,
		&recipient1: 290,
		&recipient2: 0,
	})

	// multi send from recipient1 to issuer and recipient2
	// (burn must apply to both transfers. will be fixed later to apply to one transfer)
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
		&issuer:     800,
		&recipient1: 70,
		&recipient2: 100,
	})
}

func assertCoinDistribution(ctx context.Context, clientCtx tx.ClientContext, t testing.T, denom string, dist map[*sdk.AccAddress]int64) {
	bankClient := banktypes.NewQueryClient(clientCtx)
	requireT := require.New(t)

	total := int64(0)
	for acc, expectedBalance := range dist {
		total += expectedBalance
		getBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{Address: acc.String(), Denom: denom})
		requireT.NoError(err)
		requireT.Equal(sdk.NewCoin(denom, sdk.NewInt(expectedBalance)).String(), getBalance.Balance.String())
	}

	supply, err := bankClient.SupplyOf(ctx, &banktypes.QuerySupplyOfRequest{Denom: denom})
	requireT.NoError(err)
	requireT.EqualValues(sdk.NewCoin(denom, sdk.NewInt(total)).String(), supply.Amount.String())
}
