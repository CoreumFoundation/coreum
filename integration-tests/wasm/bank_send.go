package wasm

import (
	"context"
	_ "embed"
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

var (
	//go:embed testdata/bank-send/artifacts/bank_send.wasm
	bankSendWASM []byte
)

type bankInstantiatePayload struct {
	Count int `json:"count"`
}

type bankWithdrawRequest struct {
	Amount    string `json:"amount"`
	Denom     string `json:"denom"`
	Recipient string `json:"recipient"`
}

type bankMethod string

const (
	withdraw bankMethod = "withdraw"
)

// TestBankSendWasmContract runs a contract deployment flow and tests that the contract is able to use Bank module
// to disperse the native coins.
func TestBankSendWasmContract(ctx context.Context, t testing.T, chain testing.Chain) { //nolint:funlen // The test covers step-by step use case, no need split it
	adminWallet := testing.RandomWallet()
	nativeDenom := chain.NetworkConfig.TokenSymbol

	requireT := require.New(t)
	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		testing.FundedAccount{
			Wallet: adminWallet,
			Amount: testing.MustNewCoin(t, sdk.NewInt(5000000000), nativeDenom),
		},
	))

	gasPrice := testing.MustNewCoin(t, chain.NetworkConfig.Fee.FeeModel.InitialGasPrice, nativeDenom)
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
			amount:     testing.MustNewCoin(t, sdk.NewInt(10000), nativeDenom),
			label:      "bank_send",
		},
	)
	requireT.NoError(err)

	// send additional coins to contract directly
	sdkContractAddress, err := sdk.AccAddressFromBech32(contractAddr)
	requireT.NoError(err)
	bakSendTx, err := coredClient.Sign(ctx, tx.BaseInput{
		Signer:   adminWallet,
		GasPrice: gasPrice,
		GasLimit: chain.NetworkConfig.Fee.DeterministicGas.BankSend,
	}, banktypes.NewMsgSend(
		adminWallet.Address(),
		sdkContractAddress,
		sdk.NewCoins(sdk.NewInt64Coin(nativeDenom, 5000)),
	))
	requireT.NoError(err)
	// TODO (dhil) replace to new Broadcast once we finish with it
	_, err = coredClient.Broadcast(ctx, coredClient.Encode(bakSendTx))
	requireT.NoError(err)

	// get the contract balance and check total
	contractBalance, err := coredClient.BankQueryClient().Balance(ctx,
		&banktypes.QueryBalanceRequest{
			Address: contractAddr,
			Denom:   nativeDenom,
		})
	requireT.NoError(err)
	requireT.NotNil(contractBalance.Balance)
	requireT.Equal(sdk.NewInt64Coin(nativeDenom, 15000).String(), contractBalance.Balance.String())

	testWallet := testing.RandomWallet()
	// try to exceed the contract limit
	withdrawPayload, err := json.Marshal(map[bankMethod]bankWithdrawRequest{
		withdraw: {
			Amount:    "16000",
			Denom:     nativeDenom,
			Recipient: testWallet.Address().String(),
		},
	})
	requireT.NoError(err)
	// we execute here with the constant gas fee, to check the cosmoserrors,
	// since if we don't set the gas cost will be estimated and the estimation func
	// will return the error which is impossible to convert to cosmoserrors to check the type.
	err = wasmTestClient.Execute(ctx, tx.BaseInput{
		Signer:   adminWallet,
		GasPrice: gasPrice,
		GasLimit: chain.NetworkConfig.Fee.DeterministicGas.BankSend,
	}, contractAddr, withdrawPayload, types.Coin{})
	requireT.Error(err)
	require.True(t, client.IsErr(err, cosmoserrors.ErrInsufficientFunds))

	// send coin from the contract to test wallet
	withdrawPayload, err = json.Marshal(map[bankMethod]bankWithdrawRequest{
		withdraw: {
			Amount:    "5000",
			Denom:     nativeDenom,
			Recipient: testWallet.Address().String(),
		},
	})
	requireT.NoError(err)
	err = wasmTestClient.Execute(ctx, baseInput, contractAddr, withdrawPayload, types.Coin{})
	requireT.NoError(err)

	// check contract and wallet balances
	contractBalance, err = coredClient.BankQueryClient().Balance(ctx,
		&banktypes.QueryBalanceRequest{
			Address: contractAddr,
			Denom:   nativeDenom,
		})
	requireT.NoError(err)
	requireT.NotNil(contractBalance.Balance)
	requireT.Equal(sdk.NewInt64Coin(nativeDenom, 10000).String(), contractBalance.Balance.String())

	testWalletBalance, err := coredClient.BankQueryClient().Balance(ctx,
		&banktypes.QueryBalanceRequest{
			Address: testWallet.Address().String(),
			Denom:   nativeDenom,
		})
	requireT.NoError(err)
	requireT.NotNil(testWalletBalance.Balance)
	requireT.Equal(sdk.NewInt64Coin(nativeDenom, 5000).String(), testWalletBalance.Balance.String())
}
