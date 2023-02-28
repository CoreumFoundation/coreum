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
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
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

// TestAuthz tests the authz module works well with assetft module.
func TestAuthzWithAssetFT(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)

	assetftClient := assetfttypes.NewQueryClient(chain.ClientContext)
	authzClient := authztypes.NewQueryClient(chain.ClientContext)

	granter := chain.GenAccount()
	grantee := chain.GenAccount()
	recipient := chain.GenAccount()

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, granter, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetfttypes.MsgIssue{},
			&authztypes.MsgGrant{},
			&authztypes.MsgGrant{},
		},
		Amount: chain.NetworkConfig.AssetFTConfig.IssueFee,
	}))

	// mint and grant authorization
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        granter.String(),
		Symbol:        "symbol",
		Subunit:       "subunit",
		Precision:     1,
		InitialAmount: sdk.NewInt(1000),
		Features: []assetfttypes.Feature{
			assetfttypes.Feature_freezing,
			assetfttypes.Feature_whitelisting,
		},
	}
	denom := assetfttypes.BuildDenom(issueMsg.Subunit, granter)
	grantFreezeMsg, err := authztypes.NewMsgGrant(
		granter,
		grantee,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&assetfttypes.MsgFreeze{})),
		time.Now().Add(time.Minute),
	)
	require.NoError(t, err)

	grantWhitelistMsg, err := authztypes.NewMsgGrant(
		granter,
		grantee,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&assetfttypes.MsgSetWhitelistedLimit{})),
		time.Now().Add(time.Minute),
	)
	require.NoError(t, err)

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantFreezeMsg, grantWhitelistMsg, issueMsg)),
		grantFreezeMsg, grantWhitelistMsg, issueMsg,
	)
	requireT.NoError(err)

	// assert granted
	gransRes, err := authzClient.Grants(ctx, &authztypes.QueryGrantsRequest{
		Granter: granter.String(),
		Grantee: grantee.String(),
	})
	requireT.NoError(err)
	requireT.Equal(2, len(gransRes.Grants))

	// try to whitelist and freeze using the authz
	msgFreeze := &assetfttypes.MsgFreeze{
		Sender:  granter.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(1_000)),
	}

	msgWhitelist := &assetfttypes.MsgSetWhitelistedLimit{
		Sender:  granter.String(),
		Account: recipient.String(),
		Coin:    sdk.NewCoin(denom, sdk.NewInt(1_000)),
	}

	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{msgFreeze, msgWhitelist})
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, grantee, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&execMsg,
		},
	}))

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(grantee),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(&execMsg)),
		&execMsg,
	)
	requireT.NoError(err)

	freezingRes, err := assetftClient.FrozenBalance(ctx, &assetfttypes.QueryFrozenBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues("1000", freezingRes.GetBalance().Amount.String())

	whitelistingRes, err := assetftClient.WhitelistedBalance(ctx, &assetfttypes.QueryWhitelistedBalanceRequest{
		Account: recipient.String(),
		Denom:   denom,
	})
	requireT.NoError(err)
	requireT.EqualValues("1000", whitelistingRes.GetBalance().Amount.String())
}
