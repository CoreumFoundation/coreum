//go:build integrationtests

package upgrade

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v5/integration-tests"
	"github.com/CoreumFoundation/coreum/v5/pkg/client"
	"github.com/CoreumFoundation/coreum/v5/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	dextypes "github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

type dex struct {
	ftDenom string
	acc1    sdk.AccAddress
}

func (d *dex) Before(t *testing.T) {
	t.Logf("Checking DEX before upgrade")

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	// ********** check params before **********
	_, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.ErrorContains(err, "unknown service coreum.dex.v1.Query")

	// ********** issue FT to use for trading before **********

	issuer := chain.GenAccount()
	d.acc1 = chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&banktypes.MsgSend{},
		},
	})

	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        issuer.String(),
		Symbol:        "ABC",
		Subunit:       "uabc",
		Precision:     1,
		InitialAmount: sdkmath.NewInt(100_000),
		Features:      []assetfttypes.Feature{},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	// build and set FT denoms
	d.ftDenom = assetfttypes.BuildDenom(issueMsg.Subunit, issuer)

	sendMsg := &banktypes.MsgSend{
		FromAddress: issuer.String(),
		ToAddress:   d.acc1.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(d.ftDenom, sdkmath.NewInt(10_000))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
}

func (d *dex) After(t *testing.T) {
	t.Logf("Checking DEX after upgrade")

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	dexClient := dextypes.NewQueryClient(chain.ClientContext)

	// ********** check params after **********

	paramsRes, err := dexClient.Params(ctx, &dextypes.QueryParamsRequest{})
	requireT.NoError(err)
	t.Logf("DEX params after upgrade: %v", paramsRes.Params)
	expectedParams := dextypes.DefaultParams()
	expectedParams.OrderReserve = sdk.NewInt64Coin(chain.ChainSettings.Denom, 10_000_000)
	requireT.Equal(expectedParams, paramsRes.Params)

	// ********** issue FT to use for trading after **********

	// fund accounts to place an order with the issued denom, and check matching

	chain.FundAccountWithOptions(ctx, t, d.acc1, integration.BalancesOptions{
		// reserve + for gas to place an order
		Amount: paramsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(500_000)),
	})
	acc2 := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, acc2, integration.BalancesOptions{
		// reserve + for gas to place an order + amount to spend on the order
		Amount: paramsRes.Params.OrderReserve.Amount.Add(sdkmath.NewInt(1_000_000)),
	})

	acc1Balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: d.acc1.String(),
		Denom:   d.ftDenom,
	})
	requireT.NoError(err)
	requireT.Equal(acc1Balance.Balance.Amount.String(), sdkmath.NewInt(10_000).String())

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(d.acc1),
		chain.TxFactoryAuto(),
		&dextypes.MsgPlaceOrder{
			Sender:      d.acc1.String(),
			Type:        dextypes.ORDER_TYPE_LIMIT,
			ID:          "id1",
			BaseDenom:   d.ftDenom,
			QuoteDenom:  chain.ChainSettings.Denom,
			Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1")),
			Quantity:    sdkmath.NewInt(10_000),
			Side:        dextypes.SIDE_SELL,
			TimeInForce: dextypes.TIME_IN_FORCE_GTC,
		},
	)
	requireT.NoError(err)

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc2),
		chain.TxFactoryAuto(),
		&dextypes.MsgPlaceOrder{
			Sender:      acc2.String(),
			Type:        dextypes.ORDER_TYPE_LIMIT,
			ID:          "id1",
			BaseDenom:   d.ftDenom,
			QuoteDenom:  chain.ChainSettings.Denom,
			Price:       lo.ToPtr(dextypes.MustNewPriceFromString("1")),
			Quantity:    sdkmath.NewInt(10_000),
			Side:        dextypes.SIDE_BUY,
			TimeInForce: dextypes.TIME_IN_FORCE_GTC,
		},
	)
	requireT.NoError(err)

	acc1Balance, err = bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: d.acc1.String(),
		Denom:   d.ftDenom,
	})
	requireT.NoError(err)
	requireT.True(acc1Balance.Balance.IsZero())

	acc2Balance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: acc2.String(),
		Denom:   d.ftDenom,
	})
	requireT.NoError(err)
	requireT.Equal(acc2Balance.Balance.Amount.String(), sdkmath.NewInt(10_000).String())
}
