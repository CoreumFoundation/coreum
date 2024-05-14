//go:build integrationtests

package modules

import (
	"context"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v4/integration-tests"
	"github.com/CoreumFoundation/coreum/v4/pkg/client"
	"github.com/CoreumFoundation/coreum/v4/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

// TestGroupCreationAndBankSend creates group & group policy and then sends funds from group policy account.
func TestGroupCreationAndBankSend(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	groupClient := group.NewQueryClient(chain.ClientContext)

	// Setup group admin account
	admin := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&group.MsgCreateGroup{},
			&group.MsgCreateGroupPolicy{},
		},
	})

	// Generate group member accounts
	groupMembers := lo.Times(3, func(i int) group.MemberRequest {
		return group.MemberRequest{
			Address: chain.GenAccount().String(),
			Weight:  "1",
		}
	})

	// Fund group member accounts.
	// Since MsgSubmitProposal & MsgVote are non-deterministic we just fund each account with 1 CORE.
	accountsToFund := lo.Map(groupMembers, func(member group.MemberRequest, _ int) integration.FundedAccount {
		return integration.FundedAccount{
			Address: sdk.MustAccAddressFromBech32(member.Address),
			Amount:  chain.NewCoin(sdk.NewInt(1_000_000)),
		}
	})
	chain.Faucet.FundAccounts(ctx, t, accountsToFund...)

	// Create group
	createGroupMsg := group.MsgCreateGroup{
		Admin:    admin.String(),
		Members:  groupMembers,
		Metadata: "Integration test group",
	}

	result, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(&createGroupMsg)),
		&createGroupMsg,
	)
	requireT.NoError(err)

	groupsByAdmin, err := groupClient.GroupsByAdmin(ctx, &group.QueryGroupsByAdminRequest{
		Admin: admin.String(),
	})
	requireT.NoError(err)
	requireT.Len(groupsByAdmin.Groups, 1)

	grp := groupsByAdmin.Groups[0]
	t.Logf("created group, groupId:%d txHash:%s", grp.Id, result.TxHash)

	// Create group policy
	createGroupPolicyMsg, err := group.NewMsgCreateGroupPolicy(
		admin,
		grp.Id,
		"Integration test group policy",
		&group.ThresholdDecisionPolicy{
			Threshold: "2",
			Windows: &group.DecisionPolicyWindows{
				VotingPeriod:       time.Minute,
				MinExecutionPeriod: 100 * time.Millisecond, // Allow execution in 100ms after creation.
			},
		},
	)

	requireT.NoError(err)

	result, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(createGroupPolicyMsg)),
		createGroupPolicyMsg,
	)
	requireT.NoError(err)

	groupPolicies, err := groupClient.GroupPoliciesByGroup(ctx, &group.QueryGroupPoliciesByGroupRequest{
		GroupId: grp.Id,
	})
	requireT.NoError(err)

	requireT.Len(groupPolicies.GroupPolicies, 1)
	groupPolicy := groupPolicies.GroupPolicies[0]
	t.Logf("created group policy, groupPolicyAddress:%s txHash:%s", groupPolicy.Address, result.TxHash)

	groupSendCoin := chain.NewCoin(sdk.NewInt(100_000_000))
	chain.FundAccountWithOptions(ctx, t, sdk.MustAccAddressFromBech32(groupPolicy.Address), integration.BalancesOptions{
		Messages: []sdk.Msg{},
		Amount:   groupSendCoin.Amount,
	})

	// Create proposal
	groupCoinReceiver := chain.GenAccount()
	proposer := sdk.MustAccAddressFromBech32(groupMembers[0].Address)
	submitProposalMsg, err := group.NewMsgSubmitProposal(
		groupPolicy.Address,
		[]string{proposer.String()},
		[]sdk.Msg{&bank.MsgSend{
			FromAddress: groupPolicy.Address,
			ToAddress:   groupCoinReceiver.String(),
			Amount:      []sdk.Coin{groupSendCoin},
		}},
		"",
		group.Exec_EXEC_UNSPECIFIED,
		"Integration test for bank send proposal using group",
		"Integration test for bank send proposal using group",
	)
	requireT.NoError(err)

	proposal := submitGroupProposal(ctx, t, chain, proposer, submitProposalMsg)

	// Vote for proposal from other group members (except proposer).
	lo.ForEach(groupMembers[1:], func(member group.MemberRequest, _ int) {
		voteMsg := &group.MsgVote{
			ProposalId: proposal.Id,
			Voter:      member.Address,
			Option:     group.VOTE_OPTION_YES,
			Exec:       group.Exec_EXEC_TRY,
		}

		result, err = client.BroadcastTx(
			ctx,
			chain.ClientContext.WithFromAddress(sdk.MustAccAddressFromBech32(member.Address)),
			chain.TxFactory().WithSimulateAndExecute(true),
			voteMsg,
		)
		requireT.NoError(err)
	})

	_, err = groupClient.Proposal(ctx, &group.QueryProposalRequest{
		ProposalId: proposal.Id,
	})
	// The proposal will be automatically pruned after execution if successful.
	// https://docs.cosmos.network/v0.47/build/modules/group#executing-proposals
	requireT.Error(err)

	bankClient := bank.NewQueryClient(chain.ClientContext)
	receiverBalance, err := bankClient.Balance(ctx, &bank.QueryBalanceRequest{
		Address: groupCoinReceiver.String(),
		Denom:   chain.ChainSettings.Denom,
	})
	requireT.NoError(err)

	requireT.Equal(groupSendCoin.Amount, receiverBalance.Balance.Amount)
}

