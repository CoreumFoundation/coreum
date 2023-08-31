//go:build integrationtests

package modules

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v2/integration-tests"
	"github.com/CoreumFoundation/coreum/v2/pkg/client"
)

// TestAuthz tests the authz module Grant/Execute/Revoke messages execution and their deterministic gas.
func TestAuthz(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	authzClient := authztypes.NewQueryClient(chain.ClientContext)

	granter := chain.GenAccount()
	grantee := chain.GenAccount()
	recipient := chain.GenAccount()

	totalAmountToSend := sdkmath.NewInt(2_000)
	chain.FundAccountWithOptions(ctx, t, granter, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&authztypes.MsgGrant{},
			&authztypes.MsgRevoke{},
		},
		Amount: totalAmountToSend,
	})

	// init the messages provisionally to use in the authztypes.MsgExec
	msgBankSend := &banktypes.MsgSend{
		FromAddress: granter.String(),
		ToAddress:   recipient.String(),
		// send a half to have 2 messages in the Exec
		Amount: sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1_000))),
	}
	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{msgBankSend, msgBankSend})
	chain.FundAccountWithOptions(ctx, t, grantee, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			msgBankSend,
			&execMsg,
			&execMsg,
		},
	})

	// grant the bank send authorization
	grantMsg, err := authztypes.NewMsgGrant(
		granter,
		grantee,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
		lo.ToPtr(time.Now().Add(time.Minute)),
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
	requireT.ErrorIs(err, cosmoserrors.ErrInvalidPubKey)

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
	requireT.ErrorIs(err, authztypes.ErrNoAuthorizationFound)
}

// TestAuthZWithMultisig tests that the cosmos-sdk multisig accounts works with authz as grantee.
func TestAuthZWithMultisigGrantee(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	multisigPublicKey, keyNamesSet, err := chain.GenMultisigAccount(3, 2)
	requireT.NoError(err)
	multisigAddress := sdk.AccAddress(multisigPublicKey.Address())
	signer1KeyName := keyNamesSet[0]
	signer2KeyName := keyNamesSet[1]

	granter := chain.GenAccount()
	recipient := chain.GenAccount()
	amountToSendFromMultisigAccount := int64(1000)
	coinsToSendToRecipient := sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(amountToSendFromMultisigAccount)))

	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	// grant bank send authorization to multisig account
	grantMsg, err := authztypes.NewMsgGrant(
		granter,
		multisigAddress,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	chain.FundAccountWithOptions(ctx, t, granter, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{grantMsg},
		Amount:   sdkmath.NewInt(amountToSendFromMultisigAccount),
	})

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMsg)),
		grantMsg,
	)
	requireT.NoError(err)

	// create bank send msg account using authz
	msgBankSend := &banktypes.MsgSend{
		FromAddress: granter.String(),
		ToAddress:   recipient.String(),
		Amount:      coinsToSendToRecipient,
	}
	execMsg := authztypes.NewMsgExec(multisigAddress, []sdk.Msg{msgBankSend})
	chain.FundAccountWithOptions(ctx, t, multisigAddress, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&execMsg,
		},
	})

	_, err = chain.SignAndBroadcastMultisigTx(
		ctx,
		chain.ClientContext.WithFromAddress(multisigAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(&execMsg)),
		&execMsg,
		signer1KeyName)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)
	t.Log("Partially signed tx executed with expected error")

	// sign and submit with the min threshold
	txRes, err := chain.SignAndBroadcastMultisigTx(
		ctx,
		chain.ClientContext.WithFromAddress(multisigAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(&execMsg)),
		&execMsg,
		signer1KeyName, signer2KeyName)
	requireT.NoError(err)
	t.Logf("Fully signed tx executed, txHash:%s", txRes.TxHash)

	recipientBalances, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: recipient.String(),
	})
	requireT.NoError(err)
	requireT.Equal(coinsToSendToRecipient, recipientBalances.Balances)
}

// TestAuthZWithMultisig tests that the cosmos-sdk multisig accounts works with authz as granter.
func TestAuthZWithMultisigGranter(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)

	multisigPublicKey, keyNamesSet, err := chain.GenMultisigAccount(3, 2)
	requireT.NoError(err)
	multisigAddress := sdk.AccAddress(multisigPublicKey.Address())
	signer1KeyName := keyNamesSet[0]
	signer2KeyName := keyNamesSet[1]

	grantee := chain.GenAccount()
	recipient := chain.GenAccount()
	amountToSendFromMultisigAccount := int64(1000)
	coinsToSendToRecipient := sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(amountToSendFromMultisigAccount)))

	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	// grant bank send authorization to multisig account
	grantMsg, err := authztypes.NewMsgGrant(
		multisigAddress,
		grantee,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
		lo.ToPtr(time.Now().Add(time.Minute)),
	)
	require.NoError(t, err)

	chain.FundAccountWithOptions(ctx, t, multisigAddress, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			grantMsg,
		},
		Amount: sdkmath.NewInt(amountToSendFromMultisigAccount),
	})

	txRes, err := chain.SignAndBroadcastMultisigTx(
		ctx,
		chain.ClientContext.WithFromAddress(multisigAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(grantMsg)),
		grantMsg,
		signer1KeyName, signer2KeyName)
	requireT.NoError(err)
	t.Logf("Fully signed tx executed, txHash:%s", txRes.TxHash)

	// create bank send msg account using authz
	msgBankSend := &banktypes.MsgSend{
		FromAddress: multisigAddress.String(),
		ToAddress:   recipient.String(),
		Amount:      coinsToSendToRecipient,
	}

	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{msgBankSend})
	chain.FundAccountWithOptions(ctx, t, grantee, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&execMsg,
		},
	})

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(grantee),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(&execMsg)),
		&execMsg,
	)
	requireT.NoError(err)

	recipientBalances, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: recipient.String(),
	})
	requireT.NoError(err)
	requireT.Equal(coinsToSendToRecipient, recipientBalances.Balances)
}
