package client2

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TxConfig holds values common to every transaction
type TxConfig struct {
	From        sdk.AccAddress
	FromAccount *accountInfo
	Keyring     keyring.Keyring
	GasLimit    uint64
	GasPrice    *sdk.Coin
	Memo        string
}

// SetAccountNumber allows to set account number explicitly
func (c *TxConfig) SetAccountNumber(n uint64) {
	if c.FromAccount == nil {
		c.FromAccount = &accountInfo{}
	}

	c.FromAccount.Number = n
}

// SetAccountSequence allows to set account sequence explicitly
func (c *TxConfig) SetAccountSequence(seq uint64) {
	if c.FromAccount == nil {
		c.FromAccount = &accountInfo{}
	}

	c.FromAccount.Sequence = seq
}

// accountInfo stores account number and sequence
type accountInfo struct {
	Number   uint64
	Sequence uint64
}
