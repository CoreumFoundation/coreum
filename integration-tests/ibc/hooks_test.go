//go:build integrationtests

package ibc

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibchookskeeper "github.com/cosmos/ibc-apps/modules/ibc-hooks/v8/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	integrationtests "github.com/CoreumFoundation/coreum/v6/integration-tests"
	ibcwasm "github.com/CoreumFoundation/coreum/v6/integration-tests/contracts/ibc"
	"github.com/CoreumFoundation/coreum/v6/testutil/integration"
)

// TestIBCHooksCounterWASMCall tests ibc-hooks integration by deploying the ibc-hooks-counter WASM contract
// on Coreum and calling it from Osmosis.
func TestIBCHooksCounterWASMCall(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	osmosisChain := chains.Osmosis

	osmosisToCoreumChannelID := osmosisChain.AwaitForIBCChannelID(
		ctx, t, ibctransfertypes.PortID, coreumChain.ChainContext,
	)
	coreumToOsmosisChannelID := coreumChain.AwaitForIBCChannelID(
		ctx, t, ibctransfertypes.PortID, osmosisChain.ChainContext,
	)

	coreumContractAdmin := coreumChain.GenAccount()
	coreumSender := coreumChain.GenAccount()

	osmosisHookCaller1 := osmosisChain.GenAccount()
	osmosisHookCaller2 := osmosisChain.GenAccount()

	coreumChain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: coreumContractAdmin,
			Amount:  coreumChain.NewCoin(sdkmath.NewInt(20_000_000)),
		},
		integration.FundedAccount{
			Address: coreumSender,
			Amount:  coreumChain.NewCoin(sdkmath.NewInt(20_000_000)),
		},
	)

	osmosisChain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: osmosisHookCaller1,
			Amount:  osmosisChain.NewCoin(sdkmath.NewInt(20_000)),
		},
		integration.FundedAccount{
			Address: osmosisHookCaller2,
			Amount:  osmosisChain.NewCoin(sdkmath.NewInt(20_000)),
		},
	)

	// ********** Deploy contract **********

	// instantiate the contract and set the initial counter state.
	initialPayload, err := json.Marshal(ibcwasm.HooksCounterState{
		Count: 2024, // This is the initial counter value for contract instantiator. We don't use this value.
	})
	requireT.NoError(err)

	coreumContractAddr, _, err := coreumChain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		coreumChain.TxFactoryAuto(),
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

	// ********** Send funds to Osmosis **********

	sendToOsmosisCoin := coreumChain.NewCoin(sdkmath.NewInt(10_000))
	_, err = coreumChain.ExecuteIBCTransfer(
		ctx,
		t,
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(&ibctransfertypes.MsgTransfer{})),
		coreumSender,
		sendToOsmosisCoin,
		osmosisChain.ChainContext,
		osmosisHookCaller1,
	)
	requireT.NoError(err)

	expectedOsmosisRecipientBalance := sdk.NewCoin(
		ConvertToIBCDenom(osmosisToCoreumChannelID, sendToOsmosisCoin.Denom),
		sendToOsmosisCoin.Amount,
	)
	requireT.NoError(osmosisChain.AwaitForBalance(ctx, t, osmosisHookCaller1, expectedOsmosisRecipientBalance))

	// ********** Send IBC Hook Txs **********

	sendToCoreumCoin := sdk.NewCoin(
		expectedOsmosisRecipientBalance.Denom,
		expectedOsmosisRecipientBalance.Amount.Quo(sdkmath.NewInt(2)),
	)

	sendOsmosisToCoreumCoin := osmosisChain.NewCoin(sdkmath.NewInt(10_000))
	expectedOsmosisOnCoreumBalance := sdk.NewCoin(
		ConvertToIBCDenom(coreumToOsmosisChannelID, sendOsmosisToCoreumCoin.Denom),
		sendOsmosisToCoreumCoin.Amount,
	)

	ibcHookCallerOnCoreumAddr1, err := ibchookskeeper.DeriveIntermediateSender(
		coreumToOsmosisChannelID,
		osmosisChain.MustConvertToBech32Address(osmosisHookCaller1),
		coreumChain.ChainSettings.AddressPrefix)
	requireT.NoError(err)

	ibcHookCallerOnCoreumAddr2, err := ibchookskeeper.DeriveIntermediateSender(
		coreumToOsmosisChannelID,
		osmosisChain.MustConvertToBech32Address(osmosisHookCaller2),
		coreumChain.ChainSettings.AddressPrefix)
	requireT.NoError(err)

	// Verify that hook caller is separate for each sender address.
	requireT.NotEqual(ibcHookCallerOnCoreumAddr1, ibcHookCallerOnCoreumAddr2)

	// Osmosis tx memo contains info to call WASM contract on Coreum this info is propagated into
	// IBC FungibleTokenPacketData, and will be used by ibc-hook middleware to build wasm.MsgExecuteContract.
	// For more info check:
	// https://github.com/cosmos/ibc-apps/tree/main/modules/ibc-hooks#how-do-ibc-hooks-work
	// https://github.com/cosmos/ibc/blob/main/spec/app/ics-020-fungible-token-transfer/README.md#using-the-memo-field
	ibcHookMemo := fmt.Sprintf(`{"wasm":{"contract": "%s", "msg":{"increment":{}}}}`, coreumContractAddr)
	// Caller1 first iteration.
	_, err = osmosisChain.ExecuteIBCTransferWithMemo(
		ctx,
		t,
		osmosisChain.TxFactoryAuto(),
		osmosisHookCaller1,
		sendToCoreumCoin,
		coreumChain.ChainContext,
		coreumContractAddr,
		ibcHookMemo,
	)
	requireT.NoError(err)
	awaitHooksCounterContractState(
		ctx,
		t,
		coreumChain,
		coreumContractAddr,
		ibcHookCallerOnCoreumAddr1,
		0,
		sdk.Coins{coreumChain.NewCoin(sendToCoreumCoin.Amount)},
	)

	// Caller1 second iteration.
	_, err = osmosisChain.ExecuteIBCTransferWithMemo(
		ctx,
		t,
		osmosisChain.TxFactoryAuto(),
		osmosisHookCaller1,
		sendToCoreumCoin,
		coreumChain.ChainContext,
		coreumContractAddr,
		ibcHookMemo,
	)
	requireT.NoError(err)
	awaitHooksCounterContractState(
		ctx,
		t,
		coreumChain,
		coreumContractAddr,
		ibcHookCallerOnCoreumAddr1,
		1,
		sdk.Coins{coreumChain.NewCoin(sendToCoreumCoin.Amount.Add(sendToCoreumCoin.Amount))},
	)

	// Caller2 first iteration.
	_, err = osmosisChain.ExecuteIBCTransferWithMemo(
		ctx,
		t,
		osmosisChain.TxFactoryAuto(),
		osmosisHookCaller2,
		sendOsmosisToCoreumCoin,
		coreumChain.ChainContext,
		coreumContractAddr,
		ibcHookMemo,
	)
	requireT.NoError(err)
	awaitHooksCounterContractState(
		ctx,
		t,
		coreumChain,
		coreumContractAddr,
		ibcHookCallerOnCoreumAddr2,
		0,
		sdk.Coins{expectedOsmosisOnCoreumBalance},
	)
}

