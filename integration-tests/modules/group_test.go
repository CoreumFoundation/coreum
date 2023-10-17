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

func TestGroupCreation(t *testing.T) {
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
	memberFundBalance := chain.ComputeNeededBalanceFromOptions(integration.BalancesOptions{
		Messages: []sdk.Msg{&group.MsgVote{}, &group.MsgSubmitProposal{}, &bank.MsgSend{}},
	})
	accountsToFund := make([]integration.FundedAccount, 0, 3)
	groupMembers := lo.Times(3, func(i int) group.MemberRequest {
		address := chain.GenAccount()
		accountsToFund = append(accountsToFund, integration.FundedAccount{
			Address: address,
			Amount:  chain.NewCoin(memberFundBalance),
		})

		return group.MemberRequest{
			Address: address.String(),
			Weight:  "1",
		}
	})

	chain.Faucet.FundAccounts(ctx, t, accountsToFund...)

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
				MinExecutionPeriod: time.Millisecond, // TODO: Rewise.
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

	groupSendCoin := chain.NewCoin(sdk.NewInt(100_000))
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
		"Title",
		"Summary",
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

	// Vote for proposal
	for i := 0; i < 2; i++ {
		voter := sdk.MustAccAddressFromBech32(groupMembers[i+1].Address)
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

	//tally, err := groupClient.TallyResult(ctx, &group.QueryTallyResultRequest{
	//	ProposalId: proposal.Id,
	//})
	//requireT.NoError(err)
	//tally = tally
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