// TestGroupForAssetFTIssuance creates group & group policy and then issues FT using group policy account.
func TestGroupForAssetFTIssuance(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	groupClient := group.NewQueryClient(chain.ClientContext)

	admin := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&group.MsgCreateGroupWithPolicy{},
		},
	})

	groupMembers := lo.Times(3, func(i int) sdk.AccAddress { return chain.GenAccount() })
	proposer := groupMembers[0]
	voters := groupMembers[1:]

	// Fund group member accounts.
	// Since MsgSubmitProposal & MsgVote are non-deterministic we just fund each account with 1 CORE.
	accountsToFund := lo.Map(groupMembers, func(acc sdk.AccAddress, _ int) integration.FundedAccount {
		return integration.FundedAccount{
			Address: acc,
			Amount:  chain.NewCoin(sdk.NewInt(1_000_000)),
		}
	})
	chain.Faucet.FundAccounts(ctx, t, accountsToFund...)

	_, groupPolicy := createGroupWithPolicy(ctx, t, chain, admin, groupMembers)

	// Submit proposal #1
	submitProposalMsg, err := group.NewMsgSubmitProposal(
		groupPolicy.Address,
		[]string{groupMembers[0].String()},
		[]sdk.Msg{&assetfttypes.MsgIssue{
			Issuer:        groupPolicy.Address,
			Symbol:        "ABC",
			Subunit:       "uabc",
			Precision:     6,
			InitialAmount: sdkmath.NewInt(1000),
			Description:   "ABC",
			Features: []assetfttypes.Feature{
				assetfttypes.Feature_minting,
			},
		}},
		"Issue asset FT #1 using group",
		group.Exec_EXEC_UNSPECIFIED,
		"Issue asset FT using group",
		"Issue asset FT using group",
	)
	proposal1 := submitGroupProposal(ctx, t, chain, proposer, submitProposalMsg)

	// Vote for proposal #1
	lo.ForEach(voters, func(member sdk.AccAddress, _ int) {
		voteMsg := &group.MsgVote{
			ProposalId: proposal1.Id,
			Voter:      member.String(),
			Option:     group.VOTE_OPTION_NO,
			Exec:       group.Exec_EXEC_UNSPECIFIED,
		}

		_, err = client.BroadcastTx(
			ctx,
			chain.ClientContext.WithFromAddress(member),
			chain.TxFactory().WithSimulateAndExecute(true),
			voteMsg,
		)
		requireT.NoError(err)
	})

	// Withdraw proposal #1
	withdrawProposalMsg := &group.MsgWithdrawProposal{
		ProposalId: proposal1.Id,
		Address:    proposer.String(), // either proposer or group policy admin is able to withdraw.
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(proposer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(withdrawProposalMsg)),
		withdrawProposalMsg,
	)
	requireT.NoError(err)
	proposalInfo, err := groupClient.Proposal(ctx, &group.QueryProposalRequest{
		ProposalId: proposal1.Id,
	})
	requireT.NoError(err)
	requireT.Equal(group.PROPOSAL_STATUS_WITHDRAWN, proposalInfo.Proposal.Status)

	// Submit proposal #2
	submitProposalMsg.Metadata = "Issue asset FT #2 using group"
	proposal2 := submitGroupProposal(ctx, t, chain, proposer, submitProposalMsg)

	// Vote for proposal #2
	lo.ForEach(voters, func(member sdk.AccAddress, _ int) {
		voteMsg := &group.MsgVote{
			ProposalId: proposal2.Id,
			Voter:      member.String(),
			Option:     group.VOTE_OPTION_YES,
			Exec:       group.Exec_EXEC_UNSPECIFIED,
		}

		_, err = client.BroadcastTx(
			ctx,
			chain.ClientContext.WithFromAddress(member),
			chain.TxFactory().WithSimulateAndExecute(true),
			voteMsg,
		)
		requireT.NoError(err)

		// Make sure proposal is not executed.
		proposalInfo, err := groupClient.Proposal(ctx, &group.QueryProposalRequest{
			ProposalId: proposal2.Id,
		})
		requireT.NoError(err)
		requireT.Equal(group.PROPOSAL_STATUS_SUBMITTED, proposalInfo.Proposal.Status)
	})

	// Execute proposal #2 (first try)
	executeProposalMsg := &group.MsgExec{
		ProposalId: proposal2.Id,
		Executor:   proposer.String(),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(proposer),
		chain.TxFactory().WithSimulateAndExecute(true),
		executeProposalMsg,
	)
	requireT.NoError(err)

	// Proposal is accepted but not executed successfully because there is no enough balance to pay for FT issuance fee.
	proposal2Info, err := groupClient.Proposal(ctx, &group.QueryProposalRequest{
		ProposalId: proposal2.Id,
	})
	requireT.NoError(err)
	requireT.Equal(group.PROPOSAL_STATUS_ACCEPTED, proposal2Info.Proposal.Status)
	requireT.Equal(group.PROPOSAL_EXECUTOR_RESULT_FAILURE, proposal2Info.Proposal.ExecutorResult)

	// Fund group policy account with issuance fee
	chain.FundAccountWithOptions(ctx, t, sdk.MustAccAddressFromBech32(groupPolicy.Address), integration.BalancesOptions{
		Amount: chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	// Execute proposal #2 (second try)
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(proposer),
		chain.TxFactory().WithSimulateAndExecute(true),
		executeProposalMsg,
	)
	requireT.NoError(err)

	// Verify that asset is issued.
	bankClient := bank.NewQueryClient(chain.ClientContext)
	receiverBalance, err := bankClient.Balance(ctx, &bank.QueryBalanceRequest{
		Address: groupPolicy.Address,
		Denom:   assetfttypes.BuildDenom("uabc", sdk.MustAccAddressFromBech32(groupPolicy.Address)),
	})
	requireT.NoError(err)

	requireT.Equal(sdk.NewInt(1000), receiverBalance.Balance.Amount)
}

