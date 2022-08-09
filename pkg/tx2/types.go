package tx2

import "github.com/CoreumFoundation/coreum/pkg/types"

// AccountInfo stores account number and sequence
type AccountInfo struct {
	Number   uint64
	Sequence uint64
}

// Signer stores information about account which signs the transaction
type Signer struct {
	PublicKey  types.Secp256k1PublicKey
	PrivateKey types.Secp256k1PrivateKey
	Account    *AccountInfo
}

// Address returns address coresponding to the pricate key of the signer
func (s Signer) Address() types.Address {
	return s.PublicKey.Address()
}

// BaseInput holds input data common to every transaction
type BaseInput struct {
	Signer   Signer
	GasLimit uint64
	GasPrice types.Coin
	Memo     string
}