// TestIBCHooksCounterWASMCallback tests ibc-hooks integration by deploying the ibc-hooks-counter WASM contract
// on Coreum and using it as a callback for IBC transfer sent to Osmosis.
func TestIBCHooksCounterWASMCallback(t *testing.T) {
	t.Parallel()

	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	osmosisChain := chains.Osmosis

	osmosisToCoreumChannelID := osmosisChain.AwaitForIBCChannelID(
		ctx, t, ibctransfertypes.PortID, coreumChain.ChainContext,
	)

	coreumContractAdmin := coreumChain.GenAccount()
	coreumSender := coreumChain.GenAccount()

	osmosisReceiver := osmosisChain.GenAccount()

	coreumChain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: coreumContractAdmin,
			Amount:  coreumChain.NewCoin(sdkmath.NewInt(20_000_000)),
		},
		integration.FundedAccount{
			Address: coreumSender,
			Amount:  coreumChain.NewCoin(sdkmath.NewInt(20_000_000)),
		},
	)

	// ********** Deploy contract **********

	// instantiate the contract and set the initial counter state.
	initialPayload, err := json.Marshal(ibcwasm.HooksCounterState{
		Count: 0, // This is the initial counter value for contract admin. We don't use this value in tests.
	})
	requireT.NoError(err)

	coreumContractAddr, _, err := coreumChain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		coreumChain.TxFactoryAuto(),
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

	// Sudo IBCAck or IBCTimeout message will be sent to the contract specified in memo.
	// For more details check: https://github.com/cosmos/ibc-apps/blob/main/modules/ibc-hooks/wasm_hook.go#L228
	ibcCallbackMemo := fmt.Sprintf(`{"ibc_callback": "%s"}`, coreumContractAddr)
	sendToOsmosisCoin := coreumChain.NewCoin(sdkmath.NewInt(10_000))
	_, err = coreumChain.ExecuteIBCTransferWithMemo(
		ctx,
		t,
		coreumChain.TxFactory().WithGas(coreumChain.GasLimitByMsgs(&ibctransfertypes.MsgTransfer{})),
		coreumSender,
		sendToOsmosisCoin,
		osmosisChain.ChainContext,
		osmosisChain.MustConvertToBech32Address(osmosisReceiver),
		ibcCallbackMemo,
	)
	requireT.NoError(err)

	expectedOsmosisRecipientBalance := sdk.NewCoin(
		ConvertToIBCDenom(osmosisToCoreumChannelID, sendToOsmosisCoin.Denom),
		sendToOsmosisCoin.Amount,
	)
	requireT.NoError(osmosisChain.AwaitForBalance(ctx, t, osmosisReceiver, expectedOsmosisRecipientBalance))

	// Contract increments differently in callback logic.
	// For IBCAck counter associated with contract address is incremented by 1 and coins are not added.
	awaitHooksCounterContractState(
		ctx,
		t,
		coreumChain,
		coreumContractAddr,
		coreumContractAddr,
		1,
		sdk.Coins{},
	)
}

