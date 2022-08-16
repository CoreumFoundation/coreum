package client2

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/pkg/errors"
)

// signTx signs transaction with provided keyring and config.
func signTx(
	clientCtx client.Context,
	config TxConfig,
	msgs ...sdk.Msg,
) (authsigning.Tx, error) {
	if config.Keyring == nil {
		err := errors.New("sign is required but no keyring provided")
		return nil, err
	}

	keyInfo, err := config.Keyring.KeyByAddress(config.From)
	if err != nil {
		err = errors.Wrapf(err, "failed to get key info from %s", config.From.String())
		return nil, err
	}

	factory := new(tx.Factory).
		WithTxConfig(clientCtx.TxConfig).
		WithChainID(clientCtx.ChainID).
		WithGas(config.GasLimit).
		WithMemo(config.Memo).
		//nolint:nosnakecase // MixedCap can't be forced on imported constants
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	if config.GasPrice != nil {
		if err := config.GasPrice.Validate(); err != nil {
			return nil, errors.Wrap(err, "gas price is invalid")
		}
		factory = factory.WithGasPrices(config.GasPrice.String())
	}

	txBuilder, err := factory.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	signerData := authsigning.SignerData{
		ChainID:       clientCtx.ChainID,
		AccountNumber: config.FromAccount.Number,
		Sequence:      config.FromAccount.Sequence,
	}
	sigData := &signing.SingleSignatureData{
		SignMode:  factory.SignMode(),
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   keyInfo.GetPubKey(),
		Data:     sigData,
		Sequence: config.FromAccount.Sequence,
	}
	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, errors.Wrap(err, "unable to set signature on tx builder")
	}

	bytesToSign, err := clientCtx.TxConfig.SignModeHandler().GetSignBytes(
		factory.SignMode(),
		signerData,
		txBuilder.GetTx(),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode bytes to sign")
	}

	sigBytes, _, err := config.Keyring.SignByAddress(config.From, bytesToSign)
	if err != nil {
		return nil, errors.Wrap(err, "unable to sign using keyring")
	}

	sigData.Signature = sigBytes

	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, errors.Wrap(err, "unable to set signature on tx builder")
	}

	return txBuilder.GetTx(), nil
}

// buildSimTx creates an unsigned tx with an empty single signature and returns
// the encoded transaction or an error if the unsigned transaction cannot be
// built.
func buildSimTx(
	clientCtx client.Context,
	config TxConfig,
	msgs ...sdk.Msg,
) ([]byte, error) {
	factory := new(tx.Factory).
		WithTxConfig(clientCtx.TxConfig).
		WithChainID(clientCtx.ChainID).
		WithGasPrices(config.GasPrice.String()).
		WithMemo(config.Memo).
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	txb, err := factory.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	// pubKey is a default public key, is not required to be provided
	var pubKey cryptotypes.PubKey = &cosmossecp256k1.PubKey{}

	// Create an empty signature literal as the ante handler will populate with a
	// sentinel pubkey.
	sig := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode: factory.SignMode(),
		},

		Sequence: config.FromAccount.Sequence,
	}
	if err := txb.SetSignatures(sig); err != nil {
		return nil, err
	}

	return clientCtx.TxConfig.TxEncoder()(txb.GetTx())
}
