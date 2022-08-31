package tx

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/pkg/errors"
)

// Sign signs transaction
func Sign(clientCtx client.Context, input SignInput, msgs ...sdk.Msg) (authsigning.Tx, error) {
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msgs...)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set message on tx builder")
	}
	txBuilder.SetGasLimit(input.GasLimit)
	txBuilder.SetMemo(input.Memo)

	if !input.GasPrice.Amount.IsNil() {
		gasLimit := sdk.NewInt(int64(input.GasLimit))
		gasPrice := input.GasPrice.Amount
		fee := sdk.NewCoin(input.GasPrice.Denom, gasLimit.Mul(gasPrice))
		txBuilder.SetFeeAmount(sdk.NewCoins(fee))
	}

	signerData := authsigning.SignerData{
		ChainID:       clientCtx.ChainID,
		AccountNumber: input.AccountInfo.Number,
		Sequence:      input.AccountInfo.Sequence,
	}
	sigData := &signing.SingleSignatureData{
		//nolint:nosnakecase // MixedCap can't be forced on imported constants
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   input.PrivateKey.PubKey(),
		Data:     sigData,
		Sequence: input.AccountInfo.Sequence,
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
	sigBytes, err := input.PrivateKey.Sign(bytesToSign)
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
