package wasm

import (
	"context"
	_ "embed"
	"encoding/json"

	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/wasm"
)

var (
	//go:embed contracts/simple-state/artifacts/simple_state.wasm
	simpleStateWASM []byte
)

// TestSimpleStateWasmContract runs a contract deployment flow and tries to modify the state after deployment.
// This is a E2E check for the WASM integration, to ensure it works for a simple state contract (Counter).
func TestSimpleStateWasmContract(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
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

		// Instantiate the contract and set the initial counter state.
		// This step could be done within previous step, but separated there so we could check
		// the intermediate result of code storage.
		deployOut := deployWasmContract(ctx, wasm.DeployConfig{
			Network: networkConfig,
			From:    adminWallet,
			InstantiationConfig: wasm.ContractInstanceConfig{
				NeedInstantiation:  true,
				InstantiatePayload: `{"count": 1337}`,
			},
		}, simpleStateWASM, expect)

		// Query the contract state to get the initial count
		queryOut, err := wasm.Query(ctx, deployOut.ContractAddr, wasm.QueryConfig{
			Network:      networkConfig,
			QueryPayload: `{"get_count": {}}`,
		})
		expect.NoError(err)

		var response simpleStateQueryResponse
		err = json.Unmarshal(queryOut.Result, &response)
		expect.NoError(err)
		expect.Equal(1337, response.Count)

		// Execute contract to increment the count
		execOut, err := wasm.Execute(ctx, deployOut.ContractAddr, wasm.ExecuteConfig{
			Network:        networkConfig,
			From:           adminWallet,
			ExecutePayload: `{"increment": {}}`,
		})
		expect.NoError(err)
		expect.NotEmpty(execOut.ExecuteTxHash)
		expect.Equal(deployOut.ContractAddr, execOut.ContractAddress)
		expect.Equal("try_increment", execOut.MethodExecuted)

		// Query the contract once again to ensure the count has been incremented
		queryOut, err = wasm.Query(ctx, deployOut.ContractAddr, wasm.QueryConfig{
			Network:      networkConfig,
			QueryPayload: `{"get_count": {}}`,
		})
		expect.NoError(err)

		err = json.Unmarshal(queryOut.Result, &response)
		expect.NoError(err)
		expect.Equal(1338, response.Count)
	}

	return initTestState, runTestFunc
}

type simpleStateQueryResponse struct {
	Count int `json:"count"`
}
