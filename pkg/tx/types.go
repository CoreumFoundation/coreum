package tx

import "github.com/CoreumFoundation/coreum/pkg/types"

// BaseInput holds input data common to every transaction
type BaseInput struct {
	Signer   types.Wallet
	GasLimit uint64
	GasPrice types.Coin
	Memo     string
}