// TestGroupAdministration tests group administration functionality: update of metadata, admin, decision policy, etc.
func TestGroupAdministration(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	groupClient := group.NewQueryClient(chain.ClientContext)

	// Generate & fund group admin & member accounts
	admin := chain.GenAccount()
	groupMembers := lo.Times(5, func(i int) sdk.AccAddress {
		return chain.GenAccount()
	})
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&group.MsgCreateGroupWithPolicy{},
			&group.MsgUpdateGroupMembers{},
			&group.MsgUpdateGroupMetadata{},
			&group.MsgUpdateGroupPolicyDecisionPolicy{},
			&group.MsgUpdateGroupPolicyMetadata{},
			&group.MsgUpdateGroupAdmin{},
			&group.MsgUpdateGroupPolicyAdmin{},
		},
	})

	// Create group & group policy
	grp, groupPolicy := createGroupWithPolicy(ctx, t, chain, admin, groupMembers)

	// Update members & metadata in group
	groupMembersNew := groupMembers[1:] // remove first member
	updateGroupMembersMsg := &group.MsgUpdateGroupMembers{
		Admin:   admin.String(),
		GroupId: grp.Id,
		MemberUpdates: []group.MemberRequest{
			{
				Address: groupMembers[0].String(),
				Weight:  "0", // to remove member we should set its weight to 0
			},
		},
	}
	updateGroupMetadataMsg := &group.MsgUpdateGroupMetadata{
		Admin:    admin.String(),
		GroupId:  grp.Id,
		Metadata: "New group metadata",
	}

	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(updateGroupMembersMsg, updateGroupMetadataMsg)),
		updateGroupMembersMsg, updateGroupMetadataMsg,
	)
	requireT.NoError(err)

	groupMembersResp, err := groupClient.GroupMembers(ctx, &group.QueryGroupMembersRequest{
		GroupId: grp.Id,
	})
	requireT.NoError(err)
	requireT.Len(groupMembersResp.Members, len(groupMembersNew))

	groupInfoResp, err := groupClient.GroupInfo(ctx, &group.QueryGroupInfoRequest{
		GroupId: grp.Id,
	})
	requireT.NoError(err)
	requireT.Equal(updateGroupMetadataMsg.Metadata, groupInfoResp.Info.Metadata)

	// Update decision policy & metadata in group policy
	updateGroupPolicyDecisionPolicyMsg, err := group.NewMsgUpdateGroupPolicyDecisionPolicy(
		admin,
		sdk.MustAccAddressFromBech32(groupPolicy.Address),
		&group.ThresholdDecisionPolicy{
			Threshold: "3",
			Windows: &group.DecisionPolicyWindows{
				VotingPeriod:       time.Minute,
				MinExecutionPeriod: 100 * time.Millisecond,
			},
		},
	)
	requireT.NoError(err)
	updateGroupPolicyMetadataMsg := &group.MsgUpdateGroupPolicyMetadata{
		Admin:              admin.String(),
		GroupPolicyAddress: groupPolicy.Address,
		Metadata:           "New group policy metadata",
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(updateGroupPolicyDecisionPolicyMsg, updateGroupPolicyMetadataMsg)),
		updateGroupPolicyDecisionPolicyMsg, updateGroupPolicyMetadataMsg,
	)
	requireT.NoError(err)

	groupPolicyInfoRes, err := groupClient.GroupPolicyInfo(ctx, &group.QueryGroupPolicyInfoRequest{
		Address: groupPolicy.Address,
	})
	requireT.NoError(err)

	requireT.Equal(
		updateGroupPolicyDecisionPolicyMsg.DecisionPolicy.String(),
		groupPolicyInfoRes.Info.DecisionPolicy.String(),
	)
	requireT.Equal(updateGroupPolicyMetadataMsg.Metadata, groupPolicyInfoRes.Info.Metadata)

	// Update admin in both group & group policy
	adminNew := chain.GenAccount()
	updateGroupAdminMsg := &group.MsgUpdateGroupAdmin{
		Admin:    admin.String(),
		GroupId:  grp.Id,
		NewAdmin: adminNew.String(),
	}
	updateGroupPolicyAdminMsg := &group.MsgUpdateGroupPolicyAdmin{
		Admin:              admin.String(),
		GroupPolicyAddress: groupPolicy.Address,
		NewAdmin:           adminNew.String(),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(updateGroupAdminMsg, updateGroupPolicyAdminMsg)),
		updateGroupAdminMsg, updateGroupPolicyAdminMsg,
	)
	requireT.NoError(err)

	groupInfoResp, err = groupClient.GroupInfo(ctx, &group.QueryGroupInfoRequest{
		GroupId: grp.Id,
	})
	requireT.NoError(err)
	requireT.Equal(updateGroupAdminMsg.NewAdmin, groupInfoResp.Info.Admin)

	groupPolicyInfoRes, err = groupClient.GroupPolicyInfo(ctx, &group.QueryGroupPolicyInfoRequest{
		Address: groupPolicy.Address,
	})
	requireT.NoError(err)
	requireT.Equal(updateGroupPolicyAdminMsg.NewAdmin, groupPolicyInfoRes.Info.Admin)

	// Leave group
	memberToLeaveGroup := groupMembersNew[len(groupMembersNew)-1]
	chain.FundAccountWithOptions(
		ctx,
		t,
		sdk.MustAccAddressFromBech32(memberToLeaveGroup.String()),
		integration.BalancesOptions{
			Messages: []sdk.Msg{&group.MsgLeaveGroup{}},
		},
	)
	leaveGroupMsg := &group.MsgLeaveGroup{
		Address: memberToLeaveGroup.String(),
		GroupId: grp.Id,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sdk.MustAccAddressFromBech32(memberToLeaveGroup.String())),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(leaveGroupMsg)),
		leaveGroupMsg,
	)
	requireT.NoError(err)

	groupMembersResp, err = groupClient.GroupMembers(ctx, &group.QueryGroupMembersRequest{
		GroupId: grp.Id,
	})
	requireT.NoError(err)
	requireT.Len(groupMembersResp.Members, len(groupMembersNew)-1)
}

