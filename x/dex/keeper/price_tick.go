package keeper

import (
	"errors"
	"math/big"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

func (k Keeper) getPriceTick(ctx sdk.Context, params types.Params, baseDenom, quoteDenom string) (*big.Rat, error) {
	baseURA, err := k.getAssetFTUnifiedRefAmount(ctx, baseDenom, params.DefaultUnifiedRefAmount)
	if err != nil {
		return nil, err
	}

	quoteURA, err := k.getAssetFTUnifiedRefAmount(ctx, quoteDenom, params.DefaultUnifiedRefAmount)
	if err != nil {
		return nil, err
	}

	return computePriceTick(baseURA.BigInt(), quoteURA.BigInt(), params.PriceTickExponent), nil
}

func (k Keeper) validatePriceTick(
	ctx sdk.Context,
	params types.Params,
	baseDenom, quoteDenom string,
	price types.Price,
) error {
	priceTickRat, err := k.getPriceTick(ctx, params, baseDenom, quoteDenom)
	if err != nil {
		return err
	}

	return validatePriceTick(price.Rat(), priceTickRat)
}

func (k Keeper) getAssetFTUnifiedRefAmount(
	ctx sdk.Context,
	denom string,
	defaultVal sdkmath.LegacyDec,
) (sdkmath.LegacyDec, error) {
	settings, err := k.assetFTKeeper.GetDEXSettings(ctx, denom)
	if err != nil {
		if !errors.Is(err, assetfttypes.ErrDEXSettingsNotFound) {
			return sdkmath.LegacyDec{}, err
		}
		return defaultVal, nil
	}
	if settings.UnifiedRefAmount == nil {
		return defaultVal, nil
	}

	return *settings.UnifiedRefAmount, nil
}

func validatePriceTick(price *big.Rat, priceTick *big.Rat) error {
	_, remainder := cbig.RatQuoWithIntRemainder(price, priceTick)
	if !cbig.IntEqZero(remainder) {
		return sdkerrors.Wrapf(
			types.ErrInvalidInput,
			"invalid price %s, the price must be multiple of %s",
			price.String(), priceTick.String(),
		)
	}

	return nil
}

// computePriceTick returns the price tick of a given ref amounts and price tick exponent.
func computePriceTick(
	baseURA,
	quoteURA *big.Int,
	priceTickExponent int32,
) *big.Rat {
	// price_tick_size = 10^price_tick_exponent * round_up_pow10(ura_quote/ura_base) =
	// 10^(price_tick_exponent + log10_round_up(ura_quote/ura_base)
	exponent := int64(priceTickExponent) + cbig.RatLog10RoundUp(cbig.NewRatFromBigInts(quoteURA, baseURA))

	return cbig.RatTenToThePower(exponent)
}
