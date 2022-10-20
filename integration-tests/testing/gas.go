package testing

import sdk "github.com/cosmos/cosmos-sdk/types"

// ComputeFeeAmount computes the fee amount based on the current gas price and limit.
func ComputeFeeAmount(gasPrice sdk.Dec, gasLimit uint64) sdk.Int {
	// Ceil().RoundInt() is here to be compatible with the sdk's TxFactory
	// https://github.com/cosmos/cosmos-sdk/blob/ff416ee63d32da5d520a8b2d16b00da762416146/client/tx/factory.go#L223
	return gasPrice.Mul(sdk.NewIntFromUint64(gasLimit).ToDec()).Ceil().RoundInt()
}
