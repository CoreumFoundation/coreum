package tx

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SignInput contains input field to the sign function
type SignInput struct {
	PrivateKey  secp256k1.PrivKey
	AccountInfo AccountInfo
	GasLimit    uint64
	GasPrice    sdk.Coin
	Memo        string
}

// AccountInfo contains account number and sequence used to sign a tx
type AccountInfo struct {
	// Number is the account number as stored on blockchain
	Number uint64

	// Sequence is the sequence of next transaction to sign
	Sequence uint64
}
