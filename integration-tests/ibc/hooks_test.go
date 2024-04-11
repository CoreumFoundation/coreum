//go:build integrationtests

package ibc

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibchookskeeper "github.com/cosmos/ibc-apps/modules/ibc-hooks/v7/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	ibcwasm "github.com/CoreumFoundation/coreum/v4/integration-tests/contracts/ibc"
	"github.com/CoreumFoundation/coreum/v4/testutil/integration"
)

func TestIBCHooksCounter(t *testing.T) {
	ctx, chains := integrationtests.NewChainsTestingContext(t)
	requireT := require.New(t)
	coreumChain := chains.Coreum
	osmosisChain := chains.Osmosis

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
			Amount:  osmosisChain.NewCoin(sdkmath.NewInt(20_000_000)),
		},
		integration.FundedAccount{
			Address: osmosisHookCaller2,
			Amount:  osmosisChain.NewCoin(sdkmath.NewInt(20_000_000)),
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

	osmosisToCoreumChannelID := osmosisChain.AwaitForIBCChannelID(
		ctx, t, ibctransfertypes.PortID, coreumChain.ChainSettings.ChainID,
	)
	coreumToOsmosisChannelID := coreumChain.AwaitForIBCChannelID(
		ctx, t, ibctransfertypes.PortID, osmosisChain.ChainSettings.ChainID,
	)

	// ********** Send funds to Osmosis **********

	sendToOsmosisCoin := coreumChain.NewCoin(sdkmath.NewInt(10_000_000))
	_, err = coreumChain.ExecuteIBCTransfer(
		ctx, t, coreumSender, sendToOsmosisCoin, osmosisChain.ChainContext, osmosisHookCaller1,
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
		expectedOsmosisRecipientBalance.Amount.Quo(sdk.NewInt(2)),
	)

	sendOsmosisToCoreumCoin := osmosisChain.NewCoin(sdk.NewInt(10_000))
	expectedOsmosisOnCoreumBalance := sdk.NewCoin(
		ConvertToIBCDenom(coreumToOsmosisChannelID, sendOsmosisToCoreumCoin.Denom),
		sendOsmosisToCoreumCoin.Amount,
	)

	ibcHookCallerOnCoreumAddr1, err := ibchookskeeper.DeriveIntermediateSender(
		coreumToOsmosisChannelID,
		osmosisChain.MustConvertToBech32Address(osmosisHookCaller1),
		coreumChain.Chain.ChainSettings.AddressPrefix)
	requireT.NoError(err)

	ibcHookCallerOnCoreumAddr2, err := ibchookskeeper.DeriveIntermediateSender(
		coreumToOsmosisChannelID,
		osmosisChain.MustConvertToBech32Address(osmosisHookCaller2),
		coreumChain.Chain.ChainSettings.AddressPrefix)
	requireT.NoError(err)

	// Verify that hook caller is separate for each sender address.
	requireT.NotEqual(ibcHookCallerOnCoreumAddr1, ibcHookCallerOnCoreumAddr2)

	// Tx memo contains info to call WASM contract this info is propagated into IBC FungibleTokenPacketData,
	// and will be used by ibc-hook middleware to build wasm.MsgExecuteContract.
	// For more info check:
	// https://github.com/cosmos/ibc-apps/tree/main/modules/ibc-hooks#how-do-ibc-hooks-work
	// https://github.com/cosmos/ibc/blob/main/spec/app/ics-020-fungible-token-transfer/README.md#using-the-memo-field
	ibcHookMemo := fmt.Sprintf(`{"wasm":{"contract": "%s", "msg":{"increment":{}}}}`, coreumContractAddr)
	// Caller1 first iteration.
	_, err = executeIBCTransferWithMemo(
		ctx,
		t,
		osmosisChain,
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
	_, err = executeIBCTransferWithMemo(
		ctx,
		t,
		osmosisChain,
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
	_, err = executeIBCTransferWithMemo(
		ctx,
		t,
		osmosisChain,
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

	t.Logf("Awaiting for contract state contract:%s count:%d total_funds:%s",
		contractAddr, expectedCount, expectedFunds.String())

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
		require.NoError(t, err)

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
		require.NoError(t, err)
		var totalFundsResponse ibcwasm.HooksTotalFundsState
		require.NoError(t, json.Unmarshal(queryTotalFundsOut, &totalFundsResponse))
		if !totalFundsResponse.TotalFunds.IsEqual(expectedFunds) {
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

// executeIBCTransferWithMemo is similar to ChainContext.ExecuteIBCTransfer method but it also supports memo.
func executeIBCTransferWithMemo(
	ctx context.Context,
	t *testing.T,
	srcChain integration.Chain,
	senderAddress sdk.AccAddress,
	coin sdk.Coin,
	recipientChainContext integration.ChainContext,
	recipientAddress string,
	memo string,
) (*sdk.TxResponse, error) {
	t.Helper()

	sender := srcChain.MustConvertToBech32Address(senderAddress)
	t.Logf("Sending IBC transfer sender: %s, receiver: %s, amount: %s.", sender, recipientAddress, coin.String())

	recipientChannelID := srcChain.AwaitForIBCChannelID(
		ctx,
		t,
		ibctransfertypes.PortID,
		recipientChainContext.ChainSettings.ChainID,
	)
	height, err := srcChain.GetLatestConsensusHeight(
		ctx,
		ibctransfertypes.PortID,
		recipientChannelID,
	)
	require.NoError(t, err)

	ibcSend := ibctransfertypes.MsgTransfer{
		SourcePort:    ibctransfertypes.PortID,
		SourceChannel: recipientChannelID,
		Token:         coin,
		Sender:        sender,
		Receiver:      recipientAddress,
		TimeoutHeight: ibcclienttypes.Height{
			RevisionNumber: height.RevisionNumber,
			RevisionHeight: height.RevisionHeight + 1000,
		},
		Memo: memo,
	}

	return srcChain.BroadcastTxWithSigner(
		ctx,
		srcChain.TxFactory().WithSimulateAndExecute(true),
		senderAddress,
		&ibcSend,
	)
}
