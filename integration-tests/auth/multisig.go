package auth

import (
	"context"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdktx "github.com/cosmos/cosmos-sdk/client/tx"
	sdkhd "github.com/cosmos/cosmos-sdk/crypto/hd"
	sdkkeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdkmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	multisigtypes "github.com/cosmos/cosmos-sdk/crypto/types/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdksigning "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/go-bip39"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// TestMultisig tests the cosmos-sdk multisig accounts and API.
func TestMultisig(ctx context.Context, t testing.T, chain testing.Chain) { //nolint:funlen // The test covers step-by step use case, no need split it
	const (
		key1 = "key1"
		key2 = "key2"
		key3 = "key3"
	)

	faucetWallet := testing.RandomWallet()
	recipientWallet := testing.RandomWallet()
	nativeDenom := chain.NetworkConfig.TokenSymbol
	initialGasPrice := chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice
	bankSendGas := chain.NetworkConfig.Fee.DeterministicGas.BankSend

	amountToSendFromMultisigAccount := int64(1000)

	requireT := require.New(t)
	requireT.NoError(chain.Faucet.FundAccounts(ctx,
		testing.FundedAccount{
			// TODO (dhil): the test uses the faucetWallet since the FundedAccount consumes the Wallet instead of the address.
			// Once start using the sdk types directly this code will be refactored, and multisig account will be funded directly.
			Wallet: faucetWallet,
			Amount: testing.MustNewCoin(t,
				testing.ComputeNeededBalance(
					initialGasPrice,
					bankSendGas,
					2,
					sdk.NewInt(amountToSendFromMultisigAccount)),
				nativeDenom,
			),
		},
	))

	// generate the keyring and collect the keys to use for the multisig account
	keyNamesSet := []string{key1, key2, key3}
	kr := sdkkeyring.NewInMemory()
	publicKeySet := make([]cryptotypes.PubKey, 0, len(keyNamesSet))
	for _, key := range keyNamesSet {
		info := addRandomAccountToKeyring(requireT, key, kr)
		publicKeySet = append(publicKeySet, info.GetPubKey())
	}

	// create multisig account
	const multisigThreshold = 2
	multisigPublicKey := sdkmultisig.NewLegacyAminoPubKey(multisigThreshold, publicKeySet)
	multisigAddress, err := sdk.AccAddressFromHex(multisigPublicKey.Address().String())

	// fund the multisig account
	coredClient := chain.Client
	gasPrice := testing.MustNewCoin(t, chain.NetworkConfig.Fee.FeeModel.Params().InitialGasPrice, nativeDenom)
	requireT.NoError(err)
	coinsToFundMultisigAddress := sdk.NewCoins(sdk.NewCoin(nativeDenom, testing.ComputeNeededBalance(
		initialGasPrice,
		bankSendGas,
		1, sdk.NewInt(amountToSendFromMultisigAccount))))

	bankSendTx, err := coredClient.Sign(
		ctx,
		tx.BaseInput{
			Signer:   faucetWallet,
			GasPrice: gasPrice,
			GasLimit: bankSendGas,
		},
		banktypes.NewMsgSend(
			faucetWallet.Address(),
			multisigAddress,
			coinsToFundMultisigAddress),
	)
	requireT.NoError(err)
	// TODO (dhil) replace to new Broadcast once we finish with it
	_, err = coredClient.Broadcast(ctx, coredClient.Encode(bankSendTx))
	requireT.NoError(err)

	multisigBalances, err := coredClient.BankQueryClient().AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: multisigAddress.String(),
	})
	requireT.NoError(err)
	requireT.Equal(coinsToFundMultisigAddress, multisigBalances.Balances)

	// prepare account to be funded from the multisig
	recipientAddr := recipientWallet.Address()
	coinsToSendToRecipient := sdk.NewCoins(sdk.NewInt64Coin(nativeDenom, 1000))

	// TODO (dhil): this will be refactored once we migrate fully to the new tx package
	clientCtx := coredClient.GetClientCtx().
		WithBroadcastMode(flags.BroadcastBlock)

	// prepare the tx factory to sign with the account seq and number of the multisig account
	multisigAccInfo, err := tx.GetAccountInfo(ctx, clientCtx, multisigAddress)
	requireT.NoError(err)
	txF := sdktx.Factory{}. // TODO (dhil) move/use this code to/from helpers after the migration to the ne tx package
				WithAccountNumber(multisigAccInfo.GetAccountNumber()).
				WithSequence(multisigAccInfo.GetSequence()).
				WithChainID(string(chain.NetworkConfig.ChainID)).
				WithKeybase(kr).
				WithTxConfig(clientCtx.TxConfig).
				WithGas(bankSendGas).
				WithGasPrices(sdk.NewCoin(nativeDenom, initialGasPrice).String()).
				WithSignMode(sdksigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON) //nolint:nosnakecase // the sdk constant

	bankSendMsg := banktypes.NewMsgSend(
		multisigAddress,
		recipientAddr,
		coinsToSendToRecipient)

	// sign and submit with just one key to check the tx rejection
	txBuilder, err := txF.BuildUnsignedTx(bankSendMsg)
	requireT.NoError(err)
	err = tx.SignTx(txF, key1, txBuilder, false)
	requireT.NoError(err)
	multisigTx := createMulisignTx(requireT, txBuilder, multisigAccInfo.GetSequence(), multisigPublicKey)
	_, err = tx.BroadcastRawTx(ctx, clientCtx, coredClient.Encode(multisigTx))
	requireT.Error(err)
	require.True(t, client.IsErr(err, sdkerrors.ErrUnauthorized))

	// sign and submit with the min threshold
	txBuilder, err = txF.BuildUnsignedTx(bankSendMsg)
	requireT.NoError(err)
	err = tx.SignTx(txF, key1, txBuilder, false)
	requireT.NoError(err)
	err = tx.SignTx(txF, key2, txBuilder, false)
	requireT.NoError(err)
	multisigTx = createMulisignTx(requireT, txBuilder, multisigAccInfo.GetSequence(), multisigPublicKey)
	_, err = tx.BroadcastRawTx(ctx, clientCtx, coredClient.Encode(multisigTx))
	requireT.NoError(err)

	recipientBalances, err := coredClient.BankQueryClient().AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address: recipientAddr.String(),
	})
	requireT.NoError(err)
	requireT.Equal(coinsToSendToRecipient, recipientBalances.Balances)
}

func addRandomAccountToKeyring(requireT *require.Assertions, name string, kr sdkkeyring.Keyring) sdkkeyring.Info {
	mnemonic := generateRandomMnemonic(requireT)

	accInfo, err := kr.NewAccount(name, mnemonic, "", "", sdkhd.Secp256k1)
	requireT.NoError(err)

	return accInfo
}

func generateRandomMnemonic(requireT *require.Assertions) string {
	const mnemonicEntropySize = 256
	entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
	requireT.NoError(err)

	mnemonic, err := bip39.NewMnemonic(entropySeed)
	requireT.NoError(err)

	return mnemonic
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
