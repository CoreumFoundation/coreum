//go:build integrationtests

package ibc

import (
	"context"
	_ "embed"
	"encoding/json"
	"reflect"
	"testing"
	"time"
	"unsafe"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v4/modules/core/04-channel/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
)

var (
	//go:embed testdata/wasm/ibc-transfer/artifacts/ibc_transfer.wasm
	ibcTransferWASM []byte
	//go:embed testdata/wasm/ibc-call/artifacts/ibc_call.wasm
	ibcClassWASM []byte
)

type ibcTimeoutBlock struct {
	Revision uint64 `json:"revision"`
	Height   uint64 `json:"height"`
}

type ibcTimeout struct {
	Block ibcTimeoutBlock `json:"block"`
}

//nolint:tagliatelle // wasm requirements
type ibcTransferRequest struct {
	ChannelID string     `json:"channel_id"`
	ToAddress string     `json:"to_address"`
	Amount    sdk.Coin   `json:"amount"`
	Timeout   ibcTimeout `json:"timeout"`
}

type ibcTransferMethod string

const (
	ibcTransferMethodTransfer ibcTransferMethod = "transfer"
)

type ibcCallChannelRequest struct {
	Channel string `json:"channel"`
}

type ibcCallCountResponse struct {
	Count uint32 `json:"count"`
}

type ibcCallMethod string

const (
	ibcCallMethodIncrement ibcCallMethod = "increment"
	ibcCallMethodGetCount  ibcCallMethod = "get_count"
)

// TestIBCTransferFromSmartContract tests the IBCTransfer from the contract.
func TestIBCTransferFromSmartContract(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	osmosisChain := chains.Osmosis

	osmosisToCoreumChannelID := osmosisChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID)
	coreumToOsmosisChannelID := coreumChain.AwaitForIBCChannelID(ctx, t, ibctransfertypes.PortID, osmosisChain.ChainSettings.ChainID)

	coreumAdmin := coreumChain.GenAccount()
	osmosisRecipient := osmosisChain.GenAccount()

	coreumChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: coreumAdmin,
		Amount:  coreumChain.NewCoin(sdk.NewInt(2000000)),
	})
	sendToOsmosisCoin := coreumChain.NewCoin(sdk.NewInt(1000))

	coreumBankClient := banktypes.NewQueryClient(coreumChain.ClientContext)

	// deploy the contract and fund it
	initialPayload, err := json.Marshal(struct{}{})
	requireT.NoError(err)
	contractAddr, _, err := coreumChain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		coreumChain.TxFactory().WithSimulateAndExecute(true),
		coreumAdmin,
		ibcTransferWASM,
		integrationtests.InstantiateConfig{
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Amount:     sendToOsmosisCoin,
			Label:      "ibc_transfer",
		},
	)
	requireT.NoError(err)

	// get the contract balance and check total
	contractBalance, err := coreumBankClient.Balance(ctx,
		&banktypes.QueryBalanceRequest{
			Address: contractAddr,
			Denom:   sendToOsmosisCoin.Denom,
		})
	requireT.NoError(err)
	requireT.Equal(sendToOsmosisCoin.Amount.String(), contractBalance.Balance.Amount.String())

	coreumChainHeight, err := coreumChain.GetLatestConsensusHeight(
		ctx,
		ibctransfertypes.PortID,
		coreumToOsmosisChannelID,
	)
	requireT.NoError(err)

	transferPayload, err := json.Marshal(map[ibcTransferMethod]ibcTransferRequest{
		ibcTransferMethodTransfer: {
			ChannelID: coreumToOsmosisChannelID,
			ToAddress: osmosisChain.ConvertToBech32Address(osmosisRecipient),
			Amount:    sendToOsmosisCoin,
			Timeout: ibcTimeout{
				Block: ibcTimeoutBlock{
					Revision: coreumChainHeight.RevisionNumber,
					Height:   coreumChainHeight.RevisionHeight + 1000,
				},
			},
		},
	})
	requireT.NoError(err)

	_, err = coreumChain.Wasm.ExecuteWASMContract(
		ctx,
		coreumChain.TxFactory().WithSimulateAndExecute(true),
		coreumAdmin,
		contractAddr,
		transferPayload,
		sdk.Coin{},
	)
	requireT.NoError(err)

	contractBalance, err = coreumBankClient.Balance(ctx,
		&banktypes.QueryBalanceRequest{
			Address: contractAddr,
			Denom:   sendToOsmosisCoin.Denom,
		})
	requireT.NoError(err)
	requireT.Equal(sdk.ZeroInt().String(), contractBalance.Balance.Amount.String())

	expectedOsmosisRecipientBalance := sdk.NewCoin(convertToIBCDenom(osmosisToCoreumChannelID, sendToOsmosisCoin.Denom), sendToOsmosisCoin.Amount)
	osmosisChain.AwaitForBalance(ctx, t, osmosisRecipient, expectedOsmosisRecipientBalance)
}

