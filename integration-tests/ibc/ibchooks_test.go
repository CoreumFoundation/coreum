//go:build integrationtests

package ibc

import (
	"encoding/json"
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	osmosisContractAdmin := osmosisChain.GenAccount()
	osmosisCaller := osmosisChain.GenAccount()
	osmosisRecepient := osmosisChain.GenAccount()

	coreumChain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: coreumContractAdmin,
			Amount:  coreumChain.NewCoin(sdkmath.NewInt(2000000)),
		},
		integration.FundedAccount{
			Address: coreumCaller,
			Amount:  coreumChain.NewCoin(sdkmath.NewInt(2000000)),
		},
	)

	osmosisChain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: osmosisContractAdmin,
			Amount:  osmosisChain.NewCoin(sdkmath.NewInt(2000000)),
		},
		integration.FundedAccount{
			Address: osmosisCaller,
			Amount:  osmosisChain.NewCoin(sdkmath.NewInt(2000000)),
		},
	)

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

	sendToOsmosisCoin := coreumChain.NewCoin(sdkmath.NewInt(1000))
	txRes, err := coreumChain.ExecuteIBCTransfer(
		ctx, t, coreumCaller, sendToOsmosisCoin, osmosisChain.ChainContext, osmosisRecepient,
	)
	requireT.NoError(err)
	fmt.Println(txRes.TxHash)

	expectedOsmosisRecipientBalance := sdk.NewCoin(
		ConvertToIBCDenom(osmosisToCoreumChannelID, sendToOsmosisCoin.Denom),
		sendToOsmosisCoin.Amount,
	)
	requireT.NoError(osmosisChain.AwaitForBalance(ctx, t, osmosisRecepient, expectedOsmosisRecipientBalance))

}
