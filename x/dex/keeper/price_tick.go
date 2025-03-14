package keeper

import (
	"math/big"

	sdkerrors "cosmossdk.io/errors"

	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func validatePriceTick(price *big.Rat, baseURA, quoteURA *big.Int, priceTickExponent int32) error {
	priceTick := computePriceTick(baseURA, quoteURA, priceTickExponent)
	if !isPriceTickValid(price, priceTick) {
		return sdkerrors.Wrapf(
			types.ErrInvalidInput,
			"invalid price %s, the price must be multiple of %s",
			price.String(), priceTick.String(),
		)
	}

	return nil
}

func isPriceTickValid(price *big.Rat, priceTick *big.Rat) bool {
	_, remainder := cbig.RatQuoWithIntRemainder(price, priceTick)
	return cbig.IntEqZero(remainder)
}

// computePriceTick returns the price tick of a given ref amounts and price tick exponent.
func computePriceTick(
	baseURA,
	quoteURA *big.Int,
	priceTickExponent int32,
) *big.Rat {
	// price_tick = 10^price_tick_exponent * round_up_pow10(ura_quote/ura_base) =
	// 10^(price_tick_exponent + log10_round_up(ura_quote/ura_base)
	exponent := int64(priceTickExponent) + cbig.RatLog10RoundUp(cbig.NewRatFromBigInts(quoteURA, baseURA))

	return cbig.RatTenToThePower(exponent)
}
