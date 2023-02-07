//go:build integrationtests

package modules

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
)

// TestAuthz tests the authz module Grant/Execute/Revoke messages execution and their deterministic gas.
func TestAuthz(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	authzClient := authztypes.NewQueryClient(chain.ClientContext)

	granter := chain.GenAccount()
	grantee := chain.GenAccount()
	recipient := chain.GenAccount()

	totalAmountToSend := sdk.NewInt(2_000)
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, granter, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&authztypes.MsgGrant{},
			&authztypes.MsgRevoke{},
		},
		Amount: totalAmountToSend,
	}))

	// init the messages provisionally to use in the authztypes.MsgExec
	msgBankSend := &banktypes.MsgSend{
		FromAddress: granter.String(),
		ToAddress:   recipient.String(),
		// send a half to have 2 messages in the Exec
		Amount: sdk.NewCoins(chain.NewCoin(sdk.NewInt(1_000))),
	}
	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{msgBankSend, msgBankSend})
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, grantee, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			msgBankSend,
			&execMsg,
			&execMsg,
		},
	}))

	// grant the bank send authorization
	grantMsg, err := authztypes.NewMsgGrant(
		granter,
		grantee,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
		time.Now().Add(time.Minute),
	)
	require.NoError(t, err)

	txResult, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMsg)),
		grantMsg,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(grantMsg), uint64(txResult.GasUsed))

	// assert granted
	gransRes, err := authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: granter.String(),
		Grantee: grantee.String(),
	})
	requireT.NoError(err)
	requireT.Equal(1, len(gransRes.Grants))

	// try to send from grantee directly
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(grantee),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgBankSend)),
		msgBankSend,
	)
	requireT.ErrorIs(sdkerrors.ErrInvalidPubKey, err)

	// try to send using the authz
	txResult, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(grantee),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(&execMsg)),
		&execMsg,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(&execMsg), uint64(txResult.GasUsed))

	recipientBalancesRes, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: recipient.String(),
	})
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoins(chain.NewCoin(totalAmountToSend)).String(), recipientBalancesRes.Balances.String())

	// revoke the grant
	revokeMsg := authztypes.NewMsgRevoke(granter, grantee, sdk.MsgTypeURL(&banktypes.MsgSend{}))
	txResult, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(&revokeMsg)),
		&revokeMsg,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(&revokeMsg), uint64(txResult.GasUsed))

	gransRes, err = authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: granter.String(),
		Grantee: grantee.String(),
	})
	requireT.NoError(err)
	requireT.Equal(0, len(gransRes.Grants))

	// try to send with the revoked grant
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(grantee),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(&execMsg)),
		&execMsg,
	)
	requireT.ErrorIs(sdkerrors.ErrUnauthorized, err)
}
