package testing

import (
	"github.com/CoreumFoundation/coreum/pkg/types"
	"github.com/CoreumFoundation/coreum/tests/testing/rnd"
)

// RandomWallet generates wallet with random name and private key
func RandomWallet() types.Wallet {
	_, privKey := types.GenerateSecp256k1Key()
	return types.Wallet{Name: rnd.GetRandomName(), Key: privKey}
}

// MustCoin panics if `err` is not nil, returns `coin` otherwise
func MustCoin(coin types.Coin, err error) types.Coin {
	if err != nil {
		panic(err)
	}
	return coin
}
