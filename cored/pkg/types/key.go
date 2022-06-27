package types

import (
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Secp256k1PrivateKey is a secp256k1 private key
type Secp256k1PrivateKey []byte

// PubKey returns public key for corresponding key
func (key Secp256k1PrivateKey) PubKey() Secp256k1PublicKey {
	privKey := cosmossecp256k1.PrivKey{Key: key}
	return privKey.PubKey().Bytes()
}

// Address returns bech32 encoded wallet address for corresponding key
func (key Secp256k1PrivateKey) Address() string {
	privKey := cosmossecp256k1.PrivKey{Key: key}
	return sdk.AccAddress(privKey.PubKey().Address()).String()
}

// Secp256k1PublicKey is a secp256k1 public key
type Secp256k1PublicKey []byte

// GenerateSecp256k1Key generates random secp256k1 key pair
func GenerateSecp256k1Key() (Secp256k1PublicKey, Secp256k1PrivateKey) {
	privKey := cosmossecp256k1.GenPrivKey()
	return privKey.PubKey().Bytes(), privKey.Bytes()
}
