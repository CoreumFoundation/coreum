package wasm

import (
	"context"
	_ "embed"
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
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

// TestPinningAndUnpinningSmartContractUsingGovernance deploys simple smart contract, verifies that it works properly and then tests that
// pinning and unpinning through proposals works correctly. We also verify that pinned smart contract consumes less gas.
func TestPinningAndUnpinningSmartContractUsingGovernance(ctx context.Context, t testing.T, chain testing.Chain) {
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
	gasUsedBeforePinning := incrementAndVerify(ctx, clientCtx, txf, contractAddr, requireT, 1338)

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

	gasUsedAfterPinning := incrementAndVerify(ctx, clientCtx, txf, contractAddr, requireT, 1339)

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

	gasUsedAfterUnpinning := incrementAndVerify(ctx, clientCtx, txf, contractAddr, requireT, 1340)

	logger.Get(ctx).Info("Gas saved on poinned contract",
		zap.Int64("gasBeforePinning", gasUsedBeforePinning),
		zap.Int64("gasAfterPinning", gasUsedAfterPinning))

	assertT := assert.New(t)
	assertT.Less(gasUsedAfterPinning, gasUsedBeforePinning)
	assertT.Greater(gasUsedAfterUnpinning, gasUsedAfterPinning)
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
) int64 {
	// execute contract to increment the count
	incrementPayload, err := methodToEmptyBodyPayload(increment)
	requireT.NoError(err)
	gasUsed, err := Execute(ctx, clientCtx, txf, contractAddr, incrementPayload, sdk.Coin{})
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

	return gasUsed
}
