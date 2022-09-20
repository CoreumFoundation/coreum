package types

import (
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Wallet stores information related to wallet
// TODO: Drop this type
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

// Address returns cosmos acc address from the pub key of the wallet.
func (w Wallet) Address() sdk.AccAddress {
	privKey := cosmossecp256k1.PrivKey{Key: w.Key}
	return sdk.AccAddress(privKey.PubKey().Address())
}
