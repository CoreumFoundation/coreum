package testing

import (
	"context"
	"encoding/hex"
	"reflect"
	"sync"

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

// ChainContext is a types used to store the components required for the test chains subcomponents.
type ChainContext struct {
	ClientContext cosmosclient.Context
	NetworkConfig app.NetworkConfig
	mu            sync.Mutex
}

// NewChainContext returns a new instance if the ChainContext.
func NewChainContext(clientCtx cosmosclient.Context, networkCfg app.NetworkConfig) *ChainContext {
	return &ChainContext{
		ClientContext: clientCtx,
		NetworkConfig: networkCfg,
		mu:            sync.Mutex{},
	}
}

// RandomWallet generates a wallet for the chain with random name and
// private key and stores mnemonic in Keyring.
func (c *ChainContext) RandomWallet() sdk.AccAddress {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Generate and store a new mnemonic using temporary keyring
	keyInfo, mnemonic, err := keyring.NewInMemory().NewMnemonic("tmp", keyring.English, "", "", hd.Secp256k1)
	// we are using panics here, since we are sure it will not error out, and handling error
	// upstream is a waste of time.
	if err != nil {
		panic(err)
	}

	// Store generated mnemonic using account address as UID
	if _, err = c.ClientContext.Keyring.NewAccount(keyInfo.GetAddress().String(), mnemonic, "", "", hd.Secp256k1); err != nil {
		panic(err)
	}

	return keyInfo.GetAddress()
}

// TxFactory returns factory with present values for the Chain.
func (c *ChainContext) TxFactory() tx.Factory {
	return tx.Factory{}.
		WithKeybase(c.ClientContext.Keyring).
		WithChainID(string(c.NetworkConfig.ChainID)).
		WithTxConfig(c.ClientContext.TxConfig).
		WithGasPrices(c.NewDecCoin(c.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice).String())
}

// NewCoin helper function to initialize sdk.Coin by passing just amount.
func (c *ChainContext) NewCoin(amount sdk.Int) sdk.Coin {
	return sdk.NewCoin(c.NetworkConfig.TokenSymbol, amount)
}

// NewDecCoin helper function to initialize sdk.DecCoin by passing just amount.
func (c *ChainContext) NewDecCoin(amount sdk.Dec) sdk.DecCoin {
	return sdk.NewDecCoinFromDec(c.NetworkConfig.TokenSymbol, amount)
}

// GasLimitByMsgs calculates sum of gas limits required for message types passed.
// It panics if unsupported message type specified.
func (c *ChainContext) GasLimitByMsgs(msgs ...sdk.Msg) uint64 {
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
func (c *ChainContext) AccAddressToLegacyWallet(accAddr sdk.AccAddress) types.Wallet {
	name := accAddr.String()
	privKeyHex, err := keyring.NewUnsafe(c.ClientContext.Keyring).UnsafeExportPrivKeyHex(name)
	if err != nil {
		panic(err)
	}

	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		panic(err)
	}

	return types.Wallet{Name: name, Key: privKeyBytes}
}

// ChainConfig defines the config arguments required for the test chain initialisation.
type ChainConfig struct {
	RPCAddress     string
	NetworkConfig  app.NetworkConfig
	FundingPrivKey types.Secp256k1PrivateKey
}

// Chain holds network and client for the blockchain
type Chain struct {
	*ChainContext
	Client     client.Client
	Faucet     *Faucet
	Governance *Governance
}

// NewChain creates an instance of the new Chain.
func NewChain(ctx context.Context, cfg ChainConfig) (*Chain, error) {
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
		WithKeyring(keyring.NewInMemory()).
		WithBroadcastMode(flags.BroadcastBlock)

	chainContext := NewChainContext(clientContext, cfg.NetworkConfig)
	faucet := NewFaucet(coredClient, cfg.NetworkConfig, cfg.FundingPrivKey)
	governance, err := NewGovernance(ctx, chainContext, faucet)
	if err != nil {
		return nil, errors.Wrap(err, "can't init chain governance")
	}

	return &Chain{
		ChainContext: chainContext,
		Client:       coredClient,
		Faucet:       faucet,
		Governance:   governance,
	}, nil
}
