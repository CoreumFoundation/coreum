package types

import (
	"math/big"

	"github.com/pkg/errors"
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

// String returns string representation of the wallet
func (w Wallet) String() string {
	return w.Name + "@" + w.Key.Address()
}

// NewCoin returns a new instance of coin type
func NewCoin(amount *big.Int, denom string) (Coin, error) {
	c := Coin{
		Amt: amount,
		Dnm: denom,
	}
	if c.Dnm == "" {
		return Coin{}, errors.New("denom is empty")
	}
	if c.Amt == nil {
		return Coin{}, errors.New("amount is nil")
	}
	if c.Amt.Cmp(big.NewInt(0)) == -1 {
		return Coin{}, errors.New("amount is negative")
	}

	return c, nil
}

// Coin stores amount and denom of token
type Coin struct {
	// Amount is stored amount
	Amt *big.Int `json:"amount"`

	// Denom is a token symbol
	Dnm string `json:"denom"`
}

// String returns string representation of coin
func (c Coin) String() string {
	return c.Amt.String() + c.Dnm
}
