package testing

import (
	"encoding/hex"

	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// Chain holds network and client for the blockchain
type Chain struct {
	Client        client.Client
	ClientContext cosmosclient.Context

	NetworkConfig app.NetworkConfig
	Faucet        Faucet

	Keyring keyring.UnsafeKeyring
}

// RandomWallet generates a wallet for the chain with random name and
// private key and stores mnemonic in Keyring.
func (c Chain) RandomWallet() types.Wallet {
	name := uuid.New().String()
	_, _, err := c.Keyring.NewMnemonic(name, keyring.English, "", "", hd.Secp256k1)
	if err != nil {
		// we are using panic here, since we are sure it will not error out, and handling error
		// upstream is a waste of time.
		panic(err)
	}
	privKeyHex, err := c.Keyring.UnsafeExportPrivKeyHex(name)
	if err != nil {
		panic(err)
	}

	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		panic(err)
	}

	return types.Wallet{Name: name, Key: privKeyBytes}
}

// TxFactory returns factory with present values for the Chain.
func (c Chain) TxFactory() tx.Factory {
	return tx.Factory{}.
		WithKeybase(c.Keyring).
		WithChainID(string(c.NetworkConfig.ChainID)).
		WithTxConfig(c.ClientContext.TxConfig).
		WithGasPrices(sdk.NewCoin(c.NetworkConfig.TokenSymbol, c.NetworkConfig.Fee.FeeModel.InitialGasPrice).String())
}
