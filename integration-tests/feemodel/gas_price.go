package feemodel

import (
	"context"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	feemodeltypes "github.com/CoreumFoundation/coreum/x/feemodel/types"
)

// TestQueryingMinGasPrice check that it's possible to query current minimum gas price required by the network.
func TestQueryingMinGasPrice(ctx context.Context, t testing.T, chain testing.Chain) {
	feemodelClient := feemodeltypes.NewQueryClient(chain.ClientContext)
	res, err := feemodelClient.MinGasPrice(ctx, &feemodeltypes.QueryMinGasPriceRequest{})
	require.NoError(t, err)

	logger.Get(ctx).Info("Queried minimum gas price required", zap.Stringer("gasPrice", res.MinGasPrice))

	params := chain.NetworkConfig.Fee.FeeModel.Params()
	model := feemodeltypes.NewModel(params)

	require.False(t, res.MinGasPrice.Amount.IsNil())
	assert.True(t, res.MinGasPrice.Amount.GTE(model.CalculateGasPriceWithMaxDiscount()))
	assert.True(t, res.MinGasPrice.Amount.LTE(model.CalculateMaxGasPrice()))
	assert.Equal(t, chain.NetworkConfig.BaseDenom, res.MinGasPrice.Denom)
}
