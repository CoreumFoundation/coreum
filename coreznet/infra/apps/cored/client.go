package cored

import (
	"context"
	"fmt"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/cosmos/cosmos-sdk/client"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// NewClient creates new client for cored
func NewClient(chainID string, addr string) Client {
	rpcClient, err := client.NewClientFromNode("tcp://" + addr)
	must.OK(err)
	clientCtx := NewContext(chainID, rpcClient)
	return Client{
		clientCtx:       clientCtx,
		bankQueryClient: banktypes.NewQueryClient(clientCtx),
	}
}

// Client is the client for cored blockchain
type Client struct {
	clientCtx       client.Context
	bankQueryClient banktypes.QueryClient
}

// GetNumberSequence returns account number and account sequence for provided address
func (c Client) GetNumberSequence(address string) (uint64, uint64, error) {
	addr, err := sdk.AccAddressFromBech32(address)
	must.OK(err)
	return c.clientCtx.AccountRetriever.GetAccountNumberSequence(c.clientCtx, addr)
}

// QBankBalances queries for bank balances owned by wallet
func (c Client) QBankBalances(ctx context.Context, wallet Wallet) (map[string]Balance, error) {
	// FIXME (wojtek): support pagination
	resp, err := c.bankQueryClient.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{Address: wallet.Key.Address()})
	if err != nil {
		return nil, err
	}

	balances := map[string]Balance{}
	for _, b := range resp.Balances {
		balances[b.Denom] = Balance{Amount: b.Amount.BigInt(), Denom: b.Denom}
	}
	return balances, nil
}

// Sign takes message, creates transaction and signs it
func (c Client) Sign(signerKey Secp256k1PrivateKey, accNum, accSeq uint64, msg sdk.Msg) (authsigning.Tx, error) {
	if accNum == 0 || accSeq == 0 {
		accNum2, accSeq2, err := c.GetNumberSequence(signerKey.Address())
		if err != nil {
			return nil, err
		}
		if accNum == 0 {
			accNum = accNum2
		}
		if accSeq == 0 {
			accSeq = accSeq2
		}
	}

	return signTx(c.clientCtx, signerKey, accNum, accSeq, msg), nil
}

// Encode encodes transaction to be broadcasted
func (c Client) Encode(signedTx authsigning.Tx) []byte {
	return must.Bytes(c.clientCtx.TxConfig.TxEncoder()(signedTx))
}

// Broadcast broadcasts encoded transaction and returns tx hash
func (c Client) Broadcast(encodedTx []byte) (string, error) {
	// FIXME (wojciech): Find a way to exit from this function early if ctx is canceled
	txResp, err := c.clientCtx.BroadcastTxCommit(encodedTx)
	if err != nil {
		return "", err
	}

	if txResp.Code != 0 {
		return "", fmt.Errorf("trasaction failed: %s", txResp.RawLog)
	}
	return txResp.TxHash, nil
}

// TxBankSend creates a transaction sending tokens from one wallet to another
func (c Client) TxBankSend(sender, receiver Wallet, balance Balance) ([]byte, error) {
	fromAddress, err := sdk.AccAddressFromBech32(sender.Key.Address())
	must.OK(err)
	toAddress, err := sdk.AccAddressFromBech32(receiver.Key.Address())
	must.OK(err)

	signedTx, err := c.Sign(sender.Key, sender.AccountNumber, sender.AccountSequence, banktypes.NewMsgSend(fromAddress, toAddress, sdk.Coins{
		{
			Denom:  balance.Denom,
			Amount: sdk.NewIntFromBigInt(balance.Amount),
		},
	}))
	if err != nil {
		return nil, err
	}

	return c.Encode(signedTx), nil
}

func signTx(clientCtx client.Context, signerKey Secp256k1PrivateKey, accNum, accSeq uint64, msg sdk.Msg) authsigning.Tx {
	privKey := &cosmossecp256k1.PrivKey{Key: signerKey}
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	txBuilder.SetGasLimit(200000)
	must.OK(txBuilder.SetMsgs(msg))

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
	must.OK(txBuilder.SetSignatures(sig))

	bytesToSign := must.Bytes(clientCtx.TxConfig.SignModeHandler().GetSignBytes(signing.SignMode_SIGN_MODE_DIRECT, signerData, txBuilder.GetTx()))
	sigBytes, err := privKey.Sign(bytesToSign)
	must.OK(err)

	sigData.Signature = sigBytes

	must.OK(txBuilder.SetSignatures(sig))

	return txBuilder.GetTx()
}