func awaitHooksCounterContractState(
	ctx context.Context,
	t *testing.T,
	coreumChain integration.CoreumChain,
	contractAddr string,
	callerAddr string,
	expectedCount int,
	expectedFunds sdk.Coins,
) {
	t.Helper()

	t.Logf("Awaiting for contract state contract:%s address:%s count:%d total_funds:%s",
		contractAddr, callerAddr, expectedCount, expectedFunds.String())

	retryCtx, retryCancel := context.WithTimeout(ctx, time.Minute)
	defer retryCancel()
	require.NoError(t, retry.Do(retryCtx, time.Second, func() error {
		getCountPayload, err := json.Marshal(map[ibcwasm.HooksMethod]ibcwasm.HooksBodyRequest{
			ibcwasm.HooksGetCount: {
				Addr: callerAddr,
			},
		})
		require.NoError(t, err)
		queryCountOut, err := coreumChain.Wasm.QueryWASMContract(ctx, contractAddr, getCountPayload)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return retry.Retryable(errors.New("counter is still not found for address: " + callerAddr))
			}
			require.NoError(t, err)
		}

		var countResponse ibcwasm.HooksCounterState
		require.NoError(t, json.Unmarshal(queryCountOut, &countResponse))
		if countResponse.Count != expectedCount {
			return retry.Retryable(errors.Errorf(
				"counter is still not equal to expected, current:%d, expected:%d",
				countResponse.Count,
				expectedCount,
			))
		}

		getTotalFundsPayload, err := json.Marshal(map[ibcwasm.HooksMethod]ibcwasm.HooksBodyRequest{
			ibcwasm.HooksGetTotalFunds: {
				Addr: callerAddr,
			},
		})
		require.NoError(t, err)
		queryTotalFundsOut, err := coreumChain.Wasm.QueryWASMContract(ctx, contractAddr, getTotalFundsPayload)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return retry.Retryable(errors.New("counter is still not found for address: " + callerAddr))
			}
			require.NoError(t, err)
		}

		var totalFundsResponse ibcwasm.HooksTotalFundsState
		require.NoError(t, json.Unmarshal(queryTotalFundsOut, &totalFundsResponse))
		if !totalFundsResponse.TotalFunds.Equal(expectedFunds) {
			return retry.Retryable(errors.Errorf(
				"total_funds is still not equal to expected, current:%s, expected:%s",
				totalFundsResponse.TotalFunds.String(),
				expectedFunds.String(),
			))
		}
		require.Equal(t, expectedFunds.String(), totalFundsResponse.TotalFunds.String())

		return nil
	}))
}
