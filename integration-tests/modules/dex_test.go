//go:build integrationtests

package modules

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	"github.com/CoreumFoundation/coreum/v4/pkg/client"
	"github.com/CoreumFoundation/coreum/v4/testutil/integration"
	dextypes "github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

// TestPlaceAndGetOrder tests the dex modules ability to save and read an order.
func TestPlaceAndGetOrder(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	acc := chain.GenAccount()
	requireT := require.New(t)

	chain.FundAccountWithOptions(ctx, t, acc, integration.BalancesOptions{
		Messages: []sdk.Msg{&dextypes.MsgPlaceOrder{}},
	})

	price, err := dextypes.NewPriceFromString("123e-1")
	require.NoError(t, err)

	placeOrderMsg := &dextypes.MsgPlaceOrder{
		Sender:     acc.String(),
		ID:         "id1",
		BaseDenom:  "denom1",
		QuoteDenom: "denom2",
		Price:      price,
		Quantity:   sdkmath.NewInt(123),
		Side:       dextypes.Side_buy,
	}

	txResult, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(acc),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(placeOrderMsg)),
		placeOrderMsg,
	)
	requireT.NoError(err)
	// validate the deterministic gas
	requireT.Equal(chain.GasLimitByMsgs(placeOrderMsg), uint64(txResult.GasUsed))
}
