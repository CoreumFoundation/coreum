//go:build integrationtests

package modules

import (
	"testing"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	sdkmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	multisigtypes "github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdksigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TODO (wojtek): once we have other coins add test verifying that transaction offering fee in coin other than CORE is rejected

// TestAuthFeeLimits verifies that invalid message gas won't be accepted.
func TestAuthFeeLimits(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	sender := chain.GenAccount()

	maxBlockGas := chain.NetworkConfig.Fee.FeeModel.Params().MaxBlockGas
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   sdk.NewInt(maxBlockGas + 100),
	}))

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(1))),
	}

	gasPriceWithMaxDiscount := chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice.
		Mul(sdk.OneDec().Sub(chain.NetworkConfig.Fee.FeeModel.Params().MaxDiscount))

	// the gas price is too low
	_, err := tx.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(chain.NewDecCoin(gasPriceWithMaxDiscount.QuoInt64(2)).String()),
		msg)
	require.True(t, sdkerrors.ErrInsufficientFee.Is(err))

	// no gas price
	_, err = tx.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(chain.GasLimitByMsgs(msg)).
			WithGasPrices(""),
		msg)
	require.True(t, sdkerrors.ErrInsufficientFee.Is(err))

	// more gas than MaxBlockGas
	_, err = tx.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(uint64(maxBlockGas+1)),
		msg)
	require.Error(t, err)

	// gas equal MaxBlockGas, the tx should pass
	_, err = tx.BroadcastTx(ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().
			WithGas(uint64(maxBlockGas)),
		msg)
	require.NoError(t, err)
}

// TestAuthMultisig tests the cosmos-sdk multisig accounts and API.
func TestAuthMultisig(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)

	signer1KeyInfo, err := chain.ClientContext.Keyring().KeyByAddress(chain.GenAccount())
	requireT.NoError(err)

	signer2KeyInfo, err := chain.ClientContext.Keyring().KeyByAddress(chain.GenAccount())
	requireT.NoError(err)

	signer3KeyInfo, err := chain.ClientContext.Keyring().KeyByAddress(chain.GenAccount())
	requireT.NoError(err)

	recipient := chain.GenAccount()
	amountToSendFromMultisigAccount := int64(1000)

	// generate the keyring and collect the keys to use for the multisig account
	keyNamesSet := []string{signer1KeyInfo.GetName(), signer2KeyInfo.GetName(), signer3KeyInfo.GetName()}
	kr := chain.ClientContext.Keyring()
	publicKeySet := make([]cryptotypes.PubKey, 0, len(keyNamesSet))
	for _, key := range keyNamesSet {
		info, err := kr.Key(key)
		requireT.NoError(err)
		publicKeySet = append(publicKeySet, info.GetPubKey())
	}

	// create multisig account
	const multisigThreshold = 2
	multisigPublicKey := sdkmultisig.NewLegacyAminoPubKey(multisigThreshold, publicKeySet)
	multisigAddress, err := sdk.AccAddressFromHex(multisigPublicKey.Address().String())
	requireT.NoError(err)

	// fund the multisig account
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, multisigAddress, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   sdk.NewInt(amountToSendFromMultisigAccount),
	}))

	// prepare account to be funded from the multisig
	recipientAddr := recipient.String()
	coinsToSendToRecipient := sdk.NewCoins(chain.NewCoin(sdk.NewInt(amountToSendFromMultisigAccount)))

	clientCtx := chain.ClientContext
	// prepare the tx factory to sign with the account seq and number of the multisig account
	multisigAccInfo, err := tx.GetAccountInfo(ctx, clientCtx, multisigAddress)
	requireT.NoError(err)
	txF := chain.TxFactory().
		WithGas(chain.GasLimitByMsgs(&banktypes.MsgSend{})).
		WithAccountNumber(multisigAccInfo.GetAccountNumber()).
		WithSequence(multisigAccInfo.GetSequence()).
		WithKeybase(kr).
		WithSignMode(sdksigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON)

	bankSendMsg := &banktypes.MsgSend{
		FromAddress: multisigAddress.String(),
		ToAddress:   recipientAddr,
		Amount:      coinsToSendToRecipient,
	}
	// sign and submit with just one key to check the tx rejection
	txBuilder, err := txF.BuildUnsignedTx(bankSendMsg)
	requireT.NoError(err)

	err = tx.Sign(txF, signer1KeyInfo.GetName(), txBuilder, false)
	requireT.NoError(err)
	multisigTx := createMulisignTx(requireT, txBuilder, multisigAccInfo.GetSequence(), multisigPublicKey)
	encodedTx, err := clientCtx.TxConfig().TxEncoder()(multisigTx)
	requireT.NoError(err)
	_, err = tx.BroadcastRawTx(ctx, clientCtx, encodedTx)
	requireT.True(sdkerrors.ErrUnauthorized.Is(err))
	logger.Get(ctx).Info("Partially signed tx executed with expected error")

	// sign and submit with the min threshold
	txBuilder, err = txF.BuildUnsignedTx(bankSendMsg)
	requireT.NoError(err)
	err = tx.Sign(txF, signer1KeyInfo.GetName(), txBuilder, false)
	requireT.NoError(err)
	err = tx.Sign(txF, signer2KeyInfo.GetName(), txBuilder, false)
	requireT.NoError(err)
	multisigTx = createMulisignTx(requireT, txBuilder, multisigAccInfo.GetSequence(), multisigPublicKey)
	encodedTx, err = clientCtx.TxConfig().TxEncoder()(multisigTx)
	requireT.NoError(err)
	result, err := tx.BroadcastRawTx(ctx, clientCtx, encodedTx)
	requireT.NoError(err)
	logger.Get(ctx).Info("Fully signed tx executed", zap.String("txHash", result.TxHash))

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

	ctx, chain := integrationtests.NewTestingContext(t)

	sender := chain.GenAccount()

	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, sender, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{&banktypes.MsgSend{}},
		Amount:   sdk.NewInt(10),
	}))

	clientCtx := chain.ClientContext
	accInfo, err := tx.GetAccountInfo(ctx, clientCtx, sender)
	require.NoError(t, err)

	msg := &banktypes.MsgSend{
		FromAddress: sender.String(),
		ToAddress:   sender.String(),
		Amount:      sdk.NewCoins(chain.NewCoin(sdk.NewInt(1))),
	}

	_, err = tx.BroadcastTx(ctx,
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
