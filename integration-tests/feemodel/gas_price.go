package feemodel

import (
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/x/feemodel/types"
)

// TestQueryingMinGasPrice check that it's possible to query current minimum gas price required by the network.
func TestQueryingMinGasPrice(ctx context.Context, t testing.T, chain testing.Chain) {
	res, err := chain.Client.FeemodelQueryClient().MinGasPrice(ctx, &types.QueryMinGasPriceRequest{})
	require.NoError(t, err)

	logger.Get(ctx).Info("Queried minimum gas price required", zap.Stringer("gasPrice", res.MinGasPrice))

	params := chain.NetworkConfig.Fee.FeeModel.Params()
	model := types.NewModel(params)

	require.False(t, res.MinGasPrice.Amount.IsNil())
	assert.True(t, res.MinGasPrice.Amount.GTE(model.CalculateGasPriceWithMaxDiscount()))
	assert.True(t, res.MinGasPrice.Amount.LTE(params.MaxGasPrice))
	assert.Equal(t, chain.NetworkConfig.TokenSymbol, res.MinGasPrice.Denom)
}