// createGroupWithPolicy simple helper function to creates group & group policy with hardcoded params & customizable
// member list and admin.
func createGroupWithPolicy(
	ctx context.Context,
	t *testing.T,
	chain integration.CoreumChain,
	admin sdk.AccAddress,
	groupMembers []sdk.AccAddress,
) (*group.GroupInfo, *group.GroupPolicyInfo) {
	requireT := require.New(t)
	groupClient := group.NewQueryClient(chain.ClientContext)

	membersRequest := lo.Map(groupMembers, func(member sdk.AccAddress, _ int) group.MemberRequest {
		return group.MemberRequest{
			Address: member.String(),
			Weight:  "1",
		}
	})

	// Create group & group policy
	createGroupWithPolicyMsg, err := group.NewMsgCreateGroupWithPolicy(
		admin.String(),
		membersRequest,
		"Integration test group",
		"Integration test group policy",
		false,
		&group.PercentageDecisionPolicy{
			Percentage: "0.45",
			Windows: &group.DecisionPolicyWindows{
				VotingPeriod:       time.Minute,
				MinExecutionPeriod: 100 * time.Millisecond,
			},
		})

	requireT.NoError(err)

	result, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(admin),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(createGroupWithPolicyMsg)),
		createGroupWithPolicyMsg,
	)
	requireT.NoError(err)

	groupsByAdmin, err := groupClient.GroupsByAdmin(ctx, &group.QueryGroupsByAdminRequest{
		Admin: admin.String(),
	})
	requireT.NoError(err)
	requireT.Len(groupsByAdmin.Groups, 1)
	grp := groupsByAdmin.Groups[0]

	groupPolicies, err := groupClient.GroupPoliciesByGroup(ctx, &group.QueryGroupPoliciesByGroupRequest{
		GroupId: grp.Id,
	})
	requireT.NoError(err)

	requireT.Len(groupPolicies.GroupPolicies, 1)
	groupPolicy := groupPolicies.GroupPolicies[0]
	t.Logf(
		"created group with policy, groupId: %d groupPolicyAddress:%s txHash:%s",
		grp.Id, groupPolicy.Address, result.TxHash)

	return grp, groupPolicy
}

