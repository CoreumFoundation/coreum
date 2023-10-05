//go:build integrationtests

package upgrade

import (
	"encoding/json"
	"testing"

	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	moduleswasm "github.com/CoreumFoundation/coreum/v3/integration-tests/contracts/modules"
	"github.com/CoreumFoundation/coreum/v3/testutil/integration"
)

type wasmMigrationTest struct {
	contractAddress        string
	adminAccAddress        sdk.AccAddress
	counterBeforeMigration int
}

func (wmt *wasmMigrationTest) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	admin := chain.GenAccount()

	chain.Faucet.FundAccounts(ctx, t,
		integration.NewFundedAccount(admin, chain.NewCoin(sdkmath.NewInt(5000000))),
	)

	// instantiateWASMContract the contract and set the initial counter state.
	initialCount := 2349
	initialPayload, err := json.Marshal(moduleswasm.SimpleState{
		Count: initialCount,
	})
	requireT.NoError(err)

	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		admin,
		moduleswasm.SimpleStateWASM,
		integration.InstantiateConfig{
			Admin:      admin,
			AccessType: wasmtypes.AccessTypeUnspecified,
			Payload:    initialPayload,
			Label:      "simple_state",
		},
	)
	requireT.NoError(err)

	// get the current counter
	getCountPayload, err := moduleswasm.MethodToEmptyBodyPayload(moduleswasm.SimpleGetCount)
	requireT.NoError(err)
	queryOut, err := chain.Wasm.QueryWASMContract(ctx, contractAddr, getCountPayload)
	requireT.NoError(err)
	var response moduleswasm.SimpleState
	requireT.NoError(json.Unmarshal(queryOut, &response))
	requireT.Equal(initialCount, response.Count)

	moduleswasm.IncrementSimpleStateAndVerify(ctx, txf, admin, chain, contractAddr, requireT, initialCount+1)

	wmt.contractAddress = contractAddr
	wmt.adminAccAddress = admin
	wmt.counterBeforeMigration = initialCount + 1
}

func (wmt *wasmMigrationTest) After(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

	// get the current counter and verify it's the same as before migration
	getCountPayload, err := moduleswasm.MethodToEmptyBodyPayload(moduleswasm.SimpleGetCount)
	requireT.NoError(err)
	queryOut, err := chain.Wasm.QueryWASMContract(ctx, wmt.contractAddress, getCountPayload)
	requireT.NoError(err)
	var response moduleswasm.SimpleState
	requireT.NoError(json.Unmarshal(queryOut, &response))
	requireT.Equal(wmt.counterBeforeMigration, response.Count)

	moduleswasm.IncrementSimpleStateAndVerify(ctx, txf, wmt.adminAccAddress, chain, wmt.contractAddress, requireT, wmt.counterBeforeMigration+1)
}
