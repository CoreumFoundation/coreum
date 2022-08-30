package tx

import (
	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/pkg/errors"
)

// Sign signs transaction
func Sign(clientCtx client.Context, input BaseInput, msgs ...sdk.Msg) (authsigning.Tx, error) {
	signer := input.Signer

	privKey := &cosmossecp256k1.PrivKey{Key: signer.Key}
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set message on tx builder")
	}
	txBuilder.SetGasLimit(input.GasLimit)
	txBuilder.SetMemo(input.Memo)

	if input.GasPrice.Amount != nil {
		if err := input.GasPrice.Validate(); err != nil {
			return nil, errors.Wrap(err, "gas price is invalid")
		}

		gasLimit := sdk.NewInt(int64(input.GasLimit))
		gasPrice := sdk.NewIntFromBigInt(input.GasPrice.Amount)
		fee := sdk.NewCoin(input.GasPrice.Denom, gasLimit.Mul(gasPrice))
		txBuilder.SetFeeAmount(sdk.NewCoins(fee))
	}

	signerData := authsigning.SignerData{
		ChainID:       clientCtx.ChainID,
		AccountNumber: signer.AccountNumber,
		Sequence:      signer.AccountSequence,
	}
	sigData := &signing.SingleSignatureData{
		//nolint:nosnakecase // MixedCap can't be forced on imported constants
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   privKey.PubKey(),
		Data:     sigData,
		Sequence: signer.AccountSequence,
	}
	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set signature on tx builder")
	}

	//nolint:nosnakecase // MixedCap can't be forced on imported constants
	bytesToSign, err := clientCtx.TxConfig.SignModeHandler().GetSignBytes(signing.SignMode_SIGN_MODE_DIRECT, signerData, txBuilder.GetTx())
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode bytes to sign")
	}
	sigBytes, err := privKey.Sign(bytesToSign)
	if err != nil {
		return nil, errors.Wrap(err, "unable to sign")
	}

	sigData.Signature = sigBytes

	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set signature on tx builder")
	}

	return txBuilder.GetTx(), nil
}

// BuildSimTx creates an unsigned tx with an empty single signature and returns
// the encoded transaction or an error if the unsigned transaction cannot be
// built.
func BuildSimTx(clientCtx client.Context, base BaseInput, msgs ...sdk.Msg) ([]byte, error) {
	factory := new(clienttx.Factory).
		WithTxConfig(clientCtx.TxConfig).
		WithChainID(clientCtx.ChainID).
		WithGasPrices(base.GasPrice.String()).
		WithMemo(base.Memo).
		//nolint:nosnakecase // MixedCap can't be forced on imported constants
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	txb, err := factory.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// not pass the private key around.
	var pubKey cryptotypes.PubKey = &cosmossecp256k1.PubKey{
		Key: base.Signer.Key.PubKey(),
	}

	// Create an empty signature literal as the ante handler will populate with a
	// sentinel pubkey.
	sig := signing.SignatureV2{
		PubKey: pubKey,
		Data: &signing.SingleSignatureData{
			SignMode: factory.SignMode(),
		},

		Sequence: base.Signer.AccountSequence,
	}
	if err := txb.SetSignatures(sig); err != nil {
		return nil, errors.WithStack(err)
	}

	return clientCtx.TxConfig.TxEncoder()(txb.GetTx())
}
