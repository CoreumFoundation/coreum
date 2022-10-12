package testing

import (
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/pkg/types"
)

// RandomWallet generates wallet with random name and private key
// Deprecated: Use chain.RandomWallet instead
func RandomWallet() types.Wallet {
	privKey := cosmossecp256k1.GenPrivKey()
	return types.Wallet{Name: sdk.AccAddress(privKey.PubKey().Address()).String(), Key: *privKey}
}

// ComputeNeededBalance computes the required balance for sending `numOfTransactions` number of transactions plus some extra amount.
// FIXME (wojtek): hardcode reasonable default values: https://reviewable.io/reviews/CoreumFoundation/coreum/131#-NA4cljcBl9TBFEqA81t
func ComputeNeededBalance(gasPrice sdk.Dec, transactionGasLimit uint64, numOfTransactions int, extraAmount sdk.Int) sdk.Int {
	// Ceil().RoundInt() is here to be compatible with the sdk's TxFactory
	// https://github.com/cosmos/cosmos-sdk/blob/ff416ee63d32da5d520a8b2d16b00da762416146/client/tx/factory.go#L223
	return gasPrice.Mul(sdk.NewIntFromUint64(transactionGasLimit).ToDec()).Ceil().RoundInt().MulRaw(int64(numOfTransactions)).Add(extraAmount)
}

// MustNewIntFromString returns a new instance of sdk.Int type from string and fails the test in case of error.
func MustNewIntFromString(t T, v string) sdk.Int {
	i, ok := sdk.NewIntFromString(v)
	require.True(t, ok, "creating sdk.Int from string failed")
	return i
}
