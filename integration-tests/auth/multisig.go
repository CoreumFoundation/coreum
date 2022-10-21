package auth

import (
	"context"

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
	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TestMultisig tests the cosmos-sdk multisig accounts and API.
func TestMultisig(ctx context.Context, t testing.T, chain testing.Chain) {
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
	require.NoError(t, chain.Faucet.FundAccountsWithOptions(ctx, multisigAddress, testing.BalancesOptions{
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
		WithSignMode(sdksigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON) //nolint:nosnakecase // the sdk constant

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
