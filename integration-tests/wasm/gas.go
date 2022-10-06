package wasm

import (
	"context"
	"encoding/hex"
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// TestGasWasmBankSendAndBankSend checks that a message containing a deterministic and a
// non-deterministic transaction takes gas within appropriate limits.
func TestGasWasmBankSendAndBankSend(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	adminAddress := chain.RandomWallet()
	nativeDenom := chain.NetworkConfig.TokenSymbol
	adminKeyInfo, err := chain.ChainContext.ClientContext.Keyring().KeyByAddress(adminAddress)
	requireT.NoError(err)
	unsafeKeyRing := keyring.NewUnsafe(chain.ChainContext.ClientContext.Keyring())
	adminPrivateKeyHex, err := unsafeKeyRing.UnsafeExportPrivKeyHex(adminKeyInfo.GetName())
	requireT.NoError(err)
	adminPrivateKey, err := hex.DecodeString(adminPrivateKeyHex)
	requireT.NoError(err)

	adminWallet := types.Wallet{Name: adminAddress.String(), Key: adminPrivateKey}

	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(adminWallet.Address(), chain.NewCoin(sdk.NewInt(5000000000))),
	))

	gasPrice := chain.NewDecCoin(chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice)
	baseInput := tx.BaseInput{
		Signer:   adminWallet,
		GasPrice: gasPrice,
	}
	coredClient := chain.Client
	wasmTestClient := NewClient(coredClient)

	// deploy and init contract with the initial coins amount
	initialPayload, err := json.Marshal(bankInstantiatePayload{Count: 0})
	requireT.NoError(err)
	contractAddr, err := wasmTestClient.DeployAndInstantiate(
		ctx,
		baseInput,
		bankSendWASM,
		InstantiateConfig{
			accessType: wasmtypes.AccessTypeUnspecified,
			payload:    initialPayload,
			amount:     chain.NewCoin(sdk.NewInt(10000)),
			label:      "bank_send",
		},
	)
	requireT.NoError(err)

	// send additional coins to contract directly
	sdkContractAddress, err := sdk.AccAddressFromBech32(contractAddr)
	requireT.NoError(err)
	err = chain.Faucet.FundAccounts(
		ctx,
		testing.NewFundedAccount(sdkContractAddress, sdk.NewInt64Coin(nativeDenom, 5000)),
	)
	requireT.NoError(err)

	receiver := chain.RandomWallet()
	// send coin from the contract to test wallet
	withdrawPayload, err := json.Marshal(map[bankMethod]bankWithdrawRequest{
		withdraw: {
			Amount:    "5000",
			Denom:     nativeDenom,
			Recipient: receiver.String(),
		},
	})
	requireT.NoError(err)

	wasmBankSend := &wasmtypes.MsgExecuteContract{
		Sender:   adminAddress.String(),
		Contract: contractAddr,
		Msg:      wasmtypes.RawContractMessage(withdrawPayload),
		Funds:    sdk.Coins{},
	}

	bankSend := &banktypes.MsgSend{
		FromAddress: adminAddress.String(),
		ToAddress:   receiver.String(),
		Amount:      []sdk.Coin{sdk.NewCoin(chain.NetworkConfig.TokenSymbol, sdk.NewInt(1000))},
	}

	minGasExpected := chain.GasLimitByMsgs(&banktypes.MsgSend{}, &banktypes.MsgSend{})
	maxGasExpected := minGasExpected * 10

	clientCtx := chain.ChainContext.ClientContext.WithFromName(adminKeyInfo.GetName()).WithFromAddress(adminAddress)
	txf := chain.ChainContext.TxFactory().WithGas(maxGasExpected)
	result, err := tx.BroadcastTx(ctx, clientCtx, txf, wasmBankSend, bankSend)
	require.NoError(t, err)

	require.NoError(t, err)
	assert.Greater(t, uint64(result.GasUsed), minGasExpected)
	assert.Less(t, uint64(result.GasUsed), maxGasExpected)
}
