package testing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ComputeNeededBalance computes the required balance for sending `numOfTransactions` number of transactions plus some extra amount.
// FIXME (wojtek): hardcode reasonable default values: https://reviewable.io/reviews/CoreumFoundation/coreum/131#-NA4cljcBl9TBFEqA81t
func ComputeNeededBalance(gasPrice sdk.Dec, transactionGasLimit uint64, numOfTransactions int, extraAmount sdk.Int) sdk.Int {
	// Ceil().RoundInt() is here to be compatible with the sdk's TxFactory
	// https://github.com/cosmos/cosmos-sdk/blob/ff416ee63d32da5d520a8b2d16b00da762416146/client/tx/factory.go#L223
	return gasPrice.Mul(sdk.NewIntFromUint64(transactionGasLimit).ToDec()).Ceil().RoundInt().MulRaw(int64(numOfTransactions)).Add(extraAmount)
}
