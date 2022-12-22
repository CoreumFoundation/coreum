package integrationtests

import (
	"reflect"

	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// ChainContext is a types used to store the components required for the test chains subcomponents.
type ChainContext struct {
	ClientContext tx.ClientContext
	NetworkConfig config.NetworkConfig
}

// NewChainContext returns a new instance if the ChainContext.
func NewChainContext(clientCtx tx.ClientContext, networkCfg config.NetworkConfig) ChainContext {
	return ChainContext{
		ClientContext: clientCtx,
		NetworkConfig: networkCfg,
	}
}

// GenAccount generates a new account for the chain with random name and
// private key and stores it in the chains ClientContext Keyring.
func (c ChainContext) GenAccount() sdk.AccAddress {
	// Generate and store a new mnemonic using temporary keyring
	_, mnemonic, err := keyring.NewInMemory().NewMnemonic(
		"tmp",
		keyring.English,
		sdk.GetConfig().GetFullBIP44Path(),
		"",
		hd.Secp256k1,
	)
	if err != nil {
		panic(err)
	}

	return c.ImportMnemonic(mnemonic)
}

// ImportMnemonic imports the mnemonic into the ClientContext Keyring and return its address.
// If the mnemonic is already imported the method will just return the address.
func (c ChainContext) ImportMnemonic(mnemonic string) sdk.AccAddress {
	keyInfo, err := c.ClientContext.Keyring().NewAccount(
		uuid.New().String(),
		mnemonic,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
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
	return sdk.NewCoin(c.NetworkConfig.Denom, amount)
}

// NewDecCoin helper function to initialize sdk.DecCoin by passing just amount.
func (c ChainContext) NewDecCoin(amount sdk.Dec) sdk.DecCoin {
	return sdk.NewDecCoinFromDec(c.NetworkConfig.Denom, amount)
}

// DeterministicGas returns deterministic gas config
func (c ChainContext) DeterministicGas() config.DeterministicGasRequirements {
	return c.NetworkConfig.Fee.DeterministicGas
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
		totalGasRequired += msgGas + deterministicGas.FixedGas
	}

	return totalGasRequired
}

// GasLimitByMultiSendMsgs calculates sum of gas limits required for message types passed and includes the FixedGas once.
// It panics if unsupported message type specified.
func (c ChainContext) GasLimitByMultiSendMsgs(msgs ...sdk.Msg) uint64 {
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

// BalancesOptions is the input type for the ComputeNeededBalanceFromOptions.
type BalancesOptions struct {
	Messages []sdk.Msg
	GasPrice sdk.Dec
	Amount   sdk.Int
}

// ComputeNeededBalanceFromOptions computes the required balance based on the input options.
func (c ChainContext) ComputeNeededBalanceFromOptions(options BalancesOptions) sdk.Int {
	if options.GasPrice.IsNil() {
		options.GasPrice = c.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice
	}

	if options.Amount.IsNil() {
		options.Amount = sdk.ZeroInt()
	}

	// NOTE: we assume that each message goes to one transaction, which is not
	// very accurate and may cause some over funding in cases that there are multiple
	// messages in a single transaction
	totalAmount := sdk.ZeroInt()
	for _, msg := range options.Messages {
		gas := c.GasLimitByMsgs(msg)
		// Ceil().RoundInt() is here to be compatible with the sdk's TxFactory
		// https://github.com/cosmos/cosmos-sdk/blob/ff416ee63d32da5d520a8b2d16b00da762416146/client/tx/factory.go#L223
		amt := options.GasPrice.Mul(sdk.NewDec(int64(gas))).Ceil().RoundInt()
		totalAmount = totalAmount.Add(amt)
	}

	return totalAmount.Add(options.Amount)
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
	Faucet     Faucet
	Governance Governance
}

// NewChain creates an instance of the new Chain.
func NewChain(cfg ChainConfig) Chain {
	rpcClient, err := cosmosclient.NewClientFromNode(cfg.RPCAddress)
	if err != nil {
		panic(err)
	}
	clientCtx := tx.NewClientContext(app.ModuleBasics).
		WithChainID(string(cfg.NetworkConfig.ChainID)).
		WithClient(rpcClient).
		WithKeyring(newConcurrentSafeKeyring(keyring.NewInMemory())).
		WithBroadcastMode(flags.BroadcastBlock)

	chainCtx := NewChainContext(clientCtx, cfg.NetworkConfig)
	governance := NewGovernance(chainCtx, cfg.StakerMnemonics)

	faucetAddr := chainCtx.ImportMnemonic(cfg.FundingMnemonic)
	faucet := NewFaucet(NewChainContext(clientCtx.WithFromAddress(faucetAddr), cfg.NetworkConfig))
	return Chain{
		ChainContext: chainCtx,
		Governance:   governance,
		Faucet:       faucet,
	}
}
