package types

import (
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

// Coin stores amount and denom of token
type Coin struct {
	// Amount is stored amount
	Amount Int `json:"amount"`

	// Denom is a token symbol
	Denom string `json:"denom"`
}

// NewCoin returns a new instance of coin type
func NewCoin(amount Int, denom string) (Coin, error) {
	c := Coin{
		Amount: amount,
		Denom:  denom,
	}
	if err := c.Validate(); err != nil {
		return Coin{}, err
	}

	return c, nil
}

// Validate validates data inside coin
func (c Coin) Validate() error {
	if c.Denom == "" {
		return errors.New("denom is empty")
	}
	if c.Amount.LT(NewInt(0)) {
		return errors.New("amount is negative")
	}
	return nil
}

// String returns string representation of coin
func (c Coin) String() string {
	return c.Amount.String() + c.Denom
}
