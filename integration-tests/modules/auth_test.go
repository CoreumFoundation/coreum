//go:build integrationtests

package modules

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
	"github.com/CoreumFoundation/coreum/v3/testutil/integration"
	assetfttypes "github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
)

// TestAuthFeeLimits verifies that invalid message gas won't be accepted.
func TestAuthFeeLimits(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	sender := chain.GenAccount()

	feeModel := getFeemodelParams(ctx, t, chain.ClientContext)
	maxBlockGas := feeModel.MaxBlockGas
	chain.FundAccountWithOptions(ctx, t, sender, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&assetfttypes.MsgIssue{},
		},
		NondeterministicMessagesGas: uint64(maxBlockGas) + 100,
		Amount:                      chain.QueryAssetFTParams(ctx, t).IssueFee.Amount,
	})

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
	}

	gasPriceWithMaxDiscount := feeModel.InitialGasPrice.
		Mul(sdk.OneDec().Sub(feeModel.MaxDiscount))

	// the gas price is too low
	_, err := client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(chain.NewDecCoin(gasPriceWithMaxDiscount.QuoInt64(2)).String()),
		msg)
	require.True(t, cosmoserrors.ErrInsufficientFee.Is(err))

	// no gas price
	_, err = client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(""),
		msg)
	require.True(t, cosmoserrors.ErrInsufficientFee.Is(err))

	// more gas than MaxBlockGas
	_, err = client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(uint64(maxBlockGas+1)),
		msg)
	require.Error(t, err)

	// gas equal MaxBlockGas, the tx should pass
	_, err = client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(uint64(maxBlockGas)),
		msg)
	require.NoError(t, err)

	// fee paid in another coin is rejected
	const subunit = "uzzz" // uzzz is intentionally selected to put it on second position, after ucore, in sorted coins
	issueMsg := &assetfttypes.MsgIssue{
		Issuer:        sender.String(),
		Symbol:        "ZZZ",
		Subunit:       subunit,
		Precision:     6,
		Description:   "ZZZ Description",
		InitialAmount: sdkmath.NewInt(1000),
		Features:      []assetfttypes.Feature{},
	}
	denom := assetfttypes.BuildDenom(subunit, sender)
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	require.NoError(t, err)

	_, err = client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(sdk.NewInt64Coin(denom, 1).String()),
		msg)
	require.Error(t, err)
	require.True(t, cosmoserrors.ErrInvalidCoins.Is(err))

	// fee paid both in core and another coin is rejected
	_, err = client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(chain.TxFactory().GasPrices().Add(sdk.NewInt64DecCoin(denom, 1)).Sort().String()),
		msg)
	require.Error(t, err)
	require.True(t, cosmoserrors.ErrInvalidCoins.Is(err))
}

// TestAuthMultisig tests the cosmos-sdk multisig accounts and API.
func TestAuthMultisig(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	recipient := chain.GenAccount()
	amountToSendFromMultisigAccount := int64(1000)

	multisigPublicKey, keyNamesSet, err := chain.GenMultisigAccount(3, 2)
	requireT.NoError(err)
	multisigAddress := sdk.AccAddress(multisigPublicKey.Address())
	signer1KeyName := keyNamesSet[0]
	signer2KeyName := keyNamesSet[1]

	bankClient := banktypes.NewQueryClient(chain.ClientContext)

	// fund the multisig account
	chain.FundAccountWithOptions(ctx, t, multisigAddress, integration.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   sdkmath.NewInt(amountToSendFromMultisigAccount),
	})

	// prepare account to be funded from the multisig
	recipientAddr := recipient.String()
	coinsToSendToRecipient := sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(amountToSendFromMultisigAccount)))

	bankSendMsg := &banktypes.MsgSend{
		FromAddress: multisigAddress.String(),
		ToAddress:   recipientAddr,
		Amount:      coinsToSendToRecipient,
	}
	_, err = chain.SignAndBroadcastMultisigTx(
		ctx,
		chain.ClientContext.WithFromAddress(multisigAddress),
		// We intentionally use simulation instead of using `WithGas(chain.GasLimitByMsgs(bankSendMsg))`.
		// We do it to test simulation for multisig account.
		chain.TxFactory().WithSimulateAndExecute(true),
		bankSendMsg,
		signer1KeyName)
	requireT.ErrorIs(err, cosmoserrors.ErrUnauthorized)
	t.Log("Partially signed tx executed with expected error")

	// sign and submit with the min threshold
	txRes, err := chain.SignAndBroadcastMultisigTx(
		ctx,
		chain.ClientContext.WithFromAddress(multisigAddress),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(bankSendMsg)),
		bankSendMsg,
		signer1KeyName, signer2KeyName)
	requireT.NoError(err)
	t.Logf("Fully signed tx executed, txHash:%s", txRes.TxHash)

	recipientBalances, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: recipientAddr,
	})
	requireT.NoError(err)
	requireT.Equal(coinsToSendToRecipient, recipientBalances.Balances)
}

