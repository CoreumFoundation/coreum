package wasm

import (
	"context"
	_ "embed"
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
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
func TestBankSendWasmContract(ctx context.Context, t testing.T, chain testing.Chain) {
	adminWallet := testing.RandomWallet()
	nativeDenom := chain.NetworkConfig.TokenSymbol

	requireT := require.New(t)
	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		testing.FundedAccount{
			Wallet: adminWallet,
			Amount: testing.MustNewCoin(t, sdk.NewInt(5000000000), nativeDenom),
		},
	))

	wasmTestClient := newWasmTestClient(tx.BaseInput{
		Signer:   adminWallet,
		GasPrice: testing.MustNewCoin(t, chain.NetworkConfig.Fee.FeeModel.InitialGasPrice, nativeDenom),
	}, chain.Client)

	initialPayload, err := json.Marshal(bankInstantiatePayload{
		Count: 0,
	})
	requireT.NoError(err)
	contractAddr, err := wasmTestClient.deployAndInstantiate(
		ctx,
		bankSendWASM,
		instantiateConfig{
			accessType: wasmtypes.AccessTypeUnspecified,
			payload:    initialPayload,
			// transfer some coins during instantiation, so we could withdraw them later using contract code.
			amount: testing.MustNewCoin(t, sdk.NewInt(10000), nativeDenom),
			label:  "bank_send",
		},
	)
	requireT.NoError(err)

	contractBalance, err := chain.Client.BankQueryClient().Balance(ctx,
		&banktypes.QueryBalanceRequest{
			Address: contractAddr,
			Denom:   nativeDenom,
		})
	requireT.NoError(err)
	requireT.NotNil(contractBalance.Balance)
	requireT.Equal(sdk.NewInt64Coin(nativeDenom, 10000).String(), contractBalance.Balance.String())

	testWallet := testing.RandomWallet()
	withdrawPayload, err := json.Marshal(map[bankMethod]bankWithdrawRequest{
		withdraw: {
			Amount:    "5000",
			Denom:     nativeDenom,
			Recipient: testWallet.Address().String(),
		},
	})
	requireT.NoError(err)

	err = wasmTestClient.execute(ctx, contractAddr, withdrawPayload, types.Coin{})
	requireT.NoError(err)

	// check that contract now has half of the coins
	contractBalance, err = chain.Client.BankQueryClient().Balance(ctx,
		&banktypes.QueryBalanceRequest{
			Address: contractAddr,
			Denom:   nativeDenom,
		})
	requireT.NoError(err)
	requireT.NotNil(contractBalance.Balance)
	requireT.Equal(sdk.NewInt64Coin(nativeDenom, 5000).String(), contractBalance.Balance.String())

	// check that the target test wallet has another half
	testWalletBalance, err := chain.Client.BankQueryClient().Balance(ctx,
		&banktypes.QueryBalanceRequest{
			Address: testWallet.Address().String(),
			Denom:   nativeDenom,
		})
	requireT.NoError(err)
	requireT.NotNil(testWalletBalance.Balance)
	requireT.Equal(sdk.NewInt64Coin(nativeDenom, 5000).String(), contractBalance.Balance.String())
}
