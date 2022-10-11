package tx

import (
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/pkg/errors"
)

// Sign signs transaction
// Deprecated: Use the SignTx instead.
func Sign(clientCtx ClientContext, input BaseInput, msgs ...sdk.Msg) (authsigning.Tx, error) {
	signer := input.Signer

	txBuilder := clientCtx.TxConfig().NewTxBuilder()
	err := txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set message on tx builder")
	}
	txBuilder.SetGasLimit(input.GasLimit)
	txBuilder.SetMemo(input.Memo)

	if !input.GasPrice.Amount.IsNil() {
		if err := input.GasPrice.Validate(); err != nil {
			return nil, errors.Wrap(err, "gas price is invalid")
		}

		gasLimit := sdk.NewInt(int64(input.GasLimit))

		// Ceil().RoundInt() is here to be compatible with the sdk's TxFactory
		// https://github.com/cosmos/cosmos-sdk/blob/ff416ee63d32da5d520a8b2d16b00da762416146/client/tx/factory.go#L223
		fee := sdk.NewCoin(input.GasPrice.Denom, input.GasPrice.Amount.Mul(gasLimit.ToDec()).Ceil().RoundInt())
		txBuilder.SetFeeAmount(sdk.NewCoins(fee))
	}

	signerData := authsigning.SignerData{
		ChainID:       clientCtx.ChainID(),
		AccountNumber: signer.AccountNumber,
		Sequence:      signer.AccountSequence,
	}
	sigData := &signing.SingleSignatureData{
		//nolint:nosnakecase // MixedCap can't be forced on imported constants
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   signer.Key.PubKey(),
		Data:     sigData,
		Sequence: signer.AccountSequence,
	}
	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set signature on tx builder")
	}

	//nolint:nosnakecase // MixedCap can't be forced on imported constants
	bytesToSign, err := clientCtx.TxConfig().SignModeHandler().GetSignBytes(signing.SignMode_SIGN_MODE_DIRECT, signerData, txBuilder.GetTx())
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode bytes to sign")
	}
	sigBytes, err := signer.Key.Sign(bytesToSign)
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
func BuildSimTx(clientCtx ClientContext, base BaseInput, msgs ...sdk.Msg) ([]byte, error) {
	factory := new(clienttx.Factory).
		WithTxConfig(clientCtx.TxConfig()).
		WithChainID(clientCtx.ChainID()).
		WithGasPrices(base.GasPrice.String()).
		WithMemo(base.Memo).
		//nolint:nosnakecase // MixedCap can't be forced on imported constants
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	txb, err := factory.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Create an empty signature literal as the ante handler will populate with a
	// sentinel pubkey.
	sig := signing.SignatureV2{
		PubKey: base.Signer.Key.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode: factory.SignMode(),
		},

		Sequence: base.Signer.AccountSequence,
	}
	if err := txb.SetSignatures(sig); err != nil {
		return nil, errors.WithStack(err)
	}

	return clientCtx.TxConfig().TxEncoder()(txb.GetTx())
}