// submitGroupProposal simple helper function to submit group proposal & verify that creation was successful.
func submitGroupProposal(
	ctx context.Context,
	t *testing.T,
	chain integration.CoreumChain,
	proposer sdk.AccAddress,
	submitProposalMsg *group.MsgSubmitProposal,
) *group.Proposal {
	requireT := require.New(t)
	groupClient := group.NewQueryClient(chain.ClientContext)

	proposalsBefore, err := groupClient.ProposalsByGroupPolicy(ctx, &group.QueryProposalsByGroupPolicyRequest{
		Address: submitProposalMsg.GroupPolicyAddress,
	})
	requireT.NoError(err)

	result, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(proposer),
		chain.TxFactory().WithSimulateAndExecute(true),
		submitProposalMsg,
	)
	requireT.NoError(err)

	proposalsAfter, err := groupClient.ProposalsByGroupPolicy(ctx, &group.QueryProposalsByGroupPolicyRequest{
		Address: submitProposalMsg.GroupPolicyAddress,
	})
	requireT.NoError(err)
	requireT.Len(proposalsAfter.Proposals, len(proposalsBefore.Proposals)+1)

	createdProposal := proposalsAfter.Proposals[len(proposalsAfter.Proposals)-1]
	requireT.Equal(group.PROPOSAL_STATUS_SUBMITTED, createdProposal.Status)
	t.Logf("submitted group proposal, id:%d txHash:%s", createdProposal.Id, result.TxHash)

	return createdProposal
}
