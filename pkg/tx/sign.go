package tx

import (
	"github.com/cosmos/cosmos-sdk/client"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/pkg/types"
)

// Sign signs transaction
func Sign(clientCtx client.Context, input BaseInput, msg sdk.Msg) (authsigning.Tx, error) {
	signer := input.Signer

	privKey := &cosmossecp256k1.PrivKey{Key: signer.Key}
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set message on tx builder")
	}
	txBuilder.SetGasLimit(input.GasLimit)
	txBuilder.SetMemo(input.Memo)

	if !input.GasPrice.Amount.IsDefault() {
		if err := input.GasPrice.Validate(); err != nil {
			return nil, errors.Wrap(err, "gas price is invalid")
		}

		gasLimit := types.NewInt(int64(input.GasLimit))
		fee := gasLimit.Mul(input.GasPrice.Amount)
		feeCoin := sdk.NewCoin(input.GasPrice.Denom, types.IntToSDK(fee))
		txBuilder.SetFeeAmount(sdk.NewCoins(feeCoin))
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
