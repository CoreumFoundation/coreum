//go:build integrationtests

package modules

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v6/integration-tests"
	"github.com/CoreumFoundation/coreum/v6/pkg/client"
	"github.com/CoreumFoundation/coreum/v6/testutil/integration"
)

// TestGovProposalWithDepositAndWeightedVotes - is a complex governance test which tests:
// 1. proposal submission without enough deposit,
// 2. depositing missing amount to proposal created on the 1st step,
// 3. voting using weighted votes.
func TestGovProposalWithDepositAndWeightedVotes(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	gov := chain.Governance
	missingDepositAmount := chain.NewCoin(sdkmath.NewInt(10))

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := gov.ComputeProposerBalance(ctx, false)
	requireT.NoError(err)
	proposerBalance = proposerBalance.Sub(missingDepositAmount).Add(chain.NewCoin(sdkmath.NewInt(1)))
	chain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: proposer,
			Amount:  proposerBalance,
		},
	)

	// Create proposer depositor.
	depositor := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, depositor, integration.BalancesOptions{
		Messages: []sdk.Msg{&govtypesv1.MsgDeposit{}},
		Amount:   missingDepositAmount.Amount,
	})

	proposalMsg, err := gov.NewMsgSubmitProposal(
		ctx,
		proposer,
		[]sdk.Msg{&banktypes.MsgSend{
			FromAddress: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			ToAddress:   depositor.String(),
			Amount:      []sdk.Coin{chain.NewCoin(sdkmath.NewInt(1))},
		}},
		"",
		"Send some funds to depositor",
		"Send some funds to depositor",
		false,
	)
	requireT.NoError(err)
	proposalMsg.InitialDeposit = sdk.NewCoins(proposalMsg.InitialDeposit...).Sub(sdk.Coins{missingDepositAmount}...)
	proposalID, err := gov.Propose(ctx, t, proposalMsg)
	requireT.NoError(err)

	t.Logf("Proposal created, proposalID: %d", proposalID)

	// Verify that proposal is waiting for deposit.
	requirePropStatusFunc := func(expectedStatus govtypesv1.ProposalStatus) {
		proposal, err := gov.GetProposal(ctx, proposalID)
		requireT.NoError(err)
		requireT.Equal(expectedStatus, proposal.Status)
	}
	requirePropStatusFunc(govtypesv1.StatusDepositPeriod)

	// Deposit missing amount to proposal.
	depositMsg := govtypesv1.NewMsgDeposit(depositor, proposalID, sdk.Coins{missingDepositAmount})
	result, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(depositor),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(depositMsg)),
		depositMsg,
	)
	requireT.NoError(err)
	require.Equal(t, chain.GasLimitByMsgs(depositMsg), uint64(result.GasUsed))

	t.Logf("Deposited more funds to proposal, txHash:%s, gasUsed:%d", result.TxHash, result.GasUsed)

	// Verify that proposal voting has started.
	requirePropStatusFunc(govtypesv1.StatusVotingPeriod)

	// Store proposer and depositor balances before voting has finished.
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	accBalanceFunc := func(prop sdk.AccAddress) sdk.Coin {
		accBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
			Address: prop.String(),
			Denom:   chain.ChainSettings.Denom,
		})
		requireT.NoError(err)
		return *accBalance.Balance
	}
	proposerBalanceBeforeVoting := accBalanceFunc(proposer)
	depositorBalanceBeforeVoting := accBalanceFunc(depositor)

	// Vote by all staker accounts:
	// NoWithVeto 70% & No,Yes,Abstain 10% each.
	err = gov.VoteAllWeighted(ctx,
		govtypesv1.WeightedVoteOptions{
			&govtypesv1.WeightedVoteOption{
				Option: govtypesv1.OptionNoWithVeto,
				Weight: "0.7",
			},
			&govtypesv1.WeightedVoteOption{
				Option: govtypesv1.OptionNo,
				Weight: "0.1",
			},
			&govtypesv1.WeightedVoteOption{
				Option: govtypesv1.OptionYes,
				Weight: "0.1",
			},
			&govtypesv1.WeightedVoteOption{
				Option: govtypesv1.OptionAbstain,
				Weight: "0.1",
			},
		},
		proposalID,
	)
	requireT.NoError(err)

	// Wait for proposal result.
	finalStatus, err := gov.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusRejected, finalStatus)

	// Assert that proposer & depositor deposits were not credited back.
	proposerBalanceAfterVoting := accBalanceFunc(proposer)
	depositorBalanceAfterVoting := accBalanceFunc(depositor)
	requireT.Equal(proposerBalanceBeforeVoting, proposerBalanceAfterVoting)
	requireT.Equal(depositorBalanceBeforeVoting, depositorBalanceAfterVoting)
}

