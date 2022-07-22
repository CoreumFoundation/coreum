package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// GasPriceKeeper interface returns minimum gas price required by the network
type GasPriceKeeper interface {
	GetMinGasPrice(ctx sdk.Context) sdk.Coin
}

// FeeDecorator will check if the gas price offered by transaction's fee is at least as large
// as the current minimum gas price required by the network and computd by our fee model.
// CONTRACT: Tx must implement FeeTx to use FeeDecorator
type FeeDecorator struct {
	gasPriceKeeper GasPriceKeeper
}

// NewFeeDecorator creates ante decorator refusing transactions which does not offer minimum gas price
func NewFeeDecorator(gasPriceKeeper GasPriceKeeper) FeeDecorator {
	return FeeDecorator{
		gasPriceKeeper: gasPriceKeeper,
	}
}

// AnteHandle handles transaction in ante decorator
func (fd FeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if ctx.BlockHeight() == 0 {
		// Don't enforce fee model on genesis block
		return next(ctx, tx, simulate)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	fees := feeTx.GetFee()
	minGasPrice := fd.gasPriceKeeper.GetMinGasPrice(ctx)
	for _, coin := range fees {
		if coin.GetDenom() != minGasPrice.Denom {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "fee must be paid in '%s' coin, but '%s' was offered instead", minGasPrice.Denom, coin.Denom)
		}
	}

	gasDeclared := sdk.NewInt(int64(feeTx.GetGas()))
	feeOffered := sdk.NewCoin(minGasPrice.Denom, fees.AmountOf(minGasPrice.Denom))
	feeRequired := sdk.NewCoin(minGasPrice.Denom, gasDeclared.Mul(minGasPrice.Amount))

	if feeOffered.IsLT(feeRequired) {
		return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeOffered, feeRequired)
	}

	return next(ctx, tx, simulate)
}

// GasCollectorKeeper interface collects gas used by transactions in the block
type GasCollectorKeeper interface {
	TrackGas(ctx sdk.Context, gas int64)
}

// CollectGasDecorator collects gas used by all the transactions in the block.
// CONTRACT: Tx must implement FeeTx to use FeeDecorator
type CollectGasDecorator struct {
	gasCollectorKeeper GasCollectorKeeper
}

// NewCollectGasDecorator creates ante decorator collecting gas used by trransactions in the block
func NewCollectGasDecorator(gasCollectorKeeper GasCollectorKeeper) CollectGasDecorator {
	return CollectGasDecorator{
		gasCollectorKeeper: gasCollectorKeeper,
	}
}

// AnteHandle handles transaction in ante decorator
func (cgd CollectGasDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	cgd.gasCollectorKeeper.TrackGas(ctx, int64(feeTx.GetGas()))

	return next(ctx, tx, simulate)
}