// TestAuthUnexpectedSequenceNumber test verifies that we correctly handle error reporting invalid account sequence number
// used to sign transaction.
func TestAuthUnexpectedSequenceNumber(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	sender := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, sender, integration.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   sdkmath.NewInt(10),
	})

	clientCtx := chain.ClientContext
	accInfo, err := client.GetAccountInfo(ctx, clientCtx, sender)
	require.NoError(t, err)

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdkmath.NewInt(1))),
	}

	_, err = client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithSequence(accInfo.GetSequence()+1). // incorrect sequence
			WithAccountNumber(accInfo.GetAccountNumber()).
			WithGas(chain.GasLimitByMsgs(msg)),
		msg)
	require.True(t, cosmoserrors.ErrWrongSequence.Is(err))
}

// TestAuthSignModeDirectAux tests SignModeDirectAux signing mode.
func TestAuthSignModeDirectAux(t *testing.T) {
	t.Parallel()
	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	// Tipper does not pay any tips yet because TipDecorator is not integrated yet (it is still in beta).
	tipper := chain.GenAccount()
	feePayer := chain.GenAccount()
	recipient := chain.GenAccount()

	amountToSend := sdk.NewIntFromUint64(1000)
	chain.FundAccountWithOptions(ctx, t, feePayer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
		},
	})
	chain.Faucet.FundAccounts(ctx, t, integration.FundedAccount{
		Address: tipper,
		Amount:  chain.NewCoin(amountToSend),
	})

	msg := &banktypes.MsgSend{
		FromAddress: tipper.String(),
		ToAddress:   recipient.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(amountToSend)),
	}

	tipperKey, err := chain.ClientContext.Keyring().KeyByAddress(tipper)
	requireT.NoError(err)
	tipperPubKey, err := tipperKey.GetPubKey()
	requireT.NoError(err)

	feePayerKey, err := chain.ClientContext.Keyring().KeyByAddress(feePayer)
	requireT.NoError(err)

	tipperAccountInfo, err := client.GetAccountInfo(ctx, chain.ClientContext, tipper)
	requireT.NoError(err)

	feePayerAccountInfo, err := client.GetAccountInfo(ctx, chain.ClientContext, feePayer)
	requireT.NoError(err)

	builder := clienttx.NewAuxTxBuilder()
	builder.SetChainID(chain.ClientContext.ChainID())
	requireT.NoError(builder.SetMsgs(msg))
	builder.SetAddress(tipper.String())
	requireT.NoError(builder.SetPubKey(tipperPubKey))
	builder.SetAccountNumber(tipperAccountInfo.GetAccountNumber())
	builder.SetSequence(tipperAccountInfo.GetSequence())
	requireT.NoError(builder.SetSignMode(signing.SignMode_SIGN_MODE_DIRECT_AUX))

	signBytes, err := builder.GetSignBytes()
	requireT.NoError(err)

	tipperSignature, _, err := chain.ClientContext.Keyring().SignByAddress(tipper, signBytes)
	requireT.NoError(err)

	builder.SetSignature(tipperSignature)
	tipperSignerData, err := builder.GetAuxSignerData()
	requireT.NoError(err)

	gas := chain.GasLimitByMsgs(msg)
	txBuilder := chain.ClientContext.TxConfig().NewTxBuilder()
	requireT.NoError(txBuilder.AddAuxSignerData(tipperSignerData))
	txBuilder.SetFeePayer(feePayer)
	txBuilder.SetFeeAmount(sdk.NewCoins(chain.NewCoin(chain.ChainSettings.GasPrice.Mul(sdk.NewDecFromInt(sdk.NewIntFromUint64(gas))).Ceil().RoundInt())))
	txBuilder.SetGasLimit(gas)

	requireT.NoError(clienttx.Sign(chain.TxFactory().
		WithAccountNumber(feePayerAccountInfo.GetAccountNumber()).
		WithSequence(feePayerAccountInfo.GetSequence()),
		feePayerKey.Name,
		txBuilder,
		false))
	txBytes, err := chain.ClientContext.TxConfig().TxEncoder()(txBuilder.GetTx())
	requireT.NoError(err)

	_, err = client.BroadcastRawTx(ctx, chain.ClientContext, txBytes)
	requireT.NoError(err)

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	tipperBalanceResp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: tipper.String(),
		Denom:   chain.ChainSettings.Denom,
	})
	requireT.NoError(err)

	feePayerBalanceResp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: feePayer.String(),
		Denom:   chain.ChainSettings.Denom,
	})
	requireT.NoError(err)

	recipientBalanceResp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: recipient.String(),
		Denom:   chain.ChainSettings.Denom,
	})
	requireT.NoError(err)

	requireT.Equal(chain.NewCoin(sdk.ZeroInt()).String(), tipperBalanceResp.Balance.String())
	requireT.Equal(chain.NewCoin(sdk.ZeroInt()).String(), feePayerBalanceResp.Balance.String())
	requireT.Equal(chain.NewCoin(amountToSend).String(), recipientBalanceResp.Balance.String())
}
