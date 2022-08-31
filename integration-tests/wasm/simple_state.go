package wasm

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/CoreumFoundation/coreum/pkg/tx"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

var (
	//go:embed contracts/simple-state/artifacts/simple_state.wasm
	simpleStateWASM []byte
)

type simpleState struct {
	Count int `json:"count"`
}

type simpleStateMethod string

const (
	getCount  simpleStateMethod = "get_count"
	increment simpleStateMethod = "increment"
)

// TestSimpleStateWasmContract runs a contract deployment flow and tries to modify the state after deployment.
// This is a E2E check for the WASM integration, to ensure it works for a simple state contract (Counter).
func TestSimpleStateWasmContract(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
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

		// instantiate the contract and set the initial counter state.
		// This step could be done within previous step, but separated there, so we could check
		// the intermediate result of code storage.
		initialPayload, err := json.Marshal(simpleState{
			Count: 1337,
		})
		requireT.NoError(err)
		contractAddr, err := wasmTestClient.deployAndInstantiate(
			ctx,
			simpleStateWASM,
			instantiateConfig{
				accessType: wasmtypes.AccessTypeUnspecified,
				payload:    initialPayload,
				label:      "simple_state",
			},
		)
		requireT.NoError(err)

		// Query the contract state to get the initial count
		queryOut, err := wasmTestClient.query(ctx, contractAddr, methodToEmptyBodyPayload(getCount))
		requireT.NoError(err)

		var response simpleState
		err = json.Unmarshal(queryOut, &response)
		requireT.NoError(err)
		requireT.Equal(1337, response.Count)

		// execute contract to increment the count
		err = wasmTestClient.execute(ctx, contractAddr, methodToEmptyBodyPayload(increment), types.Coin{})
		requireT.NoError(err)

		// Query the contract once again to ensure the count has been incremented
		queryOut, err = wasmTestClient.query(ctx, contractAddr, methodToEmptyBodyPayload(getCount))
		requireT.NoError(err)

		err = json.Unmarshal(queryOut, &response)
		requireT.NoError(err)
		requireT.Equal(1338, response.Count)
	}

	return initTestState, runTestFunc
}

func methodToEmptyBodyPayload(methodName simpleStateMethod) json.RawMessage {
	return json.RawMessage(fmt.Sprintf(`{"%s": {}}`, methodName))
}
