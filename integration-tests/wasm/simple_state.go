package wasm

import (
	"context"
	_ "embed"
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

var (
	//go:embed testdata/simple-state/artifacts/simple_state.wasm
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
func TestSimpleStateWasmContract(ctx context.Context, t testing.T, chain testing.Chain) {
	wallet := testing.RandomWallet()
	nativeDenom := chain.NetworkConfig.TokenSymbol

	requireT := require.New(t)
	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		testing.FundedAccount{
			Wallet: wallet,
			Amount: testing.MustNewCoin(t, sdk.NewInt(5000000000), nativeDenom),
		},
	))

	wasmTestClient := newWasmTestClient(tx.BaseInput{
		Signer:   wallet,
		GasPrice: testing.MustNewCoin(t, chain.NetworkConfig.Fee.FeeModel.InitialGasPrice, nativeDenom),
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
	getCountPayload, err := methodToEmptyBodyPayload(getCount)
	requireT.NoError(err)
	queryOut, err := wasmTestClient.query(ctx, contractAddr, getCountPayload)
	requireT.NoError(err)

	var response simpleState
	err = json.Unmarshal(queryOut, &response)
	requireT.NoError(err)
	requireT.Equal(1337, response.Count)

	// execute contract to increment the count
	incrementPayload, err := methodToEmptyBodyPayload(increment)
	requireT.NoError(err)
	err = wasmTestClient.execute(ctx, contractAddr, incrementPayload, types.Coin{})
	requireT.NoError(err)

	// Query the contract once again to ensure the count has been incremented
	queryOut, err = wasmTestClient.query(ctx, contractAddr, getCountPayload)
	requireT.NoError(err)

	err = json.Unmarshal(queryOut, &response)
	requireT.NoError(err)
	requireT.Equal(1338, response.Count)
}

func methodToEmptyBodyPayload(methodName simpleStateMethod) (json.RawMessage, error) {
	return json.Marshal(map[simpleStateMethod]struct{}{
		methodName: {},
	})
}
