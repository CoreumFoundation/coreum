package modules

import (
	"context"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v3/pkg/client"
	integrationtests "github.com/CoreumFoundation/coreum/v3/testutil/integration"
)

// SimpleState is a structure used to initizlize the simple state contract
// and also represents response returned from get_count query.
type SimpleState struct {
	Count int `json:"count"`
}

// SimpleStateMethod is a type used to represent the methods available inside the simple state contract.
type SimpleStateMethod string

const (
	// SimpleGetCount is a method used to get the current count.
	SimpleGetCount SimpleStateMethod = "get_count"
	// SimpleIncrement is a method used to increment the current count.
	SimpleIncrement SimpleStateMethod = "increment"
)

// IncrementSimpleStateAndVerify is a helper function used to increment the count inside the simple state contract
// and verify if number is equal to expected value.
func IncrementSimpleStateAndVerify(
	ctx context.Context,
	txf client.Factory,
	fromAddress sdk.AccAddress,
	chain integrationtests.CoreumChain,
	contractAddr string,
	requireT *require.Assertions,
	expectedValue int,
) int64 {
	// execute contract to increment the count
	incrementPayload, err := MethodToEmptyBodyPayload(SimpleIncrement)
	requireT.NoError(err)
	gasUsed, err := chain.Wasm.ExecuteWASMContract(ctx, txf, fromAddress, contractAddr, incrementPayload, sdk.Coin{})
	requireT.NoError(err)

	// check the update count
	getCountPayload, err := MethodToEmptyBodyPayload(SimpleGetCount)
	requireT.NoError(err)
	queryOut, err := chain.Wasm.QueryWASMContract(ctx, contractAddr, getCountPayload)
	requireT.NoError(err)

	var response SimpleState
	requireT.NoError(json.Unmarshal(queryOut, &response))
	requireT.Equal(expectedValue, response.Count)

	return gasUsed
}

// MethodToEmptyBodyPayload is a helper function used to create a payload for the given method and empty args.
func MethodToEmptyBodyPayload(methodName SimpleStateMethod) (json.RawMessage, error) {
	return json.Marshal(map[SimpleStateMethod]struct{}{
		methodName: {},
	})
}
