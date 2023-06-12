//go:build integrationtests

package modules

import (
	"testing"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdkmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	multisigtypes "github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdksigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// TestAuthFeeLimits verifies that invalid message gas won't be accepted.
func TestAuthFeeLimits(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	sender := chain.GenAccount()

	feeModel := getFeemodelParams(ctx, t, chain.ClientContext)
	maxBlockGas := feeModel.MaxBlockGas
	chain.FundAccountsWithOptions(ctx, t, sender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&banktypes.MsgSend{},
			&assetfttypes.MsgIssue{},
		},
		NondeterministicMessagesGas: uint64(maxBlockGas) + 100,
		Amount:                      getIssueFee(ctx, t, chain.ClientContext).Amount,
	})

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(1))),
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
	require.True(t, sdkerrors.ErrInsufficientFee.Is(err))

	// no gas price
	_, err = client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(""),
		msg)
	require.True(t, sdkerrors.ErrInsufficientFee.Is(err))

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
		InitialAmount: sdk.NewInt(1000),
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
	require.True(t, sdkerrors.ErrInvalidCoins.Is(err))

	// fee paid both in core and another coin is rejected
	_, err = client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(chain.TxFactory().GasPrices().Add(sdk.NewInt64DecCoin(denom, 1)).Sort().String()),
		msg)
	require.Error(t, err)
	require.True(t, sdkerrors.ErrInvalidCoins.Is(err))
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

	// fund the multisig account
	chain.FundAccountsWithOptions(ctx, t, multisigAddress, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   sdk.NewInt(amountToSendFromMultisigAccount),
	})

	// prepare account to be funded from the multisig
	recipientAddr := recipient.String()
	coinsToSendToRecipient := sdk.NewCoins(chain.NewCoin(sdk.NewInt(amountToSendFromMultisigAccount)))

	clientCtx := chain.ClientContext
	// prepare the tx factory to sign with the account number of the multisig account
	multisigAccInfo, err := client.GetAccountInfo(ctx, clientCtx, multisigAddress)
	requireT.NoError(err)
	txF := chain.TxFactory().
		WithGas(chain.GasLimitByMsgs(&banktypes.MsgSend{})).
		WithAccountNumber(multisigAccInfo.GetAccountNumber()).
		WithSignMode(sdksigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)

	bankSendMsg := &banktypes.MsgSend{
		FromAddress: multisigAddress.String(),
		ToAddress:   recipientAddr,
		Amount:      coinsToSendToRecipient,
	}
	// sign and submit with just one key to check the tx rejection
	txBuilder, err := txF.BuildUnsignedTx(bankSendMsg)
	requireT.NoError(err)

	err = client.Sign(txF, signer1KeyName, txBuilder, false)
	requireT.NoError(err)
	multisigTx := createMulisignTx(requireT, txBuilder, multisigAccInfo.GetSequence(), multisigPublicKey)
	encodedTx, err := clientCtx.TxConfig().TxEncoder()(multisigTx)
	requireT.NoError(err)
	_, err = client.BroadcastRawTx(ctx, clientCtx, encodedTx)
	requireT.True(sdkerrors.ErrUnauthorized.Is(err))
	t.Log("Partially signed tx executed with expected error")

	// sign and submit with the min threshold
	txBuilder, err = txF.BuildUnsignedTx(bankSendMsg)
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
	t.Logf("Fully signed tx executed, txHash:%s", result.TxHash)

	bankClient := banktypes.NewQueryClient(clientCtx)
	recipientBalances, err := bankClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: recipientAddr,
	})
	requireT.NoError(err)
	requireT.Equal(coinsToSendToRecipient, recipientBalances.Balances)
}

