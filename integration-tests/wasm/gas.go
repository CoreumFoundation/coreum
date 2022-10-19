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
	admin := chain.GenAccount()

	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(admin, chain.NewCoin(sdk.NewInt(5000000000))),
	))

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
			Denom:     chain.NetworkConfig.TokenSymbol,
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
		Amount:      sdk.NewCoins(sdk.NewCoin(chain.NetworkConfig.TokenSymbol, sdk.NewInt(1000))),
	}

	minGasExpected := chain.GasLimitByMsgs(&banktypes.MsgSend{}, &banktypes.MsgSend{})
	maxGasExpected := minGasExpected * 10

	clientCtx = chain.ChainContext.ClientContext.WithFromAddress(admin)
	txf = chain.ChainContext.TxFactory().WithGas(maxGasExpected)
	result, err := tx.BroadcastTx(ctx, clientCtx, txf, wasmBankSend, bankSend)
	require.NoError(t, err)

	require.NoError(t, err)
	assert.Greater(t, uint64(result.GasUsed), minGasExpected)
	assert.Less(t, uint64(result.GasUsed), maxGasExpected)
}
