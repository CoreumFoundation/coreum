package integrationtests

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdkmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	multisigtypes "github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdksigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	protobufgrpc "github.com/gogo/protobuf/grpc"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	rpchttp "github.com/tendermint/tendermint/rpc/client/http"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/client"
)

// ChainSettings represent common settings for the chains.
type ChainSettings struct {
	ChainID       string
	Denom         string
	AddressPrefix string
	GasPrice      sdk.Dec
	GasAdjustment float64
	CoinType      uint32
	RPCAddress    string
}

// ChainContext is a types used to store the components required for the test chains subcomponents.
type ChainContext struct {
	ClientContext client.Context
	ChainSettings ChainSettings
}

// NewChainContext returns a new instance if the ChainContext.
func NewChainContext(
	clientCtx client.Context,
	chainSettings ChainSettings,
) ChainContext {
	return ChainContext{
		ClientContext: clientCtx,
		ChainSettings: chainSettings,
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

// ConvertToBech32Address converts the address to bech32 address string.
func (c ChainContext) ConvertToBech32Address(address sdk.AccAddress) string {
	bech32Address, err := bech32.ConvertAndEncode(c.ChainSettings.AddressPrefix, address)
	if err != nil {
		panic(err)
	}

	return bech32Address
}

// ImportMnemonic imports the mnemonic into the ClientContext Keyring and returns its address.
// If the mnemonic is already imported the method will just return the address.
func (c ChainContext) ImportMnemonic(mnemonic string) sdk.AccAddress {
	keyInfo, err := c.ClientContext.Keyring().NewAccount(
		uuid.New().String(),
		mnemonic,
		"",
		hd.CreateHDPath(c.ChainSettings.CoinType, 0, 0).String(),
		hd.Secp256k1,
	)
	if err != nil {
		panic(err)
	}

	return keyInfo.GetAddress()
}

// TxFactory returns factory with present values for the Chain.
func (c ChainContext) TxFactory() client.Factory {
	txf := client.Factory{}.
		WithKeybase(c.ClientContext.Keyring()).
		WithChainID(c.ChainSettings.ChainID).
		WithTxConfig(c.ClientContext.TxConfig()).
		WithGasPrices(c.NewDecCoin(c.ChainSettings.GasPrice).String())
	if c.ChainSettings.GasAdjustment != 0 {
		txf = txf.WithGasAdjustment(c.ChainSettings.GasAdjustment)
	}

	return txf
}

// NewCoin helper function to initialize sdk.Coin by passing just amount.
func (c ChainContext) NewCoin(amount sdk.Int) sdk.Coin {
	return sdk.NewCoin(c.ChainSettings.Denom, amount)
}

// NewDecCoin helper function to initialize sdk.DecCoin by passing just amount.
func (c ChainContext) NewDecCoin(amount sdk.Dec) sdk.DecCoin {
	return sdk.NewDecCoinFromDec(c.ChainSettings.Denom, amount)
}

// GenMultisigAccount generates a multisig account.
func (c ChainContext) GenMultisigAccount(signersCount, multisigThreshold int) (*sdkmultisig.LegacyAminoPubKey, []string, error) {
	keyNamesSet := []string{}
	publicKeySet := make([]cryptotypes.PubKey, 0, signersCount)
	for i := 0; i < signersCount; i++ {
		signerKeyInfo, err := c.ClientContext.Keyring().KeyByAddress(c.GenAccount())
		if err != nil {
			return nil, nil, err
		}
		keyNamesSet = append(keyNamesSet, signerKeyInfo.GetName())
		publicKeySet = append(publicKeySet, signerKeyInfo.GetPubKey())
	}

	// create multisig account
	multisigPublicKey := sdkmultisig.NewLegacyAminoPubKey(multisigThreshold, publicKeySet)

	_, err := c.ClientContext.Keyring().SaveMultisig(uuid.New().String(), multisigPublicKey)
	if err != nil {
		return nil, nil, errors.New("storing multisig public key in keystore failed")
	}

	return multisigPublicKey, keyNamesSet, nil
}

// SignAndBroadcastMultisigTx signs the amino multisig tx with provided key names and broadcasts it.
func (c ChainContext) SignAndBroadcastMultisigTx(
	ctx context.Context,
	clientCtx client.Context,
	txf client.Factory,
	msg sdk.Msg,
	signersKeyNames ...string,
) (*sdk.TxResponse, error) {
	keyInfo, err := txf.Keybase().KeyByAddress(clientCtx.FromAddress())
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if keyInfo.GetAlgo() != hd.MultiType {
		return nil, errors.Errorf("%s is not a multisig account", c.ConvertToBech32Address(keyInfo.GetAddress()))
	}
	multisigPubKey, ok := keyInfo.GetPubKey().(*sdkmultisig.LegacyAminoPubKey)
	if !ok {
		return nil, errors.New("public key cannot be converted to multisig public key")
	}

	multisigAccInfo, err := client.GetAccountInfo(ctx, c.ClientContext, keyInfo.GetAddress())
	if err != nil {
		return nil, err
	}
	txf = txf.WithAccountNumber(multisigAccInfo.GetAccountNumber()).
		WithSequence(multisigAccInfo.GetSequence()).
		WithSignMode(sdksigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)

	// estimate gas and add adjustment
	if txf.SimulateAndExecute() {
		_, gas, err := client.CalculateGas(ctx, clientCtx, txf, msg)
		if err != nil {
			return nil, err
		}
		txf = txf.WithGas(gas)
	}
	if txf.GasAdjustment() != 0 {
		gas := uint64(txf.GasAdjustment() * float64(txf.Gas()))
		txf = txf.WithGas(gas)
	}

	txBuilder, err := txf.BuildUnsignedTx(msg)
	if err != nil {
		return nil, err
	}

	for _, signersKeyName := range signersKeyNames {
		if err := client.Sign(txf, signersKeyName, txBuilder, false); err != nil {
			return nil, err
		}
	}

	signs, err := txBuilder.GetTx().GetSignaturesV2()
	if err != nil {
		return nil, err
	}

	multisigSig := multisigtypes.NewMultisig(len(multisigPubKey.PubKeys))
	for _, sig := range signs {
		if err := multisigtypes.AddSignatureV2(multisigSig, sig, multisigPubKey.GetPubKeys()); err != nil {
			return nil, err
		}
	}

	sigV2 := sdksigning.SignatureV2{
		PubKey:   multisigPubKey,
		Data:     multisigSig,
		Sequence: multisigAccInfo.GetSequence(),
	}

	if err := txBuilder.SetSignatures(sigV2); err != nil {
		return nil, err
	}

	encodedTx, err := c.ClientContext.TxConfig().TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	return client.BroadcastRawTx(ctx, c.ClientContext, encodedTx)
}

// Chain holds network and client for the blockchain.
type Chain struct {
	ChainContext
	Faucet Faucet
	Wasm   Wasm
}

// NewChain creates an instance of the new Chain.
func NewChain(grpcClient protobufgrpc.ClientConn, rpcClient *rpchttp.HTTP, chainSettings ChainSettings, fundingMnemonic string) Chain {
	clientCtxConfig := client.DefaultContextConfig()
	clientCtxConfig.GasConfig.GasPriceAdjustment = sdk.NewDec(1)
	clientCtxConfig.GasConfig.GasAdjustment = 1
	clientCtx := client.NewContext(clientCtxConfig, app.ModuleBasics).
		WithChainID(chainSettings.ChainID).
		WithKeyring(newConcurrentSafeKeyring(keyring.NewInMemory())).
		WithBroadcastMode(flags.BroadcastBlock).
		WithGRPCClient(grpcClient).
		WithRPCClient(rpcClient)

	chainCtx := NewChainContext(clientCtx, chainSettings)

	var faucet Faucet
	if fundingMnemonic != "" {
		faucetAddr := chainCtx.ImportMnemonic(fundingMnemonic)
		faucet = NewFaucet(NewChainContext(clientCtx.WithFromAddress(faucetAddr), chainSettings))
	}

	return Chain{
		ChainContext: chainCtx,
		Faucet:       faucet,
		Wasm:         NewWasm(chainCtx),
	}
}
