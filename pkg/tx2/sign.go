package tx2

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/pkg/errors"
)

// Sign signs transaction for being broadcasted
func Sign(clientCtx client.Context, input BaseInput, msgs ...sdk.Msg) (authsigning.Tx, error) {
	signer := input.Signer

	factory := new(tx.Factory).
		WithTxConfig(clientCtx.TxConfig).
		WithChainID(clientCtx.ChainID).
		WithGas(input.GasLimit).
		WithMemo(input.Memo).
		//nolint:nosnakecase // MixedCap can't be forced on imported constants
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	if input.GasPrice.Amount != nil {
		if err := input.GasPrice.Validate(); err != nil {
			return nil, errors.Wrap(err, "gas price is invalid")
		}
		factory = factory.WithGasPrices(input.GasPrice.String())
	}

	txBuilder, err := factory.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	privKey := &cosmossecp256k1.PrivKey{Key: signer.PrivateKey}

	signerData := authsigning.SignerData{
		ChainID:       clientCtx.ChainID,
		AccountNumber: signer.Account.Number,
		Sequence:      signer.Account.Sequence,
	}
	sigData := &signing.SingleSignatureData{
		SignMode:  factory.SignMode(),
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   privKey.PubKey(),
		Data:     sigData,
		Sequence: signer.Account.Sequence,
	}

	if err := txBuilder.SetSignatures(sig); err != nil {
		return nil, errors.Wrap(err, "unable to set signature on tx builder")
	}

	// If private key is empty it means transaction is "signed" only for simulation
	if signer.PrivateKey != nil {
		bytesToSign, err := clientCtx.TxConfig.SignModeHandler().GetSignBytes(factory.SignMode(), signerData, txBuilder.GetTx())
		if err != nil {
			return nil, errors.Wrap(err, "unable to encode bytes to sign")
		}
		sigBytes, err := privKey.Sign(bytesToSign)
		if err != nil {
			return nil, errors.Wrap(err, "unable to sign")
		}

		sigData.Signature = sigBytes

		if err := txBuilder.SetSignatures(sig); err != nil {
			return nil, errors.Wrap(err, "unable to set signature on tx builder")
		}
	}

	return txBuilder.GetTx(), nil
}
