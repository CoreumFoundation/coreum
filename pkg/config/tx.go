package config

import (
	"crypto/ed25519"

	"github.com/CoreumFoundation/coreum/pkg/types"
	"github.com/cosmos/cosmos-sdk/client"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
)

// GenerateAddValidatorTx generates transaction of type MsgCreateValidator
func GenerateAddValidatorTx(
	clientCtx client.Context,
	validatorPublicKey ed25519.PublicKey,
	stakerPrivateKey types.Secp256k1PrivateKey,
	stakedBalance string,
) ([]byte, error) {
	amount, err := sdk.ParseCoinNormalized(stakedBalance)
	if err != nil {
		return nil, errors.Wrapf(err, "not able to parse stake balances %s", stakedBalance)
	}

	commission := stakingtypes.CommissionRates{
		Rate:          sdk.MustNewDecFromStr("0.1"),
		MaxRate:       sdk.MustNewDecFromStr("0.2"),
		MaxChangeRate: sdk.MustNewDecFromStr("0.01"),
	}

	valPubKey := &cosmosed25519.PubKey{Key: validatorPublicKey}
	stakerPrivKey := &cosmossecp256k1.PrivKey{Key: stakerPrivateKey}
	stakerAddress := sdk.AccAddress(stakerPrivKey.PubKey().Address())

	msg, err := stakingtypes.NewMsgCreateValidator(sdk.ValAddress(stakerAddress), valPubKey, amount, stakingtypes.Description{Moniker: stakerAddress.String()}, commission, sdk.OneInt())
	if err != nil {
		return nil, errors.Wrap(err, "not able to make CreateValidatorMessage")
	}

	tx, err := signTx(clientCtx, stakerPrivateKey, 0, 0, msg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to sign transaction")
	}
	encodedTx, err := clientCtx.TxConfig.TxJSONEncoder()(tx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode transaction")
	}
	return encodedTx, nil
}

func signTx(clientCtx client.Context, signerKey types.Secp256k1PrivateKey, accNum, accSeq uint64, msg sdk.Msg) (authsigning.Tx, error) {
	privKey := &cosmossecp256k1.PrivKey{Key: signerKey}
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	txBuilder.SetGasLimit(200000)
	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set message on tx builder")
	}

	signerData := authsigning.SignerData{
		ChainID:       clientCtx.ChainID,
		AccountNumber: accNum,
		Sequence:      accSeq,
	}
	sigData := &signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   privKey.PubKey(),
		Data:     sigData,
		Sequence: accSeq,
	}
	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set signature on tx builder")
	}

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
