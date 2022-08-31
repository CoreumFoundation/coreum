package testing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/pkg/types"
)

// RandomWallet generates wallet with random name and private key
func RandomWallet() types.Wallet {
	_, privKey := types.GenerateSecp256k1Key()
	return types.Wallet{Name: privKey.Address(), Key: privKey}
}

// ComputeNeededBalance computes the required balance for sending `numOfMessages` number of messages plus some extra amount.
// FIXME (wojtek): hardcode reasonable default values: https://reviewable.io/reviews/CoreumFoundation/coreum/131#-NA4cljcBl9TBFEqA81t
func ComputeNeededBalance(gasPrice sdk.Int, messageGasLimit uint64, numOfMessages int, extraAmount sdk.Int) sdk.Int {
	return gasPrice.MulRaw(int64(messageGasLimit)).MulRaw(int64(numOfMessages)).Add(extraAmount)
}

// MustNewCoin returns a new instance of coin type and fails the test in case of the validation error.
func MustNewCoin(t T, amount sdk.Int, denom string) types.Coin {
	c, err := types.NewCoin(amount.BigInt(), denom)
	require.NoError(t, err)
	return c
}

// MustNewIntFromString returns a new instance of sdk.Int type from string and fails the test in case of error.
func MustNewIntFromString(t T, v string) sdk.Int {
	i, ok := sdk.NewIntFromString(v)
	require.True(t, ok, "creating sdk.Int from string failed")
	return i
}