// TestAuthMultisigSequences tests the cosmos-sdk sequences behaviour for multisig account.
func TestAuthMultisigSequences(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	recipient := chain.GenAccount()

	amountToSendFromMultisigAccount := int64(1000)
	amountToSendFromSigner1Account := int64(2000)

	multisigPublicKey, keyNamesSet, err := chain.GenMultisigAccount(3, 2)
	requireT.NoError(err)
	multisigAddress := sdk.AccAddress(multisigPublicKey.Address())
	signer1KeyName := keyNamesSet[0]
	signer2KeyName := keyNamesSet[1]

	signer1Info, err := chain.ClientContext.Keyring().Key(signer1KeyName)
	requireT.NoError(err)
	signer1Address := signer1Info.GetAddress()

	// fund the multisig account
	chain.FundAccountsWithOptions(ctx, t, multisigAddress, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   sdk.NewInt(amountToSendFromMultisigAccount),
	})

	// fund the sender1 account
	chain.FundAccountsWithOptions(ctx, t, signer1Address, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   sdk.NewInt(amountToSendFromSigner1Account),
	})

	// send a tx from sender1 account so the sequence is increased
	msg := &banktypes.MsgSend{
		FromAddress: signer1Address.String(),
		ToAddress:   chain.GenAccount().String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(amountToSendFromSigner1Account))),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(signer1Address),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(&banktypes.MsgSend{})),
		msg)
	require.NoError(t, err)
	signer1AccInfo, err := client.GetAccountInfo(ctx, chain.ClientContext, signer1Address)
	// signer1 account sequence increased to 1.
	requireT.EqualValues(1, signer1AccInfo.GetSequence())

	// prepare account to be funded from the multisig
	recipientAddr := recipient.String()
	coinsToSendToRecipient := sdk.NewCoins(chain.NewCoin(sdk.NewInt(amountToSendFromMultisigAccount)))

	clientCtx := chain.ClientContext
	// prepare the tx factory to sign with the account number of the multisig account
	multisigAccInfo, err := client.GetAccountInfo(ctx, clientCtx, multisigAddress)
	requireT.NoError(err)
	requireT.EqualValues(0, multisigAccInfo.GetSequence())
	txF := chain.TxFactory().
		WithGas(chain.GasLimitByMsgs(&banktypes.MsgSend{})).
		WithAccountNumber(multisigAccInfo.GetAccountNumber()).
		WithSignMode(sdksigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)

	bankSendMsg := &banktypes.MsgSend{
		FromAddress: multisigAddress.String(),
		ToAddress:   recipientAddr,
		Amount:      coinsToSendToRecipient,
	}

	// sign and submit with the min threshold
	txBuilder, err := txF.BuildUnsignedTx(bankSendMsg)
	requireT.NoError(err)

	// sign from signer1 account
	err = client.Sign(txF, signer1KeyName, txBuilder, false)
	requireT.NoError(err)
	signatures, err := txBuilder.GetTx().GetSignaturesV2()
	requireT.NoError(err)
	requireT.Len(signatures, 1)
	// Even though signer1 account sequence is 1 in multisig it uses sequence for multisig acc which is 0.
	requireT.EqualValues(multisigAccInfo.GetSequence(), signatures[0].Sequence)

	err = client.Sign(txF, signer2KeyName, txBuilder, false)
	requireT.NoError(err)
	signatures, err = txBuilder.GetTx().GetSignaturesV2()
	requireT.NoError(err)
	requireT.Len(signatures, 2)
	requireT.EqualValues(multisigAccInfo.GetSequence(), signatures[1].Sequence)

	multisigTx := createMulisignTx(requireT, txBuilder, multisigAccInfo.GetSequence(), multisigPublicKey)
	encodedTx, err := clientCtx.TxConfig().TxEncoder()(multisigTx)
	requireT.NoError(err)
	result, err := client.BroadcastRawTx(ctx, clientCtx, encodedTx)
	requireT.NoError(err)
	t.Logf("Fully signed tx executed, txHash:%s", result.TxHash)

	bankClient := banktypes.NewQueryClient(clientCtx)
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

	chain.FundAccountsWithOptions(ctx, t, sender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   sdk.NewInt(10),
	})

	clientCtx := chain.ClientContext
	accInfo, err := client.GetAccountInfo(ctx, clientCtx, sender)
	require.NoError(t, err)

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(1))),
	}

	_, err = client.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithSequence(accInfo.GetSequence()+1). // incorrect sequence
			WithAccountNumber(accInfo.GetAccountNumber()).
			WithGas(chain.GasLimitByMsgs(msg)),
		msg)
	require.True(t, sdkerrors.ErrWrongSequence.Is(err))
}

func createMulisignTx(requireT *require.Assertions, txBuilder sdkclient.TxBuilder, accSec uint64, multisigPublicKey *sdkmultisig.LegacyAminoPubKey) authsigning.Tx {
	signs, err := txBuilder.GetTx().GetSignaturesV2()
	requireT.NoError(err)

	multisigSig := multisigtypes.NewMultisig(len(multisigPublicKey.PubKeys))
	for _, sig := range signs {
		requireT.NoError(multisigtypes.AddSignatureV2(multisigSig, sig, multisigPublicKey.GetPubKeys()))
	}

	sigV2 := sdksigning.SignatureV2{
		PubKey:   multisigPublicKey,
		Data:     multisigSig,
		Sequence: accSec,
	}

	requireT.NoError(txBuilder.SetSignatures(sigV2))

	return txBuilder.GetTx()
}
