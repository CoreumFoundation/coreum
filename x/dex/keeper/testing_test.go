package keeper_test

import (
	sdkmath "cosmossdk.io/math"

	"github.com/CoreumFoundation/coreum/v5/x/dex/keeper"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

// defaultQuantityStep is currently equal to 10000 for default UnifiedRefAmount=10^6 and QuantityStepExponent=-2.
var defaultQuantityStep = func() sdkmath.Int {
	p := types.DefaultParams()
	quantityStep, _ := keeper.ComputeQuantityStep(
		p.DefaultUnifiedRefAmount.BigInt(),
		p.QuantityStepExponent-sdkmath.LegacyPrecision,
	)
	return sdkmath.NewIntFromBigInt(quantityStep)
}()
