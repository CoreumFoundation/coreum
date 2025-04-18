//go:build integrationtests

package modules

import (
	"testing"

	tenderminttypes "github.com/cometbft/cometbft/proto/tendermint/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v6/integration-tests"
	"github.com/CoreumFoundation/coreum/v6/testutil/integration"
)

func TestUpdatingMaxBlockSize(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	consensusClient := consensustypes.NewQueryClient(chain.ClientContext)
	consensusParams, err := consensusClient.Params(ctx, &consensustypes.QueryParamsRequest{})
	requireT.NoError(err)
	consensusParams.Params.Block.MaxBytes--

	gov := chain.Governance

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := gov.ComputeProposerBalance(ctx, false)
	requireT.NoError(err)
	chain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: proposer,
			Amount:  proposerBalance,
		},
	)

	// Propose new block size.
	proposalMsg, err := gov.NewMsgSubmitProposal(
		ctx,
		proposer,
		[]sdk.Msg{&consensustypes.MsgUpdateParams{
			Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			Block:     consensusParams.Params.Block,
			Evidence:  consensusParams.Params.Evidence,
			Validator: consensusParams.Params.Validator,
			Abci:      consensusParams.Params.Abci,
		}},
		"",
		"Reduce block size",
		"Reduce block size",
		false,
	)
	requireT.NoError(err)
	proposalID, err := gov.Propose(ctx, t, proposalMsg)
	requireT.NoError(err)

	t.Logf("Proposal created, proposalID: %d", proposalID)

	// Vote by all staker accounts:
	// NoWithVeto 70% & No,Yes,Abstain 10% each.
	requireT.NoError(gov.VoteAll(ctx, govtypesv1.OptionYes, proposalID))

	// Wait for proposal result.
	finalStatus, err := gov.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusPassed, finalStatus)

	// Verify new consensus params.
	consensusParams.Params.Abci = &tenderminttypes.ABCIParams{}
	newConsensusParams, err := consensusClient.Params(ctx, &consensustypes.QueryParamsRequest{})
	requireT.NoError(err)
	requireT.Equal(consensusParams.Params, newConsensusParams.Params)
}
