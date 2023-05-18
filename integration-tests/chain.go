package integrationtests

import (
	"context"
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdkmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v4/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"github.com/cosmos/ibc-go/v4/modules/core/exported"
	ibctmlightclienttypes "github.com/cosmos/ibc-go/v4/modules/light-clients/07-tendermint/types"
	protobufgrpc "github.com/gogo/protobuf/grpc"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
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

	return multisigPublicKey, keyNamesSet, nil
}

// ExecuteIBCTransfer executes IBC transfer transaction.
func (c ChainContext) ExecuteIBCTransfer(
	ctx context.Context,
	senderAddress sdk.AccAddress,
	coin sdk.Coin,
	recipientChainContext ChainContext,
	recipientAddress sdk.AccAddress,
) (*sdk.TxResponse, error) {
	log := logger.Get(ctx)
	sender := c.ConvertToBech32Address(senderAddress)
	receiver := recipientChainContext.ConvertToBech32Address(recipientAddress)
	log.Info(fmt.Sprintf("Sending IBC transfer from %s, to %s, %s.", sender, receiver, coin.String()))

	recipientChannelID, err := c.GetIBCChannelID(ctx, recipientChainContext.ChainSettings.ChainID)
	if err != nil {
		return nil, err
	}

	height, err := queryLatestConsensusHeight(
		ctx,
		c.ClientContext,
		ibctransfertypes.PortID,
		recipientChannelID,
	)
	if err != nil {
		return nil, err
	}

	ibcSend := ibctransfertypes.MsgTransfer{
		SourcePort:    ibctransfertypes.PortID,
		SourceChannel: recipientChannelID,
		Token:         coin,
		Sender:        sender,
		Receiver:      receiver,
		TimeoutHeight: ibcclienttypes.Height{
			RevisionNumber: height.RevisionNumber,
			RevisionHeight: height.RevisionHeight + 1000,
		},
	}

	return BroadcastTxWithSigner(
		ctx,
		c,
		c.TxFactory().WithSimulateAndExecute(true),
		senderAddress,
		&ibcSend,
	)
}

// AwaitForBalance queries for the balance with retry and timeout.
func (c ChainContext) AwaitForBalance(
	ctx context.Context,
	address sdk.AccAddress,
	coin sdk.Coin,
) error {
	log := logger.Get(ctx)
	log.Info(fmt.Sprintf("Waiting for account %s balance, expected amount:%s.", c.ConvertToBech32Address(address), coin.String()))

	bankClient := banktypes.NewQueryClient(c.ClientContext)
	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute)
	defer retryCancel()
	err := retry.Do(retryCtx, time.Second, func() error {
		requestCtx, requestCancel := context.WithTimeout(retryCtx, 5*time.Second)
		defer requestCancel()

		balancesRes, err := bankClient.AllBalances(requestCtx, &banktypes.QueryAllBalancesRequest{
			Address: c.ConvertToBech32Address(address),
		})
		if err != nil {
			return err
		}

		if balancesRes.Balances.AmountOf(coin.Denom).String() != coin.Amount.String() {
			return retry.Retryable(errors.Errorf("balances is still not enough, all balances:%s", balancesRes.String()))
		}

		return nil
	})
	if err != nil {
		return err
	}
	log.Info("Received expected amount.")

	return nil
}

