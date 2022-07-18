// This content was copied and modified based on github.com/cosmos/cosmos-sdk/x/auth/ante/fee.go
// Original content: https://github.com/cosmos/cosmos-sdk/blob/ad9e5620fb3445c716e9de45cfcdb56e8f1745bf/x/auth/ante/fee.go

package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// MempoolFeeDecorator will check if the transaction's fee is at least as large
// as the local validator's minimum gasFee (defined in validator config).
// If fee is too low, decorator returns error and tx is rejected from mempool.
// Note this only applies when ctx.CheckTx = true
// If fee is high enough or not CheckTx, then call next AnteHandler
// CONTRACT: Tx must implement FeeTx to use MempoolFeeDecorator
type MempoolFeeDecorator struct {
	minGasPrice sdk.Coin
}

// NewMempoolFeeDecorator creates ante decorator refusing transactions which does not offer minimum gas price
func NewMempoolFeeDecorator(minGasPrice sdk.Coin) MempoolFeeDecorator {
	return MempoolFeeDecorator{
		minGasPrice: minGasPrice,
	}
}

// AnteHandle handles transaction in ante decorator
func (mfd MempoolFeeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return ctx, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	fees := feeTx.GetFee()
	for _, coin := range fees {
		if coin.GetDenom() != mfd.minGasPrice.Denom {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInvalidCoins, "fee must be paid in '%s' coin, but '%s' was offered instead", mfd.minGasPrice.Denom, coin.Denom)
		}
	}

	// Ensure that the provided fees meet a minimum threshold for the validator,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() && !simulate {
		gasDeclared := sdk.NewInt(int64(feeTx.GetGas()))
		feeOffered := sdk.NewCoin(mfd.minGasPrice.Denom, fees.AmountOf(mfd.minGasPrice.Denom))
		feeRequired := sdk.NewCoin(mfd.minGasPrice.Denom, gasDeclared.Mul(mfd.minGasPrice.Amount))

		if feeOffered.IsLT(feeRequired) {
			return ctx, sdkerrors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeOffered, feeRequired)
		}
	}

	return next(ctx, tx, simulate)
}
