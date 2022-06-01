package cored

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/cosmos/cosmos-sdk/client"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"
)

// NewClient creates new client for cored
func NewClient(chainID string, addr string) Client {
	rpcClient, err := client.NewClientFromNode("tcp://" + addr)
	must.OK(err)

	clientCtx := NewContext(chainID, rpcClient)

	return &coreClient{
		clientCtx: clientCtx,
		clientConfig: ClientConfig{
			BroadcastTimeout:    20 * time.Second,
			BroadcastStatusPoll: 250 * time.Millisecond,
		},

		bankQueryClient: banktypes.NewQueryClient(clientCtx),
	}
}

// Client is the client interface for cored blockchain
type Client interface {
	GetNumberSequence(address string) (uint64, uint64, error)
	QBankBalances(ctx context.Context, wallet Wallet) (map[string]Balance, error)
	Sign(signer Wallet, msg sdk.Msg) (authsigning.Tx, error)
	Encode(signedTx authsigning.Tx) []byte
	Broadcast(ctx context.Context, encodedTx []byte) (*sdk.TxResponse, error)
	TxBankSend(sender, receiver Wallet, balance Balance) ([]byte, error)
}

type ClientConfig struct {
	BroadcastTimeout    time.Duration
	BroadcastStatusPoll time.Duration
}

type coreClient struct {
	clientCtx    client.Context
	clientConfig ClientConfig

	bankQueryClient banktypes.QueryClient
}

// GetNumberSequence returns account number and account sequence for provided address
func (c *coreClient) GetNumberSequence(address string) (uint64, uint64, error) {
	addr, err := sdk.AccAddressFromBech32(address)
	must.OK(err)
	return c.clientCtx.AccountRetriever.GetAccountNumberSequence(c.clientCtx, addr)
}

// QBankBalances queries for bank balances owned by wallet
func (c *coreClient) QBankBalances(ctx context.Context, wallet Wallet) (map[string]Balance, error) {
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
func (c *coreClient) Sign(signer Wallet, msg sdk.Msg) (authsigning.Tx, error) {
	if signer.AccountNumber == 0 && signer.AccountSequence == 0 {
		var err error
		signer.AccountNumber, signer.AccountSequence, err = c.GetNumberSequence(signer.Key.Address())
		if err != nil {
			return nil, err
		}
	}

	return signTx(c.clientCtx, signer.Key, signer.AccountNumber, signer.AccountSequence, msg), nil
}

// Encode encodes transaction to be broadcasted
func (c *coreClient) Encode(signedTx authsigning.Tx) []byte {
	return must.Bytes(c.clientCtx.TxConfig.TxEncoder()(signedTx))
}

var ErrTxTimedOut = errors.New("transaction broadcast timed out")

// Broadcast broadcasts encoded transaction and returns tx hash
func (c *coreClient) Broadcast(ctx context.Context, encodedTx []byte) (*sdk.TxResponse, error) {
	res, err := c.clientCtx.BroadcastTxSync(encodedTx)
	if err != nil {
		return res, err
	}

	awaitCtx, cancelFn := context.WithTimeout(ctx, c.clientConfig.BroadcastTimeout)
	defer cancelFn()

	txHash, _ := hex.DecodeString(res.TxHash)
	t := time.NewTimer(c.clientConfig.BroadcastStatusPoll)

	for {
		select {
		case <-awaitCtx.Done():
			err := errors.Wrapf(ErrTxTimedOut, "%s", res.TxHash)
			t.Stop()
			return nil, err
		case <-t.C:
			resultTx, err := c.clientCtx.Client.Tx(awaitCtx, txHash, false)
			if err != nil {
				if errRes := client.CheckTendermintError(err, encodedTx); errRes != nil {
					return errRes, err
				}

				t.Reset(c.clientConfig.BroadcastStatusPoll)
				continue

			} else if resultTx.Height > 0 {
				res = sdk.NewResponseResultTx(resultTx, res.Tx, res.Timestamp)
				t.Stop()
				return res, err
			}

			t.Reset(c.clientConfig.BroadcastStatusPoll)
		}
	}
}

// TxBankSend creates a transaction sending tokens from one wallet to another
func (c *coreClient) TxBankSend(sender, receiver Wallet, balance Balance) ([]byte, error) {
	fromAddress, err := sdk.AccAddressFromBech32(sender.Key.Address())
	must.OK(err)
	toAddress, err := sdk.AccAddressFromBech32(receiver.Key.Address())
	must.OK(err)

	signedTx, err := c.Sign(sender, banktypes.NewMsgSend(fromAddress, toAddress, sdk.Coins{
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
