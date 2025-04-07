package keeper

import (
	"math/big"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"

	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func validateQuantityStep(quantity *big.Int, baseURA sdkmath.LegacyDec, quantityStepExponent int32) error {
	baseURABigInt := baseURA.BigInt()

	// Since LegacyDec is multiplied by 10^LegacyPrecision when converting to BigInt,
	// we have to divide by same number by subtracting LegacyPrecision from exponent.
	quantityStep, exponent := ComputeQuantityStep(baseURABigInt, quantityStepExponent-sdkmath.LegacyPrecision)
	if !isQuantityStepValid(quantity, quantityStep) {
		return sdkerrors.Wrapf(
			types.ErrInvalidInput,
			"invalid quantity, has to be multiple of quantity step: 10^%d",
			exponent,
		)
	}

	return nil
}

func isQuantityStepValid(quantity *big.Int, quantityStep *big.Int) bool {
	remainder := cbig.IntRem(quantity, quantityStep)
	return cbig.IntEqZero(remainder)
}

// ComputeQuantityStep returns quantity step for an asset by unified_ref_amount and price_tick_exponent.
func ComputeQuantityStep(baseURA *big.Int, quantityStepExponent int32) (*big.Int, int64) {
	// quantity_step = max(1, 10^(quantity_step_exponent + ceil(log10(unified_ref_amount))))
	exponent := int64(quantityStepExponent) + cbig.RatLog10RoundUp(cbig.NewRatFromBigInt(baseURA))
	if exponent < 0 {
		exponent = 0
	}

	return cbig.IntTenToThePower(big.NewInt(exponent)), exponent
}
