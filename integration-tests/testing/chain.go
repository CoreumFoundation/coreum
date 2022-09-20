package testing

import (
	"context"
	"encoding/hex"
	"reflect"

	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

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
	Faucet        *Faucet

	Keyring keyring.Keyring
}

type ChainConfig struct {
	RPCAddress     string
	NetworkConfig  app.NetworkConfig
	FundingPrivKey types.Secp256k1PrivateKey
}

func NewChain(ctx context.Context, cfg ChainConfig) (Chain, error) {
	//nolint:contextcheck // `New->New->NewWithClient->New$1` should pass the context parameter
	coredClient := client.New(cfg.NetworkConfig.ChainID, cfg.RPCAddress)
	//nolint:contextcheck // `New->NewWithClient` should pass the context parameter
	rpcClient, err := cosmosclient.NewClientFromNode(cfg.RPCAddress)
	if err != nil {
		panic(err)
	}
	clientContext := app.
		NewDefaultClientContext().
		WithChainID(string(cfg.NetworkConfig.ChainID)).
		WithClient(rpcClient).
		WithBroadcastMode(flags.BroadcastBlock)

	faucet := NewFaucet(coredClient, cfg.NetworkConfig, cfg.FundingPrivKey)

	return Chain{
		Client:        coredClient,
		ClientContext: clientContext,
		NetworkConfig: cfg.NetworkConfig,
		Faucet:        faucet,
		Keyring:       keyring.NewInMemory(),
	}, nil
}

// RandomWallet generates a wallet for the chain with random name and
// private key and stores mnemonic in Keyring.
func (c Chain) RandomWallet() sdk.AccAddress {
	// Generate and store a new mnemonic using temporary keyring
	keyInfo, mnemonic, err := keyring.NewInMemory().NewMnemonic("tmp", keyring.English, "", "", hd.Secp256k1)
	// we are using panics here, since we are sure it will not error out, and handling error
	// upstream is a waste of time.
	if err != nil {
		panic(err)
	}

	// Store generated mnemonic using account address as UID
	if _, err = c.Keyring.NewAccount(keyInfo.GetAddress().String(), mnemonic, "", "", hd.Secp256k1); err != nil {
		panic(err)
	}

	return keyInfo.GetAddress()
}

// TxFactory returns factory with present values for the Chain.
func (c Chain) TxFactory() tx.Factory {
	return tx.Factory{}.
		WithKeybase(c.Keyring).
		WithChainID(string(c.NetworkConfig.ChainID)).
		WithTxConfig(c.ClientContext.TxConfig).
		WithGasPrices(c.NewDecCoin(c.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice).String())
}

// NewCoin helper function to initialize sdk.Coin by passing just amount.
func (c Chain) NewCoin(amount sdk.Int) sdk.Coin {
	return sdk.NewCoin(c.NetworkConfig.TokenSymbol, amount)
}

// NewDecCoin helper function to initialize sdk.DecCoin by passing just amount.
func (c Chain) NewDecCoin(amount sdk.Dec) sdk.DecCoin {
	return sdk.NewDecCoinFromDec(c.NetworkConfig.TokenSymbol, amount)
}

// GasLimitByMsgs calculates sum of gas limits required for message types passed.
// It panics if unsupported message type specified.
func (c Chain) GasLimitByMsgs(msgs ...sdk.Msg) uint64 {
	var totalGasRequired uint64 = 0
	for _, msg := range msgs {
		msgGas := c.NetworkConfig.Fee.DeterministicGas.GasRequiredByMessage(msg)
		if msgGas == 0 {
			panic(errors.Errorf("unsuported message type for deterministic gas: %v", reflect.TypeOf(msg).String()))
		}
		totalGasRequired += msgGas
	}

	return totalGasRequired
}

// AccAddressToLegacyWallet is temporary method to keep compatibility between
// func signatures while types.Wallet is being removed.
func (c Chain) AccAddressToLegacyWallet(accAddr sdk.AccAddress) types.Wallet {
	name := accAddr.String()
	privKeyHex, err := keyring.NewUnsafe(c.Keyring).UnsafeExportPrivKeyHex(name)
	if err != nil {
		panic(err)
	}

	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		panic(err)
	}

	return types.Wallet{Name: name, Key: privKeyBytes}
}
