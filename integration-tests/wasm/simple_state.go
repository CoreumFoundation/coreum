package wasm

import (
	"context"
	_ "embed"
	"encoding/json"
	"math/big"

	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/types"
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

	initTestState := func(ctx context.Context) error {
		// FIXME (wojtek): Temporary code for transition
		if chain.Fund != nil {
			chain.Fund(adminWallet, types.NewCoinUnsafe(big.NewInt(100000), chain.Network.TokenSymbol()))
		}
		return nil
	}

	runTestFunc := func(ctx context.Context, t testing.T) {
		requireT := require.New(t)
		networkConfig := wasm.ChainConfig{
			MinGasPrice: types.NewCoinUnsafe(chain.Network.FeeModel().InitialGasPrice.BigInt(), nativeDenom),
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
		}, simpleStateWASM, requireT)

		// Query the contract state to get the initial count
		queryOut, err := wasm.Query(ctx, deployOut.ContractAddr, wasm.QueryConfig{
			Network:      networkConfig,
			QueryPayload: `{"get_count": {}}`,
		})
		requireT.NoError(err)

		var response simpleStateQueryResponse
		err = json.Unmarshal(queryOut.Result, &response)
		requireT.NoError(err)
		requireT.Equal(1337, response.Count)

		// Execute contract to increment the count
		execOut, err := wasm.Execute(ctx, deployOut.ContractAddr, wasm.ExecuteConfig{
			Network:        networkConfig,
			From:           adminWallet,
			ExecutePayload: `{"increment": {}}`,
		})
		requireT.NoError(err)
		requireT.NotEmpty(execOut.ExecuteTxHash)
		requireT.Equal(deployOut.ContractAddr, execOut.ContractAddress)
		requireT.Equal("try_increment", execOut.MethodExecuted)

		// Query the contract once again to ensure the count has been incremented
		queryOut, err = wasm.Query(ctx, deployOut.ContractAddr, wasm.QueryConfig{
			Network:      networkConfig,
			QueryPayload: `{"get_count": {}}`,
		})
		requireT.NoError(err)

		err = json.Unmarshal(queryOut.Result, &response)
		requireT.NoError(err)
		requireT.Equal(1338, response.Count)
	}

	return initTestState, runTestFunc
}

type simpleStateQueryResponse struct {
	Count int `json:"count"`
}
