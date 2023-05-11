package integrationtests

import (
	"reflect"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdkmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/x/deterministicgas"
)

// ChainContext is a types used to store the components required for the test chains subcomponents.
type ChainContext struct {
	ClientContext          client.Context
	NetworkConfig          config.NetworkConfig
	InitialGasPrice        sdk.Dec
	DeterministicGasConfig deterministicgas.Config
}

// NewChainContext returns a new instance if the ChainContext.
func NewChainContext(
	clientCtx client.Context,
	networkCfg config.NetworkConfig,
	initialGasPrice sdk.Dec,
) ChainContext {
	return ChainContext{
		ClientContext:          clientCtx,
		NetworkConfig:          networkCfg,
		InitialGasPrice:        initialGasPrice,
		DeterministicGasConfig: deterministicgas.DefaultConfig(),
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
func (c ChainContext) TxFactory() client.Factory {
	return client.Factory{}.
		WithKeybase(c.ClientContext.Keyring()).
		WithChainID(string(c.NetworkConfig.ChainID())).
		WithTxConfig(c.ClientContext.TxConfig()).
		WithGasPrices(c.NewDecCoin(c.InitialGasPrice).String())
}

// NewCoin helper function to initialize sdk.Coin by passing just amount.
func (c ChainContext) NewCoin(amount sdk.Int) sdk.Coin {
	return sdk.NewCoin(c.NetworkConfig.Denom(), amount)
}

// NewDecCoin helper function to initialize sdk.DecCoin by passing just amount.
func (c ChainContext) NewDecCoin(amount sdk.Dec) sdk.DecCoin {
	return sdk.NewDecCoinFromDec(c.NetworkConfig.Denom(), amount)
}

// GasLimitByMsgs calculates sum of gas limits required for message types passed.
// It panics if unsupported message type specified.
func (c ChainContext) GasLimitByMsgs(msgs ...sdk.Msg) uint64 {
	var totalGasRequired uint64
	for _, msg := range msgs {
		msgGas, exists := c.DeterministicGasConfig.GasRequiredByMessage(msg)
		if !exists {
			panic(errors.Errorf("unsuported message type for deterministic gas: %v", reflect.TypeOf(msg).String()))
		}
		totalGasRequired += msgGas + c.DeterministicGasConfig.FixedGas
	}

	return totalGasRequired
}

// GasLimitByMultiSendMsgs calculates sum of gas limits required for message types passed and includes the FixedGas once.
// It panics if unsupported message type specified.
func (c ChainContext) GasLimitByMultiSendMsgs(msgs ...sdk.Msg) uint64 {
	var totalGasRequired uint64
	for _, msg := range msgs {
		msgGas, exists := c.DeterministicGasConfig.GasRequiredByMessage(msg)
		if !exists {
			panic(errors.Errorf("unsuported message type for deterministic gas: %v", reflect.TypeOf(msg).String()))
		}
		totalGasRequired += msgGas
	}

	return totalGasRequired + c.DeterministicGasConfig.FixedGas
}

// BalancesOptions is the input type for the ComputeNeededBalanceFromOptions.
type BalancesOptions struct {
	Messages                    []sdk.Msg
	NondeterministicMessagesGas uint64
	GasPrice                    sdk.Dec
	Amount                      sdk.Int
}

// ComputeNeededBalanceFromOptions computes the required balance based on the input options.
func (c ChainContext) ComputeNeededBalanceFromOptions(options BalancesOptions) sdk.Int {
	if options.GasPrice.IsNil() {
		options.GasPrice = c.InitialGasPrice
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

	return totalAmount.Add(options.GasPrice.Mul(sdk.NewDec(int64(options.NondeterministicMessagesGas))).Ceil().RoundInt()).Add(options.Amount)
}

// ChainConfig defines the config arguments required for the test chain initialisation.
type ChainConfig struct {
	ClientContext     client.Context
	GRPCAddress       string
	GaiaClientContext client.Context
	NetworkConfig     config.NetworkConfig
	InitialGasPrice   sdk.Dec
	FundingMnemonic   string
	StakerMnemonics   []string
}

// Chain holds network and client for the blockchain.
type Chain struct {
	ChainContext
	GaiaContext GaiaContext
	Faucet      Faucet
	Governance  Governance
}

// NewChain creates an instance of the new Chain.
func NewChain(cfg ChainConfig) Chain {
	chainCtx := NewChainContext(cfg.ClientContext, cfg.NetworkConfig, cfg.InitialGasPrice)
	governance := NewGovernance(chainCtx, cfg.StakerMnemonics)

	faucetAddr := chainCtx.ImportMnemonic(cfg.FundingMnemonic)
	faucet := NewFaucet(NewChainContext(cfg.ClientContext.WithFromAddress(faucetAddr), cfg.NetworkConfig, cfg.InitialGasPrice))
	return Chain{
		ChainContext: chainCtx,
		GaiaContext: GaiaContext{
			ClientContext: cfg.GaiaClientContext,
		},
		Governance: governance,
		Faucet:     faucet,
	}
}

// GenMultisigAccount generates a multisig account.
func (c ChainContext) GenMultisigAccount(
	t *testing.T,
	signersCount int,
	multisigThreshold int,
) (*sdkmultisig.LegacyAminoPubKey, []string) {
	requireT := require.New(t)
	keyNamesSet := []string{}
	publicKeySet := make([]cryptotypes.PubKey, 0, signersCount)
	for i := 0; i < signersCount; i++ {
		signerKeyInfo, err := c.ClientContext.Keyring().KeyByAddress(c.GenAccount())
		requireT.NoError(err)
		keyNamesSet = append(keyNamesSet, signerKeyInfo.GetName())
		publicKeySet = append(publicKeySet, signerKeyInfo.GetPubKey())
	}

	// create multisig account
	multisigPublicKey := sdkmultisig.NewLegacyAminoPubKey(multisigThreshold, publicKeySet)
	return multisigPublicKey, keyNamesSet
}
