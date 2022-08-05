package types

import (
	"math/big"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	cosmcrypto "github.com/cosmos/cosmos-sdk/crypto"
	cosmkeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

// Address returns cosmos acc address from the pub key of the wallet.
func (w Wallet) Address() sdk.AccAddress {
	privKey := cosmossecp256k1.PrivKey{Key: w.Key}
	return sdk.AccAddress(privKey.PubKey().Address())
}

// NewWalletFromKeyring allows to wrap an account key from keyring into an unsafe Wallet wrapper.
func NewWalletFromKeyring(kb cosmkeyring.Keyring, accAddr sdk.AccAddress) (wallet Wallet, err error) {
	keyInfo, err := kb.KeyByAddress(accAddr)
	if err != nil {
		err = errors.Wrapf(err, "failed to locate key by address %s in the keyring", accAddr.String())
		return wallet, err
	}

	armor, err := kb.ExportPrivKeyArmorByAddress(accAddr, "")
	must.OK(err)

	privKey, _, err := cosmcrypto.UnarmorDecryptPrivKey(armor, "")
	must.OK(err)

	wallet = Wallet{
		Name:            keyInfo.GetName(),
		Key:             Secp256k1PrivateKey(privKey.(*cosmossecp256k1.PrivKey).Key),
		AccountNumber:   0,
		AccountSequence: 0,
	}

	return wallet, nil
}

// Coin stores amount and denom of token
type Coin struct {
	// Amount is stored amount
	Amount *big.Int `json:"amount"`

	// Denom is a token symbol
	Denom string `json:"denom"`
}

// NewCoin returns a new instance of coin type
func NewCoin(amount *big.Int, denom string) (Coin, error) {
	c := Coin{
		Amount: big.NewInt(0).Set(amount),
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
	if c.Amount == nil {
		return errors.New("amount is nil")
	}
	if c.Amount.Cmp(big.NewInt(0)) == -1 {
		return errors.New("amount is negative")
	}
	return nil
}

// String returns string representation of coin
func (c Coin) String() string {
	return c.Amount.String() + c.Denom
}
