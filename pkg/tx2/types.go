package tx2

import "github.com/CoreumFoundation/coreum/pkg/types"

// AccountInfo stores account number and sequence
type AccountInfo struct {
	Number   uint64
	Sequence uint64
}

// Signer stores information about account which signs the transaction
// Common scenarios:
// - for broadcasting transaction both public and private keys must be set, if Account is nil client will query blockchain
//   for correct account number and sequence, if it is set - provided values are used blindly,
// - for estimating gas used by transaction, set only public key,
// - if transaction is added to genesis block set both keys and Account,
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
