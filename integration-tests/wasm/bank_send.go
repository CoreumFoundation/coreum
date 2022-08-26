package wasm

import (
	"context"
	_ "embed"
	"fmt"
	"math/big"
	"os"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/types"
	"github.com/CoreumFoundation/coreum/pkg/wasm"
)

var (
	//go:embed contracts/bank-send/artifacts/bank_send.wasm
	bankSendWASM []byte
)

// TestBankSendWasmContract runs a contract deployment flow and tests that the contract is able to use Bank module
// to disperse the native coins.
func TestBankSendWasmContract(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
	adminWallet := testing.RandomWallet()
	nativeDenom := chain.Network.TokenSymbol()
	nativeTokens := func(v string) string {
		return v + nativeDenom
	}

	initTestState := func(ctx context.Context) error {
		// FIXME (wojtek): Temporary code for transition
		if chain.Fund != nil {
			if err := fundDeployerAcc(chain, adminWallet); err != nil {
				return err
			}
		}
		return nil
	}

	runTestFunc := func(ctx context.Context, t testing.T) {
		expect := require.New(t)
		networkConfig := wasm.ChainConfig{
			MinGasPrice: nativeTokens(chain.Network.InitialGasPrice().String()),
			Client:      chain.Client,
		}

		deployOut := deployWasmContract(ctx, wasm.DeployConfig{
			Network: networkConfig,
			From:    adminWallet,
			InstantiationConfig: wasm.ContractInstanceConfig{
				NeedInstantiation:  true,
				InstantiatePayload: `{"count": 0}`,
				// transfer some coins during instantiation, so we could withdraw them later using contract code.
				Amount: nativeTokens("10000"),
			},
		}, bankSendWASM, expect)

		contractBalance, err := chain.Client.BankQueryClient().Balance(ctx,
			&banktypes.QueryBalanceRequest{
				Address: deployOut.ContractAddr,
				Denom:   nativeDenom,
			})
		expect.NoError(err)
		expect.NotNil(contractBalance.Balance)
		expect.Equal(nativeDenom, contractBalance.Balance.Denom)
		expect.Equal("10000", contractBalance.Balance.Amount.String())

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
		expect.NoError(err)
		expect.NotEmpty(execOut.ExecuteTxHash)
		expect.Equal(deployOut.ContractAddr, execOut.ContractAddress)
		expect.Equal("try_withdraw", execOut.MethodExecuted)

		// check that contract now has half of the coins
		contractBalance, err = chain.Client.BankQueryClient().Balance(ctx,
			&banktypes.QueryBalanceRequest{
				Address: deployOut.ContractAddr,
				Denom:   nativeDenom,
			})
		expect.NoError(err)
		expect.NotNil(contractBalance.Balance)
		expect.Equal(nativeDenom, contractBalance.Balance.Denom)
		expect.Equal("5000", contractBalance.Balance.Amount.String())

		// check that the target test wallet has another half
		testWalletBalance, err := chain.Client.BankQueryClient().Balance(ctx,
			&banktypes.QueryBalanceRequest{
				Address: testWallet.Address().String(),
				Denom:   nativeDenom,
			})
		expect.NoError(err)
		expect.NotNil(testWalletBalance.Balance)
		expect.Equal(nativeDenom, testWalletBalance.Balance.Denom)
		expect.Equal("5000", testWalletBalance.Balance.Amount.String())
		// bank send invoked by the contract code succeeded! 〠
	}

	return initTestState, runTestFunc
}

func fundDeployerAcc(chain testing.Chain, wallet types.Wallet) error {
	bv, ok := big.NewInt(0).SetString("1000000000000", 10)
	if !ok {
		panic("invalid amount")
	}
	balance, err := types.NewCoin(bv, chain.Network.TokenSymbol())
	if err != nil {
		return err
	}
	chain.Fund(wallet, balance)
	return nil
}

func deployWasmContract(
	ctx context.Context,
	config wasm.DeployConfig,
	contractData []byte,
	expect *require.Assertions,
) *wasm.DeployOutput {
	wasmFile, err := os.CreateTemp("", "test_contract.wasm")
	expect.NoError(err)

	_, err = wasmFile.Write(contractData)
	expect.NoError(err)

	config.ArtefactPath = wasmFile.Name()
	deployOut, err := wasm.Deploy(ctx, config)

	expect.NoError(err)
	expect.NotEmpty(deployOut.StoreTxHash)
	expect.NotEmpty(deployOut.ContractAddr)

	return deployOut
}
