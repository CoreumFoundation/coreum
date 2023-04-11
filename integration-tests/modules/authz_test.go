//go:build integrationtests

package modules

import (
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdksigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
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
	requireT.ErrorIs(err, sdkerrors.ErrInvalidPubKey)

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
	requireT.ErrorIs(err, sdkerrors.ErrUnauthorized)
}

// TestAuthZWithMultisig tests that the cosmos-sdk multisig accounts works with authz as grantee.
func TestAuthZWithMultisigGrantee(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)

	multisigPublicKey, keyNamesSet := chain.GenMultisigAccount(t, 3, 2)
	multisigAddress := sdk.AccAddress(multisigPublicKey.Address())
	signer1KeyName := keyNamesSet[0]
	signer2KeyName := keyNamesSet[1]

	granter := chain.GenAccount()
	recipient := chain.GenAccount()
	amountToSendFromMultisigAccount := int64(1000)
	coinsToSendToRecipient := sdk.NewCoins(chain.NewCoin(sdk.NewInt(amountToSendFromMultisigAccount)))

	// grant bank send authorization to multisig account
	grantMsg, err := authztypes.NewMsgGrant(
		granter,
		multisigAddress,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
		time.Now().Add(time.Minute),
	)
	require.NoError(t, err)

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, granter, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{grantMsg},
		Amount:   sdk.NewInt(amountToSendFromMultisigAccount),
	}))

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
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, multisigAddress, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&execMsg,
		},
	}))

	// prepare the tx factory to sign with the account seq and number of the multisig account
	clientCtx := chain.ClientContext
	multisigAccInfo, err := client.GetAccountInfo(ctx, clientCtx, multisigAddress)
	requireT.NoError(err)
	txF := chain.TxFactory().
		WithGas(chain.GasLimitByMsgs(&execMsg)).
		WithAccountNumber(multisigAccInfo.GetAccountNumber()).
		WithSequence(multisigAccInfo.GetSequence()).
		WithSignMode(sdksigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)

	// sign and submit with just one key to check the tx rejection
	txBuilder, err := txF.BuildUnsignedTx(&execMsg)
	requireT.NoError(err)

	err = client.Sign(txF, signer1KeyName, txBuilder, false)
	requireT.NoError(err)
	multisigTx := createMulisignTx(requireT, txBuilder, multisigAccInfo.GetSequence(), multisigPublicKey)
	encodedTx, err := clientCtx.TxConfig().TxEncoder()(multisigTx)
	requireT.NoError(err)
	_, err = client.BroadcastRawTx(ctx, clientCtx, encodedTx)
	requireT.True(sdkerrors.ErrUnauthorized.Is(err))
	logger.Get(ctx).Info("Partially signed tx executed with expected error")

	// sign and submit with the min threshold
	txBuilder, err = txF.BuildUnsignedTx(&execMsg)
	requireT.NoError(err)
	err = client.Sign(txF, signer1KeyName, txBuilder, false)
	requireT.NoError(err)
	err = client.Sign(txF, signer2KeyName, txBuilder, false)
	requireT.NoError(err)
	multisigTx = createMulisignTx(requireT, txBuilder, multisigAccInfo.GetSequence(), multisigPublicKey)
	encodedTx, err = clientCtx.TxConfig().TxEncoder()(multisigTx)
	requireT.NoError(err)
	result, err := client.BroadcastRawTx(ctx, clientCtx, encodedTx)
	requireT.NoError(err)
	logger.Get(ctx).Info("Fully signed tx executed", zap.String("txHash", result.TxHash))

	bankClient := banktypes.NewQueryClient(clientCtx)
	recipientBalances, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: recipient.String(),
	})
	requireT.NoError(err)
	requireT.Equal(coinsToSendToRecipient, recipientBalances.Balances)
}

// TestAuthZWithMultisig tests that the cosmos-sdk multisig accounts works with authz as granter.
func TestAuthZWithMultisigGranter(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)

	multisigPublicKey, keyNamesSet := chain.GenMultisigAccount(t, 3, 2)
	multisigAddress := sdk.AccAddress(multisigPublicKey.Address())
	signer1KeyName := keyNamesSet[0]
	signer2KeyName := keyNamesSet[1]

	grantee := chain.GenAccount()
	recipient := chain.GenAccount()
	amountToSendFromMultisigAccount := int64(1000)
	coinsToSendToRecipient := sdk.NewCoins(chain.NewCoin(sdk.NewInt(amountToSendFromMultisigAccount)))

	// grant bank send authorization to multisig account
	grantMsg, err := authztypes.NewMsgGrant(
		multisigAddress,
		grantee,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&banktypes.MsgSend{})),
		time.Now().Add(time.Minute),
	)
	require.NoError(t, err)

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, multisigAddress, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			grantMsg,
		},
		Amount: sdk.NewInt(amountToSendFromMultisigAccount),
	}))

	// prepare the tx factory to sign with the account seq and number of the multisig account
	clientCtx := chain.ClientContext
	multisigAccInfo, err := client.GetAccountInfo(ctx, clientCtx, multisigAddress)
	requireT.NoError(err)
	txF := chain.TxFactory().
		WithGas(chain.GasLimitByMsgs(grantMsg)).
		WithAccountNumber(multisigAccInfo.GetAccountNumber()).
		WithSequence(multisigAccInfo.GetSequence()).
		WithSignMode(sdksigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)

	// sign and submit with the min threshold
	txBuilder, err := txF.BuildUnsignedTx(grantMsg)
	requireT.NoError(err)
	err = client.Sign(txF, signer1KeyName, txBuilder, false)
	requireT.NoError(err)
	err = client.Sign(txF, signer2KeyName, txBuilder, false)
	requireT.NoError(err)
	multisigTx := createMulisignTx(requireT, txBuilder, multisigAccInfo.GetSequence(), multisigPublicKey)
	encodedTx, err := clientCtx.TxConfig().TxEncoder()(multisigTx)
	requireT.NoError(err)
	result, err := client.BroadcastRawTx(ctx, clientCtx, encodedTx)
	requireT.NoError(err)
	logger.Get(ctx).Info("Fully signed tx executed", zap.String("txHash", result.TxHash))

	// create bank send msg account using authz
	msgBankSend := &banktypes.MsgSend{
		FromAddress: multisigAddress.String(),
		ToAddress:   recipient.String(),
		Amount:      coinsToSendToRecipient,
	}

	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{msgBankSend})
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

	bankClient := banktypes.NewQueryClient(clientCtx)
	recipientBalances, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: recipient.String(),
	})
	requireT.NoError(err)
	requireT.Equal(coinsToSendToRecipient, recipientBalances.Balances)
}
