package tx

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/pkg/types"
)

// BaseInput holds input data common to every transaction
type BaseInput struct {
	Signer   types.Wallet
	GasLimit uint64
	GasPrice sdk.DecCoin
	Memo     string
}
