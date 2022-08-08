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

// ComputeInitialBalance computes initial balance by multiplying gas price by deterministic amount of gas required by single transaction and number of expected transactions
// plus additional balance for transfers
func ComputeInitialBalance(gasPrice *big.Int, messageGasLimit uint64, numOfMessages int, freeAmount *big.Int) *big.Int {
	initialBalance := new(big.Int).Mul(gasPrice, big.NewInt(int64(messageGasLimit)))
	initialBalance.Mul(initialBalance, big.NewInt(int64(numOfMessages)))
	return initialBalance.Add(initialBalance, freeAmount)
}
