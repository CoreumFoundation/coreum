package ante_test

import (
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

func TestAnteHandler(t *testing.T) {
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.BeginNextBlock(time.Now())
	account, accountPrivKey := testApp.GenAccount(ctx)
	testApp.EndBlockAndCommit(ctx)
	ctx = testApp.BeginNextBlock(time.Now())

	// Issue a token
	msg := &types.MsgIssue{
		Issuer:  account.String(),
		Symbol:  "fttoken",
		Subunit: "fttoken",
	}

	_, _, err := testApp.SimulateFundAndSendTx(
		ctx,
		accountPrivKey,
		msg,
	)
	requireT.NoError(err)
	denom := types.BuildDenom(msg.Subunit, account)

	notAllowedDeposit := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1000)))
	allowedDeposit := sdk.NewCoins(
		sdk.NewCoin(testApp.GovKeeper.GetParams(ctx).MinDeposit[0].Denom, sdk.NewInt(1000)),
	)

	textProposal := govtypesv1beta1.NewTextProposal("Test proposal",
		strings.Repeat("Description", 20))

	testCases := []struct {
		name     string
		messages func() []sdk.Msg
	}{
		{
			name: "proposal_v1",
			messages: func() []sdk.Msg {
				msgExecLegacy, err := govv1.NewLegacyContent(textProposal,
					authtypes.NewModuleAddress(govtypes.ModuleName).String())
				requireT.NoError(err)
				msgProposalV1, err := govv1.NewMsgSubmitProposal(
					[]sdk.Msg{msgExecLegacy},
					notAllowedDeposit,
					account.String(),
					textProposal.GetDescription(),
					textProposal.GetTitle(),
					textProposal.GetTitle(),
				)
				requireT.NoError(err)
				return []sdk.Msg{msgProposalV1}
			},
		},
		{
			name: "deposit_v1",
			messages: func() []sdk.Msg {
				msgExecLegacy, err := govv1.NewLegacyContent(textProposal,
					authtypes.NewModuleAddress(govtypes.ModuleName).String())
				requireT.NoError(err)
				msgProposalV1, err := govv1.NewMsgSubmitProposal(
					[]sdk.Msg{msgExecLegacy},
					allowedDeposit,
					account.String(),
					textProposal.GetDescription(),
					textProposal.GetTitle(),
					textProposal.GetTitle(),
				)
				requireT.NoError(err)
				proposalID, err := testApp.GovKeeper.GetProposalID(ctx)
				requireT.NoError(err)
				return []sdk.Msg{
					msgProposalV1,
					&govv1.MsgDeposit{
						ProposalId: proposalID,
						Depositor:  account.String(),
						Amount:     notAllowedDeposit,
					},
				}
			},
		},
		{
			name: "proposal_v1bata1",
			messages: func() []sdk.Msg {
				msgProposalV1Beta1, err := govtypesv1beta1.NewMsgSubmitProposal(
					textProposal,
					notAllowedDeposit,
					account)
				requireT.NoError(err)
				return []sdk.Msg{msgProposalV1Beta1}
			},
		},
		{
			name: "deposit_v1beta1",
			messages: func() []sdk.Msg {
				msgExecLegacy, err := govv1.NewLegacyContent(textProposal,
					authtypes.NewModuleAddress(govtypes.ModuleName).String())
				requireT.NoError(err)
				msgProposalV1, err := govv1.NewMsgSubmitProposal(
					[]sdk.Msg{msgExecLegacy},
					allowedDeposit,
					account.String(),
					textProposal.GetDescription(),
					textProposal.GetTitle(),
					textProposal.GetTitle(),
				)
				requireT.NoError(err)
				proposalID, err := testApp.GovKeeper.GetProposalID(ctx)
				requireT.NoError(err)
				return []sdk.Msg{
					msgProposalV1,
					&govtypesv1beta1.MsgDeposit{
						ProposalId: proposalID,
						Depositor:  account.String(),
						Amount:     notAllowedDeposit,
					},
				}
			},
		},
	}

	// fund account with deposit amounts for every test
	requireT.NoError(testApp.FundAccount(
		ctx,
		account,
		notAllowedDeposit.Add(allowedDeposit...).MulInt(sdk.NewInt(int64(len(testCases)))),
	))

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testApp.EndBlockAndCommit(ctx)
			ctx = testApp.BeginNextBlock(time.Now())
			requireT = require.New(t)
			_, _, err = testApp.SimulateFundAndSendTx(
				ctx,
				accountPrivKey,
				tc.messages()...,
			)
			requireT.Error(err)
			requireT.ErrorIs(err, cosmoserrors.ErrInvalidCoins)
		})
	}
}
