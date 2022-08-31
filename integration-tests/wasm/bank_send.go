package wasm

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	"math/big"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/types"
	"github.com/CoreumFoundation/coreum/pkg/wasm"
)

var (
	//go:embed contracts/bank-send/artifacts/bank_send.wasm
	bankSendWASM []byte
)

type bankInstantiatePayload struct {
	Count int `json:"count"`
}

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
		networkConfig := wasm.ChainConfig{
			GasPrice: types.NewCoinUnsafe(chain.Network.FeeModel().InitialGasPrice.BigInt(), nativeDenom),
			Client:   chain.Client,
		}

		initialPayload, err := json.Marshal(bankInstantiatePayload{
			Count: 0,
		})
		requireT.NoError(err)
		deployOut := deployWasmContract(ctx, wasm.DeployConfig{
			Network: networkConfig,
			From:    adminWallet,
			InstantiationConfig: wasm.ContractInstanceConfig{
				NeedInstantiation:  true,
				InstantiatePayload: initialPayload,
				// transfer some coins during instantiation, so we could withdraw them later using contract code.
				Amount: types.NewCoinUnsafe(big.NewInt(10000), nativeDenom),
			},
		}, bankSendWASM, requireT)

		contractBalance, err := chain.Client.BankQueryClient().Balance(ctx,
			&banktypes.QueryBalanceRequest{
				Address: deployOut.ContractAddr,
				Denom:   nativeDenom,
			})
		requireT.NoError(err)
		requireT.NotNil(contractBalance.Balance)
		requireT.Equal(nativeDenom, contractBalance.Balance.Denom)
		requireT.Equal("10000", contractBalance.Balance.Amount.String())

		testWallet := testing.RandomWallet()
		withdrawMsg := fmt.Sprintf(
			`{"withdraw": { "amount":"5000", "denom":"%s", "recipient":"%s" }}`,
			nativeDenom,
			testWallet.Address().String(),
		)

		execOut, err := wasm.Execute(ctx, deployOut.ContractAddr, wasm.ExecuteConfig{
			// withdraw half of the coins to a test wallet, previously empty
			Network:        networkConfig,
			From:           adminWallet,
			ExecutePayload: withdrawMsg,
		})
		requireT.NoError(err)
		requireT.NotEmpty(execOut.ExecuteTxHash)
		requireT.Equal(deployOut.ContractAddr, execOut.ContractAddress)
		requireT.Equal("try_withdraw", execOut.MethodExecuted)

		// check that contract now has half of the coins
		contractBalance, err = chain.Client.BankQueryClient().Balance(ctx,
			&banktypes.QueryBalanceRequest{
				Address: deployOut.ContractAddr,
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

func deployWasmContract(
	ctx context.Context,
	config wasm.DeployConfig,
	contractData []byte,
	requireT *require.Assertions,
) *wasm.DeployOutput {
	config.Contract = contractData
	deployOut, err := wasm.Deploy(ctx, config)

	requireT.NoError(err)
	requireT.NotEmpty(deployOut.StoreTxHash)
	requireT.NotEmpty(deployOut.ContractAddr)

	return deployOut
}