// GetIBCChannelID returns the first opened channel of the IBC connected chain peer.
func (c ChainContext) GetIBCChannelID(ctx context.Context, peerChainID string) (string, error) {
	log := logger.Get(ctx)
	log.Info(fmt.Sprintf("Getting %s chain channel on %s.", peerChainID, c.ChainSettings.ChainID))

	retryCtx, retryCancel := context.WithTimeout(ctx, 3*time.Minute)
	defer retryCancel()

	ibcClient := ibcchanneltypes.NewQueryClient(c.ClientContext)
	ibcChannelClient := ibcchanneltypes.NewQueryClient(c.ClientContext)

	var channelID string
	if err := retry.Do(retryCtx, time.Second, func() error {
		requestCtx, requestCancel := context.WithTimeout(ctx, 5*time.Second)
		defer requestCancel()

		ibcChannelsRes, err := ibcClient.Channels(requestCtx, &ibcchanneltypes.QueryChannelsRequest{})
		if err != nil {
			return err
		}

		for _, ch := range ibcChannelsRes.Channels {
			if ch.State != ibcchanneltypes.OPEN {
				continue
			}

			channelClientStateRes, err := ibcChannelClient.ChannelClientState(requestCtx, &ibcchanneltypes.QueryChannelClientStateRequest{
				PortId:    ibctransfertypes.PortID,
				ChannelId: ch.ChannelId,
			})
			if err != nil {
				return err
			}

			var clientState ibctmlightclienttypes.ClientState
			err = c.ClientContext.Codec().Unmarshal(channelClientStateRes.IdentifiedClientState.ClientState.Value, &clientState)
			if err != nil {
				return err
			}

			if clientState.ChainId == peerChainID {
				channelID = ch.ChannelId
				return nil
			}
		}

		return retry.Retryable(errors.Errorf("waiting for the %s channel on the %s to open", peerChainID, c.ChainSettings.ChainID))
	}); err != nil {
		return "", err
	}

	log.Info(fmt.Sprintf("Got %s chain channel on %s, channelID:%s ", peerChainID, c.ChainSettings.ChainID, channelID))

	return channelID, nil
}

func queryLatestConsensusHeight(ctx context.Context, clientCtx client.Context, portID, channelID string) (ibcclienttypes.Height, error) {
	queryClient := ibcchanneltypes.NewQueryClient(clientCtx)
	req := &ibcchanneltypes.QueryChannelClientStateRequest{
		PortId:    portID,
		ChannelId: channelID,
	}

	clientRes, err := queryClient.ChannelClientState(ctx, req)
	if err != nil {
		return ibcclienttypes.Height{}, err
	}

	var clientState exported.ClientState
	if err := clientCtx.InterfaceRegistry().UnpackAny(clientRes.IdentifiedClientState.ClientState, &clientState); err != nil {
		return ibcclienttypes.Height{}, err
	}

	clientHeight, ok := clientState.GetLatestHeight().(ibcclienttypes.Height)
	if !ok {
		return ibcclienttypes.Height{}, sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "invalid height type. expected type: %T, got: %T",
			ibcclienttypes.Height{}, clientHeight)
	}

	return clientHeight, nil
}

// Chain holds network and client for the blockchain.
type Chain struct {
	ChainContext
	Faucet Faucet
}

// NewChain creates an instance of the new Chain.
func NewChain(grpcClient protobufgrpc.ClientConn, chainSettings ChainSettings, fundingMnemonic string) Chain {
	clientCtxConfig := client.DefaultContextConfig()
	clientCtxConfig.GasConfig.GasPriceAdjustment = sdk.NewDec(1)
	clientCtxConfig.GasConfig.GasAdjustment = 1
	clientCtx := client.NewContext(clientCtxConfig, app.ModuleBasics).
		WithChainID(chainSettings.ChainID).
		WithKeyring(newConcurrentSafeKeyring(keyring.NewInMemory())).
		WithBroadcastMode(flags.BroadcastBlock).
		WithGRPCClient(grpcClient)

	chainCtx := NewChainContext(clientCtx, chainSettings)

	var faucet Faucet
	if fundingMnemonic != "" {
		faucetAddr := chainCtx.ImportMnemonic(fundingMnemonic)
		faucet = NewFaucet(NewChainContext(clientCtx.WithFromAddress(faucetAddr), chainSettings))
	}

	return Chain{
		ChainContext: chainCtx,
		Faucet:       faucet,
	}
}
