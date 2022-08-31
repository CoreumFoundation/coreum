package wasm

import (
	"context"
	_ "embed"
	"encoding/json"
	"math/big"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

var (
	//go:embed contracts/bank-send/artifacts/bank_send.wasm
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
func TestBankSendWasmContract(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
	adminWallet := testing.RandomWallet()
	nativeDenom := chain.Network.TokenSymbol()

	initTestState := func(ctx context.Context) error {
		// FIXME (wojtek): Temporary code for transition
		if chain.Fund != nil {
			chain.Fund(adminWallet, types.NewCoinUnsafe(big.NewInt(5000000000), chain.Network.TokenSymbol()))
		}
		return nil
	}

	runTestFunc := func(ctx context.Context, t testing.T) {
		requireT := require.New(t)
		wasmTestClient := newWasmTestClient(tx.BaseInput{
			Signer:   adminWallet,
			GasPrice: types.NewCoinUnsafe(chain.Network.FeeModel().InitialGasPrice.BigInt(), nativeDenom),
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
				amount: types.NewCoinUnsafe(big.NewInt(10000), nativeDenom),
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
		requireT.Equal(nativeDenom, contractBalance.Balance.Denom)
		requireT.Equal("10000", contractBalance.Balance.Amount.String())

		testWallet := testing.RandomWallet()
		withdrawPayload, err := json.Marshal(map[bankMethod]bankWithdrawRequest{
			withdraw: {
				Amount:    "5000",
				Denom:     nativeDenom,
				Recipient: testWallet.Address().String(),
			},
		})

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
		requireT.Equal(nativeDenom, contractBalance.Balance.Denom)
		requireT.Equal("5000", contractBalance.Balance.Amount.String())

		// check that the target test wallet has another half
		testWalletBalance, err := chain.Client.BankQueryClient().Balance(ctx,
			&banktypes.QueryBalanceRequest{
				Address: testWallet.Address().String(),
				Denom:   nativeDenom,
			})
		requireT.NoError(err)
		requireT.NotNil(testWalletBalance.Balance)
		requireT.Equal(nativeDenom, testWalletBalance.Balance.Denom)
		requireT.Equal("5000", testWalletBalance.Balance.Amount.String())
		// bank send invoked by the contract code succeeded! 〠
	}

	return initTestState, runTestFunc
}