// TestExpeditedGovProposalWithDepositAndWeightedVotes tests expedited proposals.
func TestExpeditedGovProposalWithDepositAndWeightedVotes(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	gov := chain.Governance

	govParams, err := gov.QueryGovParams(ctx)
	requireT.NoError(err)

	// It is hardcoded from crust infra/apps/profiles.go and infra/apps/cored/config.go
	// remember to change these values if they are changed there
	unexpectedParams := govParams.ExpeditedVotingPeriod != lo.ToPtr(15*time.Second) ||
		len(govParams.ExpeditedMinDeposit) == 0 ||
		govParams.ExpeditedMinDeposit[0].Denom != chain.ChainSettings.Denom

	if unexpectedParams {
		govParams.ExpeditedMinDeposit = sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(2000)))
		govParams.ExpeditedVotingPeriod = lo.ToPtr(15 * time.Second)

		updateParamsMsg := &govtypesv1.MsgUpdateParams{
			Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			Params:    *govParams,
		}
		gov.ProposalFromMsgAndVote(
			ctx, t, nil,
			"-", "-", "-", govtypesv1.OptionYes,
			updateParamsMsg,
		)
	}

	missingDepositAmount := chain.NewCoin(sdkmath.NewInt(20))

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := gov.ComputeProposerBalance(ctx, true)
	requireT.NoError(err)
	proposerBalance = proposerBalance.Sub(missingDepositAmount).Add(chain.NewCoin(sdkmath.NewInt(1)))
	chain.Faucet.FundAccounts(ctx, t,
		integration.FundedAccount{
			Address: proposer,
			Amount:  proposerBalance,
		},
	)

	// Create proposer depositor.
	depositor := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, depositor, integration.BalancesOptions{
		Messages: []sdk.Msg{&govtypesv1.MsgDeposit{}},
		Amount:   missingDepositAmount.Amount,
	})

	proposalMsg, err := gov.NewMsgSubmitProposal(
		ctx,
		proposer,
		[]sdk.Msg{&banktypes.MsgSend{
			FromAddress: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			ToAddress:   depositor.String(),
			Amount:      []sdk.Coin{chain.NewCoin(sdkmath.NewInt(1))},
		}},
		"",
		"Send some funds to depositor",
		"Send some funds to depositor",
		true,
	)
	requireT.NoError(err)
	proposalMsg.InitialDeposit = sdk.NewCoins(proposalMsg.InitialDeposit...).Sub(sdk.Coins{missingDepositAmount}...)

	proposalID, err := gov.Propose(ctx, t, proposalMsg)
	requireT.NoError(err)

	// Verify that proposal is waiting for deposit.
	requirePropStatusFunc := func(expectedStatus govtypesv1.ProposalStatus) {
		proposal, err := gov.GetProposal(ctx, proposalID)
		requireT.NoError(err)
		requireT.Equal(expectedStatus, proposal.Status)
	}
	requirePropStatusFunc(govtypesv1.StatusDepositPeriod)

	// Deposit missing amount to proposal.
	depositMsg := govtypesv1.NewMsgDeposit(depositor, proposalID, sdk.Coins{missingDepositAmount})
	result, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(depositor),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(depositMsg)),
		depositMsg,
	)
	requireT.NoError(err)
	require.Equal(t, chain.GasLimitByMsgs(depositMsg), uint64(result.GasUsed))

	// Verify that proposal voting has started.
	requirePropStatusFunc(govtypesv1.StatusVotingPeriod)

	// Vote by all staker accounts:
	// NoWithVeto 70% & No,Yes,Abstain 10% each.
	err = gov.VoteAllWeighted(ctx,
		govtypesv1.WeightedVoteOptions{
			&govtypesv1.WeightedVoteOption{
				Option: govtypesv1.OptionNoWithVeto,
				Weight: "0.7",
			},
			&govtypesv1.WeightedVoteOption{
				Option: govtypesv1.OptionNo,
				Weight: "0.1",
			},
			&govtypesv1.WeightedVoteOption{
				Option: govtypesv1.OptionYes,
				Weight: "0.1",
			},
			&govtypesv1.WeightedVoteOption{
				Option: govtypesv1.OptionAbstain,
				Weight: "0.1",
			},
		},
		proposalID,
	)
	requireT.NoError(err)

	// Wait for proposal result.
	finalStatus, err := gov.WaitForVotingToFinalize(ctx, proposalID)
	requireT.NoError(err)
	requireT.Equal(govtypesv1.StatusRejected, finalStatus)
}

