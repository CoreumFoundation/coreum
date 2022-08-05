package testing

import (
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// RandomWallet generates wallet with random name and private key
func RandomWallet() types.Wallet {
	_, privKey := types.GenerateSecp256k1Key()
	return types.Wallet{Name: privKey.Address(), Key: privKey}
}
