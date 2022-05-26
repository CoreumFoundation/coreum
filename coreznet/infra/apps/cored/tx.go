package cored

import (
	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// NewTxConfig returns new configuration required by cosmos to build, sign and encode transactions
func NewTxConfig(marshaler *codec.ProtoCodec) client.TxConfig {
	return tx.NewTxConfig(marshaler, []signing.SignMode{signing.SignMode_SIGN_MODE_DIRECT})
}

// NewTxBuilder creates new tx builder
func NewTxBuilder(chainID string) *TxBuilder {
	marshaler := NewCodec()
	return &TxBuilder{
		chainID:   chainID,
		marshaler: marshaler,
		txConfig:  NewTxConfig(marshaler),
	}
}

// TxBuilder builds transactions
type TxBuilder struct {
	chainID   string
	marshaler *codec.ProtoCodec
	txConfig  client.TxConfig
}

// Sign takes message, creates transaction and signs it
func (txb *TxBuilder) Sign(signerKey Secp256k1PrivateKey, accNum, accSeq uint64, msg sdk.Msg) authsigning.Tx {
	privKey := &cosmossecp256k1.PrivKey{Key: signerKey}
	txBuilder := txb.txConfig.NewTxBuilder()
	txBuilder.SetGasLimit(200000)
	must.OK(txBuilder.SetMsgs(msg))

	signerData := authsigning.SignerData{
		ChainID:       txb.chainID,
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
	must.OK(txBuilder.SetSignatures(sig))

	bytesToSign := must.Bytes(txb.txConfig.SignModeHandler().GetSignBytes(signing.SignMode_SIGN_MODE_DIRECT, signerData, txBuilder.GetTx()))
	sigBytes, err := privKey.Sign(bytesToSign)
	must.OK(err)

	sigData.Signature = sigBytes

	must.OK(txBuilder.SetSignatures(sig))

	return txBuilder.GetTx()
}