// TestIBCCallFromSmartContract tests the IBC contract calls.
//
//nolint:funlen // there are many tests
func TestIBCCallFromSmartContract(t *testing.T) {
	// we don't enable the t.Parallel here since that test uses the config unseal hack because of the cosmos relayer
	// implementation
	restoreSDKConfig := unsealSDKConfig()
	defer restoreSDKConfig()

	// channelIBCVersion is the version defined in the ibc.rs in the smart contract
	const channelIBCVersion = "counter-1"

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	osmosisChain := chains.Osmosis

	coreumWasmClient := wasmtypes.NewQueryClient(coreumChain.ClientContext)
	osmosisWasmClient := wasmtypes.NewQueryClient(osmosisChain.ClientContext)

	coreumCaller := coreumChain.GenAccount()
	osmosisCaller := osmosisChain.GenAccount()

	coreumChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: coreumCaller,
		Amount:  coreumChain.NewCoin(sdk.NewInt(2000000)),
	})

	osmosisChain.Faucet.FundAccounts(ctx, t, integrationtests.FundedAccount{
		Address: osmosisCaller,
		Amount:  osmosisChain.NewCoin(sdk.NewInt(2000000)),
	})

	initialPayload, err := json.Marshal(struct{}{})
	requireT.NoError(err)

	coreumContractAddr, _, err := coreumChain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		coreumChain.TxFactory().WithSimulateAndExecute(true),
		coreumCaller,
		ibcClassWASM,
		integrationtests.InstantiateConfig{
			Admin:      coreumCaller,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Label:      "ibc_call",
		},
	)
	requireT.NoError(err)

	osmosisContractAddr, _, err := osmosisChain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		osmosisChain.TxFactory().WithSimulateAndExecute(true),
		osmosisCaller,
		ibcClassWASM,
		integrationtests.InstantiateConfig{
			Admin:      osmosisCaller,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Label:      "ibc_call",
		},
	)
	requireT.NoError(err)

	coreumContractInfoRes, err := coreumWasmClient.ContractInfo(ctx, &wasmtypes.QueryContractInfoRequest{
		Address: coreumContractAddr,
	})
	requireT.NoError(err)
	coreumIBCPort := coreumContractInfoRes.ContractInfo.IBCPortID
	requireT.NotEmpty(coreumIBCPort)
	t.Logf("Coreum contrac IBC port:%s", coreumIBCPort)

	osmosisContractInfoRes, err := osmosisWasmClient.ContractInfo(ctx, &wasmtypes.QueryContractInfoRequest{
		Address: osmosisContractAddr,
	})
	requireT.NoError(err)
	osmosisIBCPort := osmosisContractInfoRes.ContractInfo.IBCPortID
	requireT.NotEmpty(osmosisIBCPort)
	t.Logf("Osmisis contrac IBC port:%s", osmosisIBCPort)

	integrationtests.CreateIBCChannelsAndConnect(
		ctx,
		t,
		coreumChain.Chain,
		coreumIBCPort,
		osmosisChain,
		osmosisIBCPort,
		channelIBCVersion,
		ibcchanneltypes.UNORDERED,
	)

	coreumToOsmosisChannelID := coreumChain.AwaitForIBCChannelID(ctx, t, coreumIBCPort, osmosisChain.ChainSettings.ChainID)
	osmosisToCoreumChannelID := osmosisChain.AwaitForIBCChannelID(ctx, t, osmosisIBCPort, coreumChain.ChainSettings.ChainID)
	t.Logf("Channels are ready coreum channel ID:%s, osmosis channel ID:%s", coreumToOsmosisChannelID, osmosisToCoreumChannelID)

	t.Logf("Sendng two IBC transactions from coreum contract to osmosis contract")
	awaitWasmCallContractCounter(ctx, t, coreumChain.Chain, coreumToOsmosisChannelID, coreumContractAddr, 0)
	awaitWasmCallContractCounter(ctx, t, osmosisChain, osmosisToCoreumChannelID, osmosisContractAddr, 0)

	// execute coreum counter twice
	executeWasmCallContractCounter(ctx, requireT, coreumChain.Chain, coreumCaller, coreumToOsmosisChannelID, coreumContractAddr)
	executeWasmCallContractCounter(ctx, requireT, coreumChain.Chain, coreumCaller, coreumToOsmosisChannelID, coreumContractAddr)

	// check that current state is expected, the order of assertion is important
	awaitWasmCallContractCounter(ctx, t, osmosisChain, osmosisToCoreumChannelID, osmosisContractAddr, 2)
	awaitWasmCallContractCounter(ctx, t, coreumChain.Chain, coreumToOsmosisChannelID, coreumContractAddr, 0)

	t.Logf("Sendng three IBC transactions from osmosis contract to coreum contract")
	executeWasmCallContractCounter(ctx, requireT, osmosisChain, osmosisCaller, osmosisToCoreumChannelID, osmosisContractAddr)
	executeWasmCallContractCounter(ctx, requireT, osmosisChain, osmosisCaller, osmosisToCoreumChannelID, osmosisContractAddr)
	executeWasmCallContractCounter(ctx, requireT, osmosisChain, osmosisCaller, osmosisToCoreumChannelID, osmosisContractAddr)

	// check that current state is expected, the order of assertion is important
	awaitWasmCallContractCounter(ctx, t, coreumChain.Chain, coreumToOsmosisChannelID, coreumContractAddr, 3)
	awaitWasmCallContractCounter(ctx, t, osmosisChain, osmosisToCoreumChannelID, osmosisContractAddr, 2)
}