// TestGovCancelProposal tests cancelling proposals.
func TestGovCancelProposal(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	gov := chain.Governance
	missingDepositAmount := chain.NewCoin(sdkmath.NewInt(10))

	// Create new proposer.
	proposer := chain.GenAccount()
	proposerBalance, err := gov.ComputeProposerBalance(ctx, false)
	requireT.NoError(err)
	proposerBalance = proposerBalance.Sub(missingDepositAmount)
	chain.FundAccountWithOptions(ctx, t, proposer, integration.BalancesOptions{
		Amount: proposerBalance.Amount.Add(sdkmath.NewInt(200_000)).Add(sdkmath.NewInt(1)),
	})

	// Create proposer depositor.
	depositor := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, depositor, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&govtypesv1.MsgDeposit{},
		},
		Amount: missingDepositAmount.Amount,
	})

	proposalMsg, err := gov.NewMsgSubmitProposal(
		ctx,
		proposer,
		[]sdk.Msg{&banktypes.MsgSend{
			FromAddress: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
			ToAddress:   depositor.String(),
			Amount:      []sdk.Coin{chain.NewCoin(sdkmath.NewInt(1))},
		}},
		"",
		"Send some funds to depositor",
		"Send some funds to depositor",
		false,
	)
	requireT.NoError(err)

	proposalMsg.InitialDeposit = sdk.NewCoins(proposalMsg.InitialDeposit...).Sub(sdk.Coins{missingDepositAmount}...)
	proposalID, err := gov.Propose(ctx, t, proposalMsg)
	requireT.NoError(err)

	t.Logf("Proposal created, proposalID: %d", proposalID)

	// Store proposer and depositor balances before cancelling proposal.
	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	accBalanceFunc := func(prop sdk.AccAddress) sdk.Coin {
		accBalance, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
			Address: prop.String(),
			Denom:   chain.ChainSettings.Denom,
		})
		requireT.NoError(err)
		return *accBalance.Balance
	}

	// Verify that proposal is waiting for deposit.
	requirePropStatusFunc := func(expectedStatus govtypesv1.ProposalStatus) {
		proposal, err := gov.GetProposal(ctx, proposalID)
		requireT.NoError(err)
		requireT.Equal(expectedStatus, proposal.Status)
	}
	requirePropStatusFunc(govtypesv1.StatusDepositPeriod)

	// Deposit missing amount to proposal.
	depositMsg := govtypesv1.NewMsgDeposit(depositor, proposalID, sdk.Coins{missingDepositAmount})
	result, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(depositor),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(depositMsg)),
		depositMsg,
	)
	requireT.NoError(err)
	require.Equal(t, chain.GasLimitByMsgs(depositMsg), uint64(result.GasUsed))

	t.Logf("Deposited more funds to proposal, txHash:%s, gasUsed:%d", result.TxHash, result.GasUsed)

	// Verify that proposal voting has started.
	requirePropStatusFunc(govtypesv1.StatusVotingPeriod)

	_, err = gov.GetProposal(ctx, proposalID)
	requireT.NoError(err)

	depositorBalanceBeforeCancelling := accBalanceFunc(depositor)

	msgCancelProposal := &govtypesv1.MsgCancelProposal{
		ProposalId: proposalID,
		Proposer:   proposer.String(),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(proposer),
		chain.TxFactory().WithGas(200_000),
		msgCancelProposal,
	)
	requireT.NoError(err)

	// Proposal should not exist anymore.
	_, err = gov.GetProposal(ctx, proposalID)
	requireT.ErrorContains(err, "doesn't exist")

	params, err := gov.QueryGovParams(ctx)
	requireT.NoError(err)

	depositorRefundAfterCancelFee := sdkmath.LegacyOneDec().
		Sub(sdkmath.LegacyMustNewDecFromStr(params.ProposalCancelRatio)).
		Mul(sdkmath.LegacyNewDecFromInt(missingDepositAmount.Amount)).
		TruncateInt()

	// Assert that depositor deposits were credited back after applying cancel ratio.
	depositorBalanceAfterCancelling := accBalanceFunc(depositor)
	requireT.Equal(
		depositorBalanceAfterCancelling.Amount.Sub(depositorBalanceBeforeCancelling.Amount).String(),
		depositorRefundAfterCancelFee.String(),
	)
}
