package keeper

import (
	"errors"
	"math/big"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) getPriceTick(ctx sdk.Context, baseDenom, quoteDenom string) (*big.Rat, error) {
	params, err := k.GetParams(ctx)
	if err != nil {
		return nil, err
	}

	baseDenomRefAmount, err := k.getAssetFTUnifiedRefAmount(ctx, baseDenom, params.DefaultUnifiedRefAmount)
	if err != nil {
		return nil, err
	}

	quoteDenomRefAmount, err := k.getAssetFTUnifiedRefAmount(ctx, quoteDenom, params.DefaultUnifiedRefAmount)
	if err != nil {
		return nil, err
	}

	return ComputePriceTick(baseDenomRefAmount, quoteDenomRefAmount, params.PriceTickExponent), nil
}

func (k Keeper) validatePriceTick(
	ctx sdk.Context,
	params types.Params,
	baseDenom, quoteDenom string,
	price types.Price,
) error {
	priceTickRat, err := k.getPriceTick(ctx, baseDenom, quoteDenom)
	if err != nil {
		return err
	}

	_, remainder := cbig.RatQuoWithIntRemainder(price.Rat(), priceTickRat)
	if !cbig.IntEqZero(remainder) {
		return sdkerrors.Wrapf(
			types.ErrInvalidInput,
			"invalid price %s, the price must be multiple of %s",
			price.Rat().String(), priceTickRat.String(),
		)
	}

	return nil
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

// ComputePriceTick returns the price tick of a given ref amounts and price tick exponent.
func ComputePriceTick(baseDenomRefAmount, quoteRefAmount sdkmath.LegacyDec, priceTickExponent int32) *big.Rat {
	// 10^(floor(log10((quoteRefAmountRat / baseRefAmountRat))) + price_tick_exponent)
	exponent := ratFloorLog10(
		cbig.NewRatFromBigInts(quoteRefAmount.BigInt(), baseDenomRefAmount.BigInt()),
	) + int(priceTickExponent)
	if exponent < 0 {
		return cbig.NewRatFromBigInts(big.NewInt(1), cbig.IntTenToThePower(big.NewInt(int64(-exponent))))
	}

	return cbig.NewRatFromBigInt(cbig.IntTenToThePower(big.NewInt(int64(exponent))))
}
