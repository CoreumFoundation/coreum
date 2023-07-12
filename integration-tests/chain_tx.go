package integrationtests

import (
	"context"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/v2/pkg/client"
)

// BroadcastTxWithSigner prepares the tx with the provided signer address and broadcasts it.
// The main difference from the client.BroadcastTx is that this function uses the custom account addresses decoding with
// the custom chain prefixes, which allows to execute transactions for different chains.
func (c ChainContext) BroadcastTxWithSigner(ctx context.Context, txf client.Factory, signerAddress sdk.AccAddress, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	clientCtx := c.ClientContext.WithFromAddress(signerAddress)

	// add account info
	txf, err := addAccountInfoToTxFactory(ctx, clientCtx, txf, c.ConvertToBech32Address(signerAddress))
	if err != nil {
		return nil, err
	}

	// estimate gas and add adjustment
	if txf.SimulateAndExecute() {
		_, gas, err := client.CalculateGas(ctx, clientCtx, txf, msgs...)
		if err != nil {
			return nil, err
		}
		txf = txf.WithGas(gas)
	}
	if txf.GasAdjustment() != 0 {
		gas := uint64(txf.GasAdjustment() * float64(txf.Gas()))
		txf = txf.WithGas(gas)
	}

	unsignedTx, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	err = sign(clientCtx, txf, signerAddress, unsignedTx)
	if err != nil {
		return nil, err
	}

	txBytes, err := clientCtx.TxConfig().TxEncoder()(unsignedTx.GetTx())
	if err != nil {
		return nil, err
	}

	return client.BroadcastRawTx(ctx, clientCtx, txBytes)
}

func addAccountInfoToTxFactory(ctx context.Context, clientCtx client.Context, txf tx.Factory, address string) (client.Factory, error) {
	if txf.AccountNumber() == 0 && txf.Sequence() == 0 {
		req := &authtypes.QueryAccountRequest{
			Address: address,
		}
		authQueryClient := authtypes.NewQueryClient(clientCtx)
		res, err := authQueryClient.Account(ctx, req)
		if err != nil {
			return client.Factory{}, errors.WithStack(err)
		}

		var acc authtypes.AccountI
		if err := clientCtx.InterfaceRegistry().UnpackAny(res.Account, &acc); err != nil {
			return client.Factory{}, errors.WithStack(err)
		}

		txf = txf.
			WithAccountNumber(acc.GetAccountNumber()).
			WithSequence(acc.GetSequence())
	}

	return txf, nil
}

func sign(clientCtx client.Context, txf client.Factory, signerAddress sdk.AccAddress, txBuilder sdkclient.TxBuilder) error {
	signMode := txf.SignMode()
	if signMode == signing.SignMode_SIGN_MODE_UNSPECIFIED {
		// use the SignModeHandler's default mode if unspecified
		signMode = clientCtx.TxConfig().SignModeHandler().DefaultMode()
	}

	key, err := txf.Keybase().KeyByAddress(signerAddress)
	if err != nil {
		return err
	}
	pubKey := key.GetPubKey()
	signerData := authsigning.SignerData{
		ChainID:       txf.ChainID(),
		AccountNumber: txf.AccountNumber(),
		Sequence:      txf.Sequence(),
	}

	sigData := signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: txf.Sequence(),
	}
	var prevSignatures []signing.SignatureV2
	if err := txBuilder.SetSignatures(sig); err != nil {
		return err
	}

	// Generate the bytes to be signed.
	bytesToSign, err := clientCtx.TxConfig().SignModeHandler().GetSignBytes(signMode, signerData, txBuilder.GetTx())
	if err != nil {
		return err
	}

	// Sign those bytes
	sigBytes, _, err := txf.Keybase().SignByAddress(signerAddress, bytesToSign)
	if err != nil {
		return err
	}

	// Construct the SignatureV2 struct
	sigData = signing.SingleSignatureData{
		SignMode:  signMode,
		Signature: sigBytes,
	}
	sig = signing.SignatureV2{
		PubKey:   pubKey,
		Data:     &sigData,
		Sequence: txf.Sequence(),
	}

	prevSignatures = append(prevSignatures, sig)
	return txBuilder.SetSignatures(prevSignatures...)
}
