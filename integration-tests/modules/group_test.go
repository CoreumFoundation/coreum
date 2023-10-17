//go:build integrationtests

package modules

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
	"github.com/CoreumFoundation/coreum/v3/testutil/integration"
)

func TestGroupCreation(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	groupClient := group.NewQueryClient(chain.ClientContext)

	// Setup group admin account
	admin := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, admin, integration.BalancesOptions{
		Messages: []sdk.Msg{&group.MsgCreateGroup{}},
	})

	// Setup group member accounts
	groupMembers := lo.Times(3, func(i int) group.MemberRequest {
		address := chain.GenAccount()
		chain.FundAccountWithOptions(ctx, t, address, integration.BalancesOptions{
			Messages: []sdk.Msg{&group.MsgVote{}},
		})

		return group.MemberRequest{
			Address: address.String(),
			Weight:  "1",
		}
	})

	// Create group
	createGroupMsg := group.MsgCreateGroup{
		Admin:   admin.String(),
		Members: groupMembers,
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
		"",
		&group.ThresholdDecisionPolicy{
			Threshold: "2",
			Windows: &group.DecisionPolicyWindows{
				VotingPeriod:       time.Minute,
				MinExecutionPeriod: time.Second, // TODO: Rewise.
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
}
