package wasm

import (
	"context"
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TestGasWasmBankSendAndBankSend checks that a message containing a deterministic and a
// non-deterministic transaction takes gas within appropriate limits.
func TestGasWasmBankSendAndBankSend(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	admin, err := chain.GenFundedAccount(ctx)
	requireT.NoError(err)

	// deploy and init contract with the initial coins amount
	initialPayload, err := json.Marshal(bankInstantiatePayload{Count: 0})
	requireT.NoError(err)

	clientCtx := chain.ClientContext.WithFromAddress(admin)
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

	contractAddr, err := DeployAndInstantiate(
		ctx,
		clientCtx,
		txf,
		bankSendWASM,
		InstantiateConfig{
			accessType: wasmtypes.AccessTypeUnspecified,
			payload:    initialPayload,
			amount:     chain.NewCoin(sdk.NewInt(10000)),
			label:      "bank_send",
		},
	)
	requireT.NoError(err)

	// Send tokens
	receiver := chain.GenAccount()
	withdrawPayload, err := json.Marshal(map[bankMethod]bankWithdrawRequest{
		withdraw: {
			Amount:    "5000",
			Denom:     chain.NetworkConfig.BaseDenom,
			Recipient: receiver.String(),
		},
	})
	requireT.NoError(err)

	wasmBankSend := &wasmtypes.MsgExecuteContract{
		Sender:   admin.String(),
		Contract: contractAddr,
		Msg:      wasmtypes.RawContractMessage(withdrawPayload),
		Funds:    sdk.Coins{},
	}

	bankSend := &banktypes.MsgSend{
		FromAddress: admin.String(),
		ToAddress:   receiver.String(),
		Amount:      sdk.NewCoins(sdk.NewCoin(chain.NetworkConfig.BaseDenom, sdk.NewInt(1000))),
	}

	minGasExpected := chain.GasLimitByMsgs(&banktypes.MsgSend{}, &banktypes.MsgSend{})
	maxGasExpected := minGasExpected * 10

	gasPrice, err := tx.GetGasPrice(ctx, chain.ClientContext)
	require.NoError(t, err)

	clientCtx = chain.ClientContext.WithFromAddress(admin)
	txf = chain.TxFactory().
		WithGas(maxGasExpected).
		WithGasPrices(gasPrice.String())
	result, err := tx.BroadcastTx(ctx, clientCtx, txf, wasmBankSend, bankSend)
	require.NoError(t, err)

	require.NoError(t, err)
	assert.Greater(t, uint64(result.GasUsed), minGasExpected)
	assert.Less(t, uint64(result.GasUsed), maxGasExpected)
}
