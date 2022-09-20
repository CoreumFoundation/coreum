package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// Keeper interface exposes methods required by ante handler decorator of fee model
type Keeper interface {
	TrackGas(ctx sdk.Context, gas int64)
	GetMinGasPrice(ctx sdk.Context) sdk.DecCoin
}

// FeeDecorator will check if the gas price offered by transaction's fee is at least as large
// as the current minimum gas price required by the network and computd by our fee model.
// CONTRACT: Tx must implement FeeTx to use FeeDecorator
type FeeDecorator struct {
	keeper Keeper
}

// NewFeeDecorator creates ante decorator refusing transactions which does not offer minimum gas price
func NewFeeDecorator(keeper Keeper) FeeDecorator {
	return FeeDecorator{
		keeper: keeper,
	}
}

// AnteHandle handles transaction in ante decorator
func (fd FeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	if ctx.BlockHeight() == 0 {
		// Don't enforce fee model on genesis block
		return next(ctx, tx, simulate)
	}

	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	if err := fd.actOnFeeModelOutput(ctx, feeTx); err != nil {
		return ctx, err
	}

	fd.collectFeeModelInput(ctx, feeTx)

	return next(ctx, tx, simulate)
}

func (fd FeeDecorator) actOnFeeModelOutput(ctx sdk.Context, feeTx sdk.FeeTx) error {
	fees := feeTx.GetFee()
	minGasPrice := fd.keeper.GetMinGasPrice(ctx)
	if len(fees) == 0 {
		return sdkerrors.Wrap(sdkerrors.ErrInsufficientFee, "no fee declared for transaction")
	}
	if fees[0].Denom != minGasPrice.Denom {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "fee must be paid in '%s' coin only", minGasPrice.Denom)
	}

	gasDeclared := sdk.NewDecFromInt(sdk.NewIntFromUint64(feeTx.GetGas()))
	feeOffered := sdk.NewDecCoin(minGasPrice.Denom, fees.AmountOf(minGasPrice.Denom))
	feeRequired := sdk.NewDecCoinFromDec(minGasPrice.Denom, gasDeclared.Mul(minGasPrice.Amount))

	if feeOffered.IsLT(feeRequired) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeOffered, feeRequired)
	}
	return nil
}

func (fd FeeDecorator) collectFeeModelInput(ctx sdk.Context, feeTx sdk.FeeTx) {
	fd.keeper.TrackGas(ctx, int64(feeTx.GetGas()))
}
