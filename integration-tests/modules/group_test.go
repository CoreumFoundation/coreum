//go:build integrationtests

package modules

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
	"github.com/CoreumFoundation/coreum/v3/testutil/integration"
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

	// Setup group member accounts
	groupMembers := lo.Times(3, func(i int) group.MemberRequest {
		return group.MemberRequest{
			Address: chain.GenAccount().String(),
			Weight:  "1",
		}
	})

	chain.FundAccountWithOptions(ctx,
		t,
		sdk.MustAccAddressFromBech32(groupMembers[0].Address),
		integration.BalancesOptions{
			// First group member submits proposal
			Messages: []sdk.Msg{&group.MsgSubmitProposal{}},
		},
	)

	for i := 1; i < len(groupMembers); i++ {
		chain.FundAccountWithOptions(ctx,
			t,
			sdk.MustAccAddressFromBech32(groupMembers[i].Address),
			integration.BalancesOptions{
				// Other group members vote
				Messages: []sdk.Msg{&group.MsgVote{}},
			},
		)
	}

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
	createProposalMsg, err := group.NewMsgSubmitProposal(
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

	result, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(proposer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(createProposalMsg)),
		createProposalMsg,
	)
	requireT.NoError(err)

	proposals, err := groupClient.ProposalsByGroupPolicy(ctx, &group.QueryProposalsByGroupPolicyRequest{
		Address: groupPolicy.Address,
	})
	requireT.NoError(err)
	requireT.Len(proposals.Proposals, 1)

	proposal := proposals.Proposals[0]
	requireT.Equal(group.PROPOSAL_STATUS_SUBMITTED, proposal.Status)
	t.Logf("submitted group proposal, id:%d txHash:%s", proposal.Id, result.TxHash)

	// Vote for proposal from other group members (except proposer).
	for i := 1; i < len(groupMembers); i++ {
		voter := sdk.MustAccAddressFromBech32(groupMembers[i].Address)
		voteMsg := &group.MsgVote{
			ProposalId: proposal.Id,
			Voter:      voter.String(),
			Option:     group.VOTE_OPTION_YES,
			Exec:       group.Exec_EXEC_TRY,
		}

		result, err = client.BroadcastTx(
			ctx,
			chain.ClientContext.WithFromAddress(voter),
			chain.TxFactory().WithGas(chain.GasLimitByMsgs(voteMsg)),
			voteMsg,
		)
		requireT.NoError(err)
	}

	_, err = groupClient.Proposal(ctx, &group.QueryProposalRequest{
		ProposalId: proposal.Id,
	})
	requireT.Error(err)

	bankClient := bank.NewQueryClient(chain.ClientContext)
	receiverBalance, err := bankClient.Balance(ctx, &bank.QueryBalanceRequest{
		Address: groupCoinReceiver.String(),
		Denom:   chain.ChainSettings.Denom,
	})
	requireT.NoError(err)

	requireT.Equal(groupSendCoin.Amount, receiverBalance.Balance.Amount)
}

func TestGroupAdministration(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	groupClient := group.NewQueryClient(chain.ClientContext)

	// Setup group admin account
	admin := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&group.MsgCreateGroupWithPolicy{},
			&group.MsgCreateGroupWithPolicy{},
			&group.MsgCreateGroupWithPolicy{}, // fixme
			&group.MsgCreateGroupWithPolicy{},
			&group.MsgCreateGroupWithPolicy{},
			&group.MsgCreateGroupWithPolicy{},
			&group.MsgCreateGroupWithPolicy{},
			&group.MsgCreateGroupWithPolicy{},
			&group.MsgCreateGroupWithPolicy{},
		},
	})

	// Setup group member accounts
	groupMembers := lo.Times(5, func(i int) group.MemberRequest {
		return group.MemberRequest{
			Address: chain.GenAccount().String(),
			Weight:  "1",
		}
	})

	// Create group & group policy
	createGroupWithPolicyMsg, err := group.NewMsgCreateGroupWithPolicy(
		admin.String(),
		groupMembers,
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
	t.Logf("created group with policy, groupId: %d groupPolicyAddress:%s txHash:%s", grp.Id, groupPolicy.Address, result.TxHash)

	// Update members & metadata in group
	groupMembersNew := groupMembers[1:] // remove first member
	updateGroupMembersMsg := &group.MsgUpdateGroupMembers{
		Admin:   admin.String(),
		GroupId: grp.Id,
		MemberUpdates: []group.MemberRequest{
			{
				Address: groupMembers[0].Address,
				Weight:  "0", // to remove member we should set it's weight to 0
			},
		},
	}
	updateGroupMetadataMsg := &group.MsgUpdateGroupMetadata{
		Admin:    admin.String(),
		GroupId:  grp.Id,
		Metadata: "New group metadata",
	}

	_, err = client.BroadcastTx(
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

	requireT.Equal(updateGroupPolicyDecisionPolicyMsg.DecisionPolicy.String(), groupPolicyInfoRes.Info.DecisionPolicy.String())
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
	chain.FundAccountWithOptions(ctx, t, sdk.MustAccAddressFromBech32(memberToLeaveGroup.Address), integration.BalancesOptions{
		Messages: []sdk.Msg{&group.MsgLeaveGroup{}},
	})
	leaveGroupMsg := &group.MsgLeaveGroup{
		Address: memberToLeaveGroup.Address,
		GroupId: grp.Id,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sdk.MustAccAddressFromBech32(memberToLeaveGroup.Address)),
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
