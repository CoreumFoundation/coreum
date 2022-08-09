package testing

import (
	"math/big"

	"github.com/CoreumFoundation/coreum/pkg/types"
)

// RandomWallet generates wallet with random name and private key
func RandomWallet() types.Wallet {
	_, privKey := types.GenerateSecp256k1Key()
	return types.Wallet{Name: privKey.Address(), Key: privKey}
}

// ComputeNeededBalance computes the required balance for sending `numOfMessages` number of messages plus some extra amount.
func ComputeNeededBalance(gasPrice *big.Int, messageGasLimit uint64, numOfMessages int, extraAmount *big.Int) *big.Int {
	balance := new(big.Int).Mul(gasPrice, big.NewInt(int64(messageGasLimit)))
	balance.Mul(balance, big.NewInt(int64(numOfMessages)))
	return balance.Add(balance, extraAmount)
}
