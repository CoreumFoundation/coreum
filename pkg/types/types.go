package types

import (
	"math/big"
)

// Wallet stores information related to wallet
type Wallet struct {
	// Name is the name of the key stored in keystore
	Name string

	// Key is the private key of the wallet
	Key Secp256k1PrivateKey

	// AccountNumber is the account number as stored on blockchain
	AccountNumber uint64

	// AccountSequence is the sequence of next transaction to sign
	AccountSequence uint64
}

func (w Wallet) String() string {
	return w.Name + "@" + w.Key.Address()
}

// Balance stores balance of denom
type Balance struct {
	// Amount is stored amount
	Amount *big.Int `json:"amount"`

	// Denom is a token symbol
	Denom string `json:"denom"`
}

func (b Balance) String() string {
	return b.Amount.String() + b.Denom
}
