package testing

import (
	"encoding/hex"
	"reflect"
	"sync"

	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// ChainContext is a types used to store the components required for the test chains subcomponents.
type ChainContext struct {
	ClientContext tx.ClientContext
	NetworkConfig config.NetworkConfig
	keyringMu     *sync.RWMutex
}

// NewChainContext returns a new instance if the ChainContext.
func NewChainContext(clientCtx tx.ClientContext, networkCfg config.NetworkConfig) ChainContext {
	return ChainContext{
		ClientContext: clientCtx,
		NetworkConfig: networkCfg,
		keyringMu:     &sync.RWMutex{},
	}
}

// RandomWallet generates a wallet for the chain with random name and
// private key and stores mnemonic in Keyring.
func (c ChainContext) RandomWallet() sdk.AccAddress {
	// Generate and store a new mnemonic using temporary keyring
	_, mnemonic, err := keyring.NewInMemory().NewMnemonic("tmp", keyring.English, "", "", hd.Secp256k1)
	if err != nil {
		panic(err)
	}

	// TODO(dhil) start returning the key info instead of address.
	return c.ImportMnemonic(mnemonic)
}

// ImportMnemonic imports the mnemonic into the clientContext Keyring and return its address.
func (c ChainContext) ImportMnemonic(mnemonic string) sdk.AccAddress {
	c.keyringMu.Lock()
	defer c.keyringMu.Unlock()
	keyInfo, err := c.ClientContext.Keyring().NewAccount(uuid.New().String(), mnemonic, "", "", hd.Secp256k1)
	if err != nil {
		panic(err)
	}

	return keyInfo.GetAddress()
}

// TxFactory returns factory with present values for the Chain.
func (c ChainContext) TxFactory() tx.Factory {
	return tx.Factory{}.
		WithKeybase(c.ClientContext.Keyring()).
		WithChainID(string(c.NetworkConfig.ChainID)).
		WithTxConfig(c.ClientContext.TxConfig()).
		WithGasPrices(c.NewDecCoin(c.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice).String())
}

// NewCoin helper function to initialize sdk.Coin by passing just amount.
func (c ChainContext) NewCoin(amount sdk.Int) sdk.Coin {
	return sdk.NewCoin(c.NetworkConfig.TokenSymbol, amount)
}

// NewDecCoin helper function to initialize sdk.DecCoin by passing just amount.
func (c ChainContext) NewDecCoin(amount sdk.Dec) sdk.DecCoin {
	return sdk.NewDecCoinFromDec(c.NetworkConfig.TokenSymbol, amount)
}

// GasLimitByMsgs calculates sum of gas limits required for message types passed.
// It panics if unsupported message type specified.
func (c ChainContext) GasLimitByMsgs(msgs ...sdk.Msg) uint64 {
	deterministicGas := c.NetworkConfig.Fee.DeterministicGas
	var totalGasRequired uint64
	for _, msg := range msgs {
		msgGas, exists := deterministicGas.GasRequiredByMessage(msg)
		if !exists {
			panic(errors.Errorf("unsuported message type for deterministic gas: %v", reflect.TypeOf(msg).String()))
		}
		totalGasRequired += msgGas
	}

	return totalGasRequired + deterministicGas.FixedGas
}

// AccAddressToLegacyWallet is temporary method to keep compatibility between
// func signatures while types.Wallet is being removed.
func (c ChainContext) AccAddressToLegacyWallet(accAddr sdk.AccAddress) types.Wallet {
	c.keyringMu.RLock()
	defer c.keyringMu.RUnlock()

	info, err := c.ClientContext.Keyring().KeyByAddress(accAddr)
	if err != nil {
		panic(err)
	}

	privKeyHex, err := keyring.NewUnsafe(c.ClientContext.Keyring()).UnsafeExportPrivKeyHex(info.GetName())
	if err != nil {
		panic(err)
	}

	privKeyBytes, err := hex.DecodeString(privKeyHex)
	if err != nil {
		panic(err)
	}

	return types.Wallet{Name: info.GetName(), Key: privKeyBytes}
}

// ChainConfig defines the config arguments required for the test chain initialisation.
type ChainConfig struct {
	RPCAddress      string
	NetworkConfig   config.NetworkConfig
	FundingMnemonic string
	StakerMnemonics []string
}

// Chain holds network and client for the blockchain
type Chain struct {
	ChainContext
	Client     client.Client
	Faucet     Faucet
	Governance Governance
}

// NewChain creates an instance of the new Chain.
func NewChain(cfg ChainConfig) Chain {
	coredClient := client.New(cfg.NetworkConfig.ChainID, cfg.RPCAddress)
	rpcClient, err := cosmosclient.NewClientFromNode(cfg.RPCAddress)
	if err != nil {
		panic(err)
	}
	clientContext := tx.NewClientContext(app.ModuleBasics).
		WithChainID(string(cfg.NetworkConfig.ChainID)).
		WithClient(rpcClient).
		WithKeyring(keyring.NewInMemory()).
		WithBroadcastMode(flags.BroadcastBlock)

	chainContext := NewChainContext(clientContext, cfg.NetworkConfig)
	governance := NewGovernance(chainContext, cfg.StakerMnemonics)
	faucet := NewFaucet(NewChainContext(clientContext.WithFromAddress(chainContext.ImportMnemonic(cfg.FundingMnemonic)), cfg.NetworkConfig))
	return Chain{
		ChainContext: chainContext,
		Client:       coredClient,
		Governance:   governance,
		Faucet:       faucet,
	}
}
