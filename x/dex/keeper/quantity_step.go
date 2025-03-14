package keeper

import (
	"math/big"

	sdkerrors "cosmossdk.io/errors"
	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func validateQuantityStep(quantity *big.Int, baseURA *big.Int, quantityStepExponent int32) error {
	quantityStep := computeQuantityStep(quantity, quantityStepExponent)
	if !isQuantityStepValid(quantity, quantityStep) {
		return sdkerrors.Wrapf(
			types.ErrInvalidInput,
			"invalid quantity %s, the quantity must be multiple of %s",
			quantity.String(), quantityStep.String(),
		)
	}

	return nil
}

func isQuantityStepValid(quantity *big.Int, quantityStep *big.Int) bool {
	remainder := cbig.IntRem(quantity, quantityStep)
	return cbig.IntEqZero(remainder)
}

func computeQuantityStep(baseURA *big.Int, quantityStepExponent int32) *big.Int {
	// quantity_step = 10^quantity_step_exponent * round_up_pow10(unified_ref_amount) =
	// 10^(quantity_step_exponent + log10_round_up(unified_ref_amount))
	exponent := int64(quantityStepExponent) + cbig.RatLog10RoundUp(cbig.NewRatFromBigInt(baseURA))
	if exponent <= 0 {
		return big.NewInt(1)
	}

	return cbig.IntTenToThePower(big.NewInt(exponent))
}
