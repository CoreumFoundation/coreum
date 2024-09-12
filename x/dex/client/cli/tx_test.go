package cli_test

import (
	"fmt"
	"strings"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	coreumclitestutil "github.com/CoreumFoundation/coreum/v4/testutil/cli"
	"github.com/CoreumFoundation/coreum/v4/testutil/event"
	"github.com/CoreumFoundation/coreum/v4/testutil/network"
	assetftcli "github.com/CoreumFoundation/coreum/v4/x/asset/ft/client/cli"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v4/x/dex/client/cli"
	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

const (
	denom2 = "denom2"
	denom3 = "denom3"
)

func TestCmdPlaceOrder(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx
	denom1 := issueFT(ctx, requireT, testNetwork, sdkmath.NewInt(100))

	placeOrder(ctx, requireT, testNetwork, types.Order{
		ID:          "id1",
		Type:        types.ORDER_TYPE_LIMIT,
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(types.MustNewPriceFromString("123e-2")),
		Quantity:    sdkmath.NewInt(100),
		Side:        types.SIDE_SELL,
		TimeInForce: types.TIME_IN_FORCE_GTC,
	})
}

func TestCmdCancelOrder(t *testing.T) {
	requireT := require.New(t)
	testNetwork := network.New(t)

	ctx := testNetwork.Validators[0].ClientCtx
	denom1 := issueFT(ctx, requireT, testNetwork, sdkmath.NewInt(100))
	order := types.Order{
		ID:          "id1",
		Type:        types.ORDER_TYPE_LIMIT,
		BaseDenom:   denom1,
		QuoteDenom:  denom2,
		Price:       lo.ToPtr(types.MustNewPriceFromString("123e-2")),
		Quantity:    sdkmath.NewInt(100),
		Side:        types.SIDE_SELL,
		TimeInForce: types.TIME_IN_FORCE_GTC,
	}

	placeOrder(ctx, requireT, testNetwork, order)

	args := append(
		[]string{
			order.ID,
		}, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(
		ctx,
		testNetwork,
		cli.CmdCancelOrder(),
		args,
	)
	requireT.NoError(err)
}

func placeOrder(
	ctx client.Context,
	requireT *require.Assertions,
	testNetwork *network.Network,
	order types.Order,
) {
	var orderType string
	switch order.Type {
	case types.ORDER_TYPE_LIMIT:
		orderType = cli.OrderTypeLimit
	case types.ORDER_TYPE_MARKET:
		orderType = cli.OrderTypeMarket
	default:
		requireT.Fail(fmt.Sprintf("unknown type '%s'", order.Type))
	}

	args := append(
		[]string{
			orderType,
			order.ID,
			order.BaseDenom,
			order.QuoteDenom,
			order.Quantity.String(),
			order.Side.String(),
			"--" + cli.PriceFlag, order.Price.String(),
		}, txValidator1Args(testNetwork)...)
	_, err := coreumclitestutil.ExecTxCmd(
		ctx,
		testNetwork,
		cli.CmdPlaceOrder(),
		args,
	)
	requireT.NoError(err)
}

func issueFT(
	ctx client.Context,
	requireT *require.Assertions,
	testNetwork *network.Network,
	initialAmount sdkmath.Int,
) string {
	// args
	args := []string{
		"smb" + uuid.NewString()[:4],
		"unt" + uuid.NewString()[:4],
		"1",
		initialAmount.String(),
		"",
	}
	features := []string{
		assetfttypes.Feature_minting.String(),
	}

	args = append(args, fmt.Sprintf("--%s=%s", assetftcli.FeaturesFlag, strings.Join(features, ",")))

	args = append(args, txValidator1Args(testNetwork)...)
	res, err := coreumclitestutil.ExecTxCmd(ctx, testNetwork, assetftcli.CmdTxIssue(), args)
	requireT.NoError(err)

	requireT.NotEmpty(res.TxHash)
	requireT.Equal(uint32(0), res.Code, "can't submit Issue tx for query", res)

	eventIssuedName := proto.MessageName(&assetfttypes.EventIssued{})
	for i := range res.Events {
		if res.Events[i].Type != eventIssuedName {
			continue
		}
		eventsTokenIssued, err := event.FindTypedEvents[*assetfttypes.EventIssued](res.Events)
		requireT.NoError(err)
		return eventsTokenIssued[0].Denom
	}
	requireT.Failf("event: %s not found in the issue response", eventIssuedName)

	return ""
}

func txValidator1Args(testNetwork *network.Network) []string {
	return []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, validator1Address(testNetwork).String()),
		fmt.Sprintf("--%s=%s", flags.FlagFees,
			sdk.NewCoins(sdk.NewInt64Coin(testNetwork.Config.BondDenom, 1000000)).String()),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}
}

func validator1Address(testNetwork *network.Network) sdk.Address {
	return testNetwork.Validators[0].Address
}
