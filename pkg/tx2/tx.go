package tx

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

// TxConfig holds values common to every transaction
type TxConfig struct {
	From     sdk.AccAddress
	PrivKey  cryptotypes.PrivKey
	GasLimit uint64
	GasPrice *sdk.Coin
	Memo     string

	fromAccount *accountInfo
}

// SetAccountNumber allows to set account number explicitly
func (c *TxConfig) SetAccountNumber(n uint64) {
	if c.fromAccount == nil {
		c.fromAccount = &accountInfo{}
	}

	c.fromAccount.Number = n
}

// SetAccountSequence allows to set account sequence explicitly
func (c *TxConfig) SetAccountSequence(seq uint64) {
	if c.fromAccount == nil {
		c.fromAccount = &accountInfo{}
	}

	c.fromAccount.Sequence = seq
}

// accountInfo stores account number and sequence
type accountInfo struct {
	Number   uint64
	Sequence uint64
}

// Sign signs transaction with provided priv key and config.
func Sign(
	clientCtx client.Context,
	config TxConfig,
	msgs ...sdk.Msg,
) (authsigning.Tx, error) {
	if config.PrivKey == nil {
		err := errors.New("sign is required but no privkey provided")
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
		AccountNumber: config.fromAccount.Number,
		Sequence:      config.fromAccount.Sequence,
	}
	sigData := &signing.SingleSignatureData{
		SignMode:  factory.SignMode(),
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   config.PrivKey.PubKey(),
		Data:     sigData,
		Sequence: config.fromAccount.Sequence,
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

	sigBytes, err := config.PrivKey.Sign(bytesToSign)
	if err != nil {
		return nil, errors.Wrap(err, "unable to sign using priv key")
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

		Sequence: config.fromAccount.Sequence,
	}
	if err := txb.SetSignatures(sig); err != nil {
		return nil, err
	}

	return clientCtx.TxConfig.TxEncoder()(txb.GetTx())
}
