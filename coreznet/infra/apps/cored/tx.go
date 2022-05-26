package cored

import (
	"fmt"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/cosmos/cosmos-sdk/client"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// NewTxManager creates new tx manager
func NewTxManager(clientCtx client.Context) TxManager {
	return TxManager{
		clientCtx: clientCtx,
	}
}

// TxManager builds and broadcasts transactions
type TxManager struct {
	clientCtx client.Context
}

// Sign takes message, creates transaction and signs it
func (txm TxManager) Sign(signerKey Secp256k1PrivateKey, accNum, accSeq uint64, msg sdk.Msg) authsigning.Tx {
	privKey := &cosmossecp256k1.PrivKey{Key: signerKey}
	txBuilder := txm.clientCtx.TxConfig.NewTxBuilder()
	txBuilder.SetGasLimit(200000)
	must.OK(txBuilder.SetMsgs(msg))

	signerData := authsigning.SignerData{
		ChainID:       txm.clientCtx.ChainID,
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

	bytesToSign := must.Bytes(txm.clientCtx.TxConfig.SignModeHandler().GetSignBytes(signing.SignMode_SIGN_MODE_DIRECT, signerData, txBuilder.GetTx()))
	sigBytes, err := privKey.Sign(bytesToSign)
	must.OK(err)

	sigData.Signature = sigBytes

	must.OK(txBuilder.SetSignatures(sig))

	return txBuilder.GetTx()
}

// Broadcast broadcasts message and returns tx hash
func (txm TxManager) Broadcast(signerKey Secp256k1PrivateKey, msg sdk.Msg) (string, error) {
	signerAddress, err := sdk.AccAddressFromBech32(signerKey.Address())
	must.OK(err)
	accNum, accSeq, err := txm.clientCtx.AccountRetriever.GetAccountNumberSequence(txm.clientCtx, signerAddress)
	if err != nil {
		return "", err
	}

	// FIXME (wojciech): Find a way to exit from this function early if ctx is canceled
	txResp, err := txm.clientCtx.BroadcastTxCommit(must.Bytes(txm.clientCtx.TxConfig.TxEncoder()(txm.Sign(signerKey, accNum, accSeq, msg))))
	if err != nil {
		return "", err
	}

	if txResp.Code != 0 {
		return "", fmt.Errorf("trasaction failed: %s", txResp.RawLog)
	}
	return txResp.TxHash, nil
}
