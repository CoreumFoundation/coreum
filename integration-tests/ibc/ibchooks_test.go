//go:build integrationtests

package ibc

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	ibcwasm "github.com/CoreumFoundation/coreum/v4/integration-tests/contracts/ibc"
	"github.com/CoreumFoundation/coreum/v4/testutil/integration"
)

func TestIBCHooksCounter(t *testing.T) {
	// we don't enable the t.Parallel here since that test uses the config unseal hack because of the cosmos relayer
	// implementation
	//restoreSDKConfig := unsealSDKConfig()
	//defer restoreSDKConfig()

	// channelIBCVersion is the version defined in the ibc.rs in the smart contract
	//const channelIBCVersion = "counter-1"

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	osmosisChain := chains.Osmosis

	//coreumWasmClient := wasmtypes.NewQueryClient(coreumChain.ClientContext)
	//osmosisWasmClient := wasmtypes.NewQueryClient(osmosisChain.ClientContext)

	coreumContractAdmin := coreumChain.GenAccount()
	coreumCaller := coreumChain.GenAccount()
	coreumRecipient := coreumChain.GenAccount()

	osmosisContractAdmin := osmosisChain.GenAccount()
	osmosisCaller := osmosisChain.GenAccount()
	//osmosisRecipient := osmosisChain.GenAccount()

	coreumChain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: coreumContractAdmin,
			Amount:  coreumChain.NewCoin(sdkmath.NewInt(20_000_000)),
		},
		integration.FundedAccount{
			Address: coreumCaller,
			Amount:  coreumChain.NewCoin(sdkmath.NewInt(20_000_000)),
		},
	)

	osmosisChain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: osmosisContractAdmin,
			Amount:  osmosisChain.NewCoin(sdkmath.NewInt(20_000_000)),
		},
		integration.FundedAccount{
			Address: osmosisCaller,
			Amount:  osmosisChain.NewCoin(sdkmath.NewInt(20_000_000)),
		},
	)

	// ***** Deploy contract *****//

	// instantiate the contract and set the initial counter state.
	initialPayload, err := json.Marshal(ibcwasm.HooksCounterState{
		Count: 2024,
	})
	requireT.NoError(err)

	coreumContractAddr, _, err := coreumChain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		coreumChain.TxFactory().WithSimulateAndExecute(true),
		coreumContractAdmin,
		ibcwasm.IBCHooksCounter,
		integration.InstantiateConfig{
			Admin:      coreumContractAdmin,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Label:      "ibc_hooks_counter",
		},
	)
	requireT.NoError(err)
	fmt.Println(coreumContractAddr)

	osmosisToCoreumChannelID := osmosisChain.AwaitForIBCChannelID(
		ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID,
	)

	// ***** Send funds to Osmosis ****//

	sendToOsmosisCoin := coreumChain.NewCoin(sdkmath.NewInt(10_000_000))
	txRes, err := coreumChain.ExecuteIBCTransfer(
		ctx, t, coreumCaller, sendToOsmosisCoin, osmosisChain.ChainContext, osmosisCaller,
	)
	requireT.NoError(err)
	fmt.Println(txRes.TxHash)

	expectedOsmosisRecipientBalance := sdk.NewCoin(
		ConvertToIBCDenom(osmosisToCoreumChannelID, sendToOsmosisCoin.Denom),
		sendToOsmosisCoin.Amount,
	)
	requireT.NoError(osmosisChain.AwaitForBalance(ctx, t, osmosisCaller, expectedOsmosisRecipientBalance))

	sendToCoreumCoin := sdk.NewCoin(expectedOsmosisRecipientBalance.Denom, expectedOsmosisRecipientBalance.Amount.Quo(sdk.NewInt(2)))

	// ***** Send IBC Hook Tx *****///

	coreumToOsmosisChannelID := coreumChain.AwaitForIBCChannelID(
		ctx, t, ibctransfertypes.PortID, osmosisChain.ChainSettings.ChainID,
	)
	_ = coreumToOsmosisChannelID

	txRes, err = osmosisChain.ExecuteIBCTransferWithMemo(
		ctx,
		t,
		osmosisCaller,
		sendToCoreumCoin,
		coreumChain.ChainContext,
		coreumContractAddr, // can be empty string ?
		fmt.Sprintf(`{"wasm":{"contract": "%s", "msg":{"increment":{}}}}`, coreumContractAddr),
	)
	//requireT.NoError(err)
	fmt.Println(txRes.RawLog)
	fmt.Println(txRes.TxHash)

	tmQueryClient := tmservice.NewServiceClient(coreumChain.ClientContext)
	res, err := tmQueryClient.GetLatestBlock(ctx, &tmservice.GetLatestBlockRequest{})
	require.NoError(t, err)

	txSvcClient := sdktx.NewServiceClient(coreumChain.ClientContext)

	currentHeight := blockHeightFromResponse(res)
	for block := currentHeight - 10; block < currentHeight+25; block++ {
		fmt.Printf("querying block: %v\n", block)
		res, err := txSvcClient.GetBlockWithTxs(ctx, &sdktx.GetBlockWithTxsRequest{Height: block})
		if err != nil {
			fmt.Println("block not found waiting for 2s")
			<-time.After(2 * time.Second)
			block--
			continue
		}

		if len(res.Txs) > 0 {
			fmt.Printf("total txs in block %v: %v\n", res.Block.Header.Height, len(res.Txs))

			//res.Txs[0].Signatures
			//fmt.Printf("txid: %v", .Body.)
		}
	}

	// ***** Send IBC Transfer back *****///

	if true {
		txRes, err = osmosisChain.ExecuteIBCTransfer(
			ctx, t, osmosisCaller, sendToCoreumCoin, coreumChain.ChainContext, coreumRecipient,
		)
		requireT.NoError(err)
		fmt.Println(txRes.TxHash)

		expectedCoreumRecipientBalance := coreumChain.NewCoin(sendToCoreumCoin.Amount)
		requireT.NoError(coreumChain.AwaitForBalance(ctx, t, coreumRecipient, expectedCoreumRecipientBalance))

		fmt.Println("sleeping before sending IBC hook tx")
		<-time.After(time.Second * 10)
	}
}

func blockHeightFromResponse(res *tmservice.GetLatestBlockResponse) int64 {
	if res.SdkBlock != nil {
		return res.SdkBlock.Header.Height
	}

	return res.Block.Header.Height //nolint:staticcheck // we keep it to keep the compatibility with old versions
}
