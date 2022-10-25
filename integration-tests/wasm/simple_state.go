package wasm

import (
	"context"
	_ "embed"
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
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
	admin := chain.GenAccount()
	proposer := chain.GenAccount()

	requireT := require.New(t)

	proposerBalance, err := chain.Governance.ComputeProposerBalance(ctx)
	requireT.NoError(err)
	proposerBalance.Amount = proposerBalance.Amount.MulRaw(2)

	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		testing.NewFundedAccount(admin, chain.NewCoin(sdk.NewInt(5000000000))),
		testing.NewFundedAccount(proposer, proposerBalance),
	))

	// instantiate the contract and set the initial counter state.
	initialPayload, err := json.Marshal(simpleState{
		Count: 1337,
	})
	requireT.NoError(err)

	clientCtx := chain.ClientContext.WithFromAddress(admin)
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)

	contractAddr, codeID, err := DeployAndInstantiate(
		ctx,
		clientCtx,
		txf,
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
	queryOut, err := Query(ctx, clientCtx, contractAddr, getCountPayload)
	requireT.NoError(err)
	var response simpleState
	err = json.Unmarshal(queryOut, &response)
	requireT.NoError(err)
	requireT.Equal(1337, response.Count)

	// execute contract to increment the count
	incrementAndVerify(ctx, clientCtx, txf, contractAddr, requireT, 1338)

	// verify that smart contract is not pinned
	requireT.False(IsPinned(ctx, clientCtx, codeID))

	// pin smart contract
	proposalID, err := chain.Governance.Propose(ctx, proposer, &wasmtypes.PinCodesProposal{
		Title:       "Pin smart contract",
		Description: "Testing smart contract pinning",
		CodeIDs:     []uint64{codeID},
	})
	requireT.NoError(err)

	proposal, err := chain.Governance.GetProposal(ctx, uint64(proposalID))
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusVotingPeriod, proposal.Status)

	err = chain.Governance.VoteAll(ctx, govtypes.OptionYes, proposal.ProposalId)
	requireT.NoError(err)
	requireT.NoError(chain.Governance.WaitForVotingToPass(ctx, uint64(proposalID)))

	requireT.True(IsPinned(ctx, clientCtx, codeID))

	incrementAndVerify(ctx, clientCtx, txf, contractAddr, requireT, 1339)

	// unpin smart contract
	proposalID, err = chain.Governance.Propose(ctx, proposer, &wasmtypes.UnpinCodesProposal{
		Title:       "Unpin smart contract",
		Description: "Testing smart contract unpinning",
		CodeIDs:     []uint64{codeID},
	})
	requireT.NoError(err)

	proposal, err = chain.Governance.GetProposal(ctx, uint64(proposalID))
	requireT.NoError(err)
	requireT.Equal(govtypes.StatusVotingPeriod, proposal.Status)

	err = chain.Governance.VoteAll(ctx, govtypes.OptionYes, proposal.ProposalId)
	requireT.NoError(err)
	requireT.NoError(chain.Governance.WaitForVotingToPass(ctx, uint64(proposalID)))

	requireT.False(IsPinned(ctx, clientCtx, codeID))

	incrementAndVerify(ctx, clientCtx, txf, contractAddr, requireT, 1340)
}

func methodToEmptyBodyPayload(methodName simpleStateMethod) (json.RawMessage, error) {
	return json.Marshal(map[simpleStateMethod]struct{}{
		methodName: {},
	})
}

func incrementAndVerify(
	ctx context.Context,
	clientCtx tx.ClientContext,
	txf tx.Factory,
	contractAddr string,
	requireT *require.Assertions,
	expectedValue int,
) {
	// execute contract to increment the count
	incrementPayload, err := methodToEmptyBodyPayload(increment)
	requireT.NoError(err)
	err = Execute(ctx, clientCtx, txf, contractAddr, incrementPayload, sdk.Coin{})
	requireT.NoError(err)

	// check the update count
	getCountPayload, err := methodToEmptyBodyPayload(getCount)
	requireT.NoError(err)
	queryOut, err := Query(ctx, clientCtx, contractAddr, getCountPayload)
	requireT.NoError(err)

	var response simpleState
	err = json.Unmarshal(queryOut, &response)
	requireT.NoError(err)

	requireT.Equal(expectedValue, response.Count)
}
