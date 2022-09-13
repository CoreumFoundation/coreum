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
		testing.NewFundedAccount(wallet, testing.MustNewCoin(t, sdk.NewInt(5000000000), nativeDenom)),
	))

	baseInput := tx.BaseInput{
		Signer:   wallet,
		GasPrice: testing.MustNewCoin(t, chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice, nativeDenom),
	}
	wasmTestClient := NewClient(chain.Client)

	// instantiate the contract and set the initial counter state.
	initialPayload, err := json.Marshal(simpleState{
		Count: 1337,
	})
	requireT.NoError(err)
	contractAddr, err := wasmTestClient.DeployAndInstantiate(
		ctx,
		baseInput,
		simpleStateWASM,
		InstantiateConfig{
			accessType: wasmtypes.AccessTypeUnspecified,
			payload:    initialPayload,
			label:      "simple_state",
		},
	)
	requireT.NoError(err)

	// get the current counter state
	getCountPayload, err := methodToEmptyBodyPayload(getCount)
	requireT.NoError(err)
	queryOut, err := wasmTestClient.Query(ctx, contractAddr, getCountPayload)
	requireT.NoError(err)
	var response simpleState
	err = json.Unmarshal(queryOut, &response)
	requireT.NoError(err)
	requireT.Equal(1337, response.Count)

	// execute contract to increment the count
	incrementPayload, err := methodToEmptyBodyPayload(increment)
	requireT.NoError(err)
	err = wasmTestClient.Execute(ctx, baseInput, contractAddr, incrementPayload, types.Coin{})
	requireT.NoError(err)

	// check the update count
	queryOut, err = wasmTestClient.Query(ctx, contractAddr, getCountPayload)
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
