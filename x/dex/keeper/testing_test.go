package keeper_test

import (
	"math/big"

	sdkmath "cosmossdk.io/math"

	"github.com/CoreumFoundation/coreum/v5/x/dex/keeper"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

// defaultQuantityStep is currently equal to 10000 for default UnifiedRefAmount=10^6 and QuantityStepExponent=-2.
var defaultQuantityStep = func() sdkmath.Int {
	p := types.DefaultParams()
	return sdkmath.NewIntFromBigInt(keeper.ComputeQuantityStep(p.DefaultUnifiedRefAmount.BigInt(), p.QuantityStepExponent-sdkmath.LegacyPrecision))
}()

var defaultPriceTick = func() *big.Rat {
	p := types.DefaultParams()
	return keeper.ComputePriceTick(
		p.DefaultUnifiedRefAmount.BigInt(),
		p.DefaultUnifiedRefAmount.BigInt(),
		p.PriceTickExponent,
	)
}