func executeWasmCallContractCounter(
	ctx context.Context,
	requireT *require.Assertions,
	chain integrationtests.Chain,
	caller sdk.AccAddress,
	channelID, contractAddr string,
) {
	incrementPayload, err := json.Marshal(map[ibcCallMethod]ibcCallChannelRequest{
		ibcCallMethodIncrement: {
			Channel: channelID,
		},
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(
		ctx,
		chain.TxFactory().WithSimulateAndExecute(true),
		caller,
		contractAddr,
		incrementPayload,
		sdk.Coin{},
	)
	requireT.NoError(err)
}

func awaitWasmCallContractCounter(
	ctx context.Context,
	t *testing.T,
	chain integrationtests.Chain,
	channelID, contractAddress string,
	expectedCount uint32,
) {
	t.Helper()

	t.Logf("Awaiting for count:%d, chainID: %s, channel:%s", expectedCount, chain.ChainSettings.ChainID, channelID)

	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute)
	defer retryCancel()
	require.NoError(t, retry.Do(retryCtx, time.Second, func() error {
		getCountPayload, err := json.Marshal(map[ibcCallMethod]ibcCallChannelRequest{
			ibcCallMethodGetCount: {
				Channel: channelID,
			},
		})
		require.NoError(t, err)
		queryCountOut, err := chain.Wasm.QueryWASMContract(retryCtx, contractAddress, getCountPayload)
		require.NoError(t, err)
		var queryCountRes ibcCallCountResponse
		err = json.Unmarshal(queryCountOut, &queryCountRes)
		require.NoError(t, err)

		if queryCountRes.Count != expectedCount {
			return retry.Retryable(errors.Errorf("counter is still not equal to expected, current:%d, expected:%d", queryCountRes.Count, expectedCount))
		}

		return nil
	}))

	t.Logf("Received expected count of %d.", expectedCount)
}

func unsealSDKConfig() func() {
	config := sdk.GetConfig()
	// unseal the config
	setField(config, "sealed", false)
	setField(config, "sealedch", make(chan struct{}))

	bech32AccountAddrPrefix := config.GetBech32AccountAddrPrefix()
	bech32AccountPubPrefix := config.GetBech32AccountPubPrefix()
	bech32ValidatorAddrPrefix := config.GetBech32ValidatorAddrPrefix()
	bech32ValidatorPubPrefix := config.GetBech32ValidatorPubPrefix()
	bech32ConsensusAddrPrefix := config.GetBech32ConsensusAddrPrefix()
	bech32ConsensusPubPrefix := config.GetBech32ConsensusPubPrefix()
	coinType := config.GetCoinType()

	return func() {
		config.SetBech32PrefixForAccount(bech32AccountAddrPrefix, bech32AccountPubPrefix)
		config.SetBech32PrefixForValidator(bech32ValidatorAddrPrefix, bech32ValidatorPubPrefix)
		config.SetBech32PrefixForConsensusNode(bech32ConsensusAddrPrefix, bech32ConsensusPubPrefix)
		config.SetCoinType(coinType)

		config.Seal()
	}
}

func setField(object interface{}, fieldName string, value interface{}) {
	rs := reflect.ValueOf(object).Elem()
	field := rs.FieldByName(fieldName)
	// rf can't be read or set.
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
}
