package integration

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/url"

	sdkmath "cosmossdk.io/math"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdkmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	multisigtypes "github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdksigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/CoreumFoundation/coreum/v4/app"
	"github.com/CoreumFoundation/coreum/v4/pkg/client"
	"github.com/CoreumFoundation/coreum/v4/pkg/config"
	coreumkeyring "github.com/CoreumFoundation/coreum/v4/pkg/keyring"
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
	EncodingConfig config.EncodingConfig
	ClientContext  client.Context
	ChainSettings  ChainSettings
}

// NewChainContext returns a new instance if the ChainContext.
func NewChainContext(
	encodingConfig config.EncodingConfig,
	clientCtx client.Context,
	chainSettings ChainSettings,
) ChainContext {
	return ChainContext{
		EncodingConfig: encodingConfig,
		ClientContext:  clientCtx,
		ChainSettings:  chainSettings,
	}
}

// GenAccount generates a new account for the chain with random name and
// private key and stores it in the chains ClientContext Keyring.
func (c ChainContext) GenAccount() sdk.AccAddress {
	// Generate and store a new mnemonic using temporary keyring
	_, mnemonic, err := keyring.NewInMemory(c.EncodingConfig.Codec).NewMnemonic(
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

// MustConvertToBech32Address converts the address to bech32 address string.
func (c ChainContext) MustConvertToBech32Address(address sdk.AccAddress) string {
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

	address, err := keyInfo.GetAddress()
	if err != nil {
		panic(err)
	}

	return address
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
func (c ChainContext) NewCoin(amount sdkmath.Int) sdk.Coin {
	return sdk.NewCoin(c.ChainSettings.Denom, amount)
}

// NewDecCoin helper function to initialize sdk.DecCoin by passing just amount.
func (c ChainContext) NewDecCoin(amount sdk.Dec) sdk.DecCoin {
	return sdk.NewDecCoinFromDec(c.ChainSettings.Denom, amount)
}

// GenMultisigAccount generates a multisig account.
func (c ChainContext) GenMultisigAccount(
	signersCount, multisigThreshold int,
) (*sdkmultisig.LegacyAminoPubKey, []string, error) {
	keyNamesSet := []string{}
	publicKeySet := make([]cryptotypes.PubKey, 0, signersCount)
	for i := 0; i < signersCount; i++ {
		signerKeyInfo, err := c.ClientContext.Keyring().KeyByAddress(c.GenAccount())
		if err != nil {
			return nil, nil, err
		}
		pubKey, err := signerKeyInfo.GetPubKey()
		if err != nil {
			return nil, nil, err
		}
		keyNamesSet = append(keyNamesSet, signerKeyInfo.Name)
		publicKeySet = append(publicKeySet, pubKey)
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
	pubKey, err := keyInfo.GetPubKey()
	if err != nil {
		return nil, err
	}
	multisigPubKey, ok := pubKey.(*sdkmultisig.LegacyAminoPubKey)
	if !ok {
		return nil, errors.New("public key cannot be converted to multisig public key")
	}
	address, err := keyInfo.GetAddress()
	if err != nil {
		return nil, errors.New("failed to get address from key")
	}
	multisigAccInfo, err := client.GetAccountInfo(ctx, c.ClientContext, address)
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
func NewChain(
	grpcClient *grpc.ClientConn,
	rpcClient rpcclient.Client,
	chainSettings ChainSettings,
	fundingMnemonic string,
) Chain {
	encodingConfig := config.NewEncodingConfig(app.ModuleBasics)

	clientCtxConfig := client.DefaultContextConfig()
	clientCtxConfig.GasConfig.GasPriceAdjustment = sdk.NewDec(1)
	clientCtxConfig.GasConfig.GasAdjustment = 1
	clientCtx := client.NewContext(clientCtxConfig, app.ModuleBasics).
		WithChainID(chainSettings.ChainID).
		WithKeyring(coreumkeyring.NewConcurrentSafeKeyring(keyring.NewInMemory(encodingConfig.Codec))).
		WithBroadcastMode(flags.BroadcastSync).
		WithGRPCClient(grpcClient).
		WithClient(rpcClient).
		WithAwaitTx(true)

	chainCtx := NewChainContext(encodingConfig, clientCtx, chainSettings)

	var faucet Faucet
	if fundingMnemonic != "" {
		faucetAddr := chainCtx.ImportMnemonic(fundingMnemonic)
		faucet = NewFaucet(NewChainContext(encodingConfig, clientCtx.WithFromAddress(faucetAddr), chainSettings))
	}

	return Chain{
		ChainContext: chainCtx,
		Faucet:       faucet,
		Wasm:         NewWasm(chainCtx),
	}
}

// QueryChainSettings queries the chain setting using the provided GRPC client.
func QueryChainSettings(ctx context.Context, grpcClient *grpc.ClientConn) ChainSettings {
	clientCtx := client.NewContext(client.DefaultContextConfig(), app.ModuleBasics).
		WithGRPCClient(grpcClient)

	infoBeforeRes, err := tmservice.NewServiceClient(clientCtx).GetNodeInfo(ctx, &tmservice.GetNodeInfoRequest{})
	if err != nil {
		panic(fmt.Sprintf("failed to get node info, err: %s", err))
	}

	chainID := infoBeforeRes.DefaultNodeInfo.Network

	paramsRes, err := stakingtypes.NewQueryClient(clientCtx).Params(ctx, &stakingtypes.QueryParamsRequest{})
	if err != nil {
		panic(errors.Errorf("failed to get staking params, err: %s", err))
	}

	denom := paramsRes.Params.BondDenom

	accountsRes, err := authtypes.NewQueryClient(clientCtx).Accounts(ctx, &authtypes.QueryAccountsRequest{})
	if err != nil {
		panic(fmt.Sprintf("failed to get account params, err: %s", err))
	}

	var addressPrefix string
	for _, account := range accountsRes.Accounts {
		if account != nil && account.TypeUrl == fmt.Sprintf("/%s", proto.MessageName(&authtypes.BaseAccount{})) {
			var acc authtypes.BaseAccount
			if err := proto.Unmarshal(account.Value, &acc); err != nil {
				panic(fmt.Sprintf("failed to unpack account, err: %s", err))
			}

			addressPrefix, _, err = bech32.DecodeAndConvert(acc.Address)
			if err != nil {
				panic(fmt.Sprintf("failed to extract address prefix address:%s, err: %s", acc.Address, err))
			}
			break
		}
	}
	if addressPrefix == "" {
		panic("address prefix is empty")
	}

	return ChainSettings{
		ChainID:       chainID,
		Denom:         denom,
		AddressPrefix: addressPrefix,
	}
}

// DialGRPCClient creates the grpc connection for the given URL.
func DialGRPCClient(grpcURL string) (*grpc.ClientConn, error) {
	encodingConfig := config.NewEncodingConfig(app.ModuleBasics)
	pc, ok := encodingConfig.Codec.(codec.GRPCCodecProvider)
	if !ok {
		panic("failed to cast codec to codec.GRPCCodecProvider)")
	}

	parsedURL, err := url.Parse(grpcURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse grpc URL")
	}

	host := parsedURL.Host

	// https - tls grpc
	if parsedURL.Scheme == "https" {
		grpcClient, err := grpc.Dial(
			host,
			grpc.WithDefaultCallOptions(grpc.ForceCodec(pc.GRPCCodec())),
			grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{})),
		)
		if err != nil {
			return nil, errors.Wrap(err, "failed to dial grpc")
		}
		return grpcClient, nil
	}

	// handling of host:port URL without the protocol
	if host == "" {
		host = fmt.Sprintf("%s:%s", parsedURL.Scheme, parsedURL.Opaque)
	}
	// http - insecure
	grpcClient, err := grpc.Dial(
		host,
		grpc.WithDefaultCallOptions(grpc.ForceCodec(pc.GRPCCodec())),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, errors.Wrap(err, "failed to dial grpc")
	}

	return grpcClient, nil
}
