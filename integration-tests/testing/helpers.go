package testing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

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
