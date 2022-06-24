package cored

import (
	"context"
	"encoding/hex"
	"regexp"
	"strconv"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/cosmos/cosmos-sdk/client"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/CoreumFoundation/coreum/crust/pkg/retry"
)

const (
	requestTimeout       = 10 * time.Second
	txTimeout            = time.Minute
	txStatusPollInterval = 500 * time.Millisecond
)

var expectedSequenceRegExp = regexp.MustCompile(`account sequence mismatch, expected (\d+), got \d+`)

// NewClient creates new client for cored
func NewClient(chainID string, addr string) Client {
	rpcClient, err := client.NewClientFromNode("tcp://" + addr)
	must.OK(err)
	clientCtx := NewContext(chainID, rpcClient)
	return Client{
		clientCtx:       clientCtx,
		authQueryClient: authtypes.NewQueryClient(clientCtx),
		bankQueryClient: banktypes.NewQueryClient(clientCtx),
	}
}

// Client is the client for cored blockchain
type Client struct {
	clientCtx       client.Context
	authQueryClient authtypes.QueryClient
	bankQueryClient banktypes.QueryClient
}

// GetNumberSequence returns account number and account sequence for provided address
func (c Client) GetNumberSequence(ctx context.Context, address string) (uint64, uint64, error) {
	addr, err := sdk.AccAddressFromBech32(address)
	must.OK(err)

	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	var header metadata.MD
	res, err := c.authQueryClient.Account(requestCtx, &authtypes.QueryAccountRequest{Address: addr.String()}, grpc.Header(&header))
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}

	var acc authtypes.AccountI
	if err := c.clientCtx.InterfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
		return 0, 0, errors.WithStack(err)
	}

	return acc.GetAccountNumber(), acc.GetSequence(), nil
}

// QueryBankBalances queries for bank balances owned by wallet
func (c Client) QueryBankBalances(ctx context.Context, wallet Wallet) (map[string]Balance, error) {
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	// FIXME (wojtek): support pagination
	resp, err := c.bankQueryClient.AllBalances(requestCtx, &banktypes.QueryAllBalancesRequest{Address: wallet.Key.Address()})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	balances := map[string]Balance{}
	for _, b := range resp.Balances {
		balances[b.Denom] = Balance{Amount: b.Amount.BigInt(), Denom: b.Denom}
	}
	return balances, nil
}

// Sign takes message, creates transaction and signs it
func (c Client) Sign(ctx context.Context, signer Wallet, msg sdk.Msg) (authsigning.Tx, error) {
	if signer.AccountNumber == 0 && signer.AccountSequence == 0 {
		var err error
		signer.AccountNumber, signer.AccountSequence, err = c.GetNumberSequence(ctx, signer.Key.Address())
		if err != nil {
			return nil, err
		}
	}

	return signTx(c.clientCtx, signer.Key, signer.AccountNumber, signer.AccountSequence, msg), nil
}

// Encode encodes transaction to be broadcasted
func (c Client) Encode(signedTx authsigning.Tx) []byte {
	return must.Bytes(c.clientCtx.TxConfig.TxEncoder()(signedTx))
}

// Broadcast broadcasts encoded transaction and returns tx hash
func (c Client) Broadcast(ctx context.Context, encodedTx []byte) (string, error) {
	var txHash string
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	res, err := c.clientCtx.Client.BroadcastTxSync(requestCtx, encodedTx)
	// nolint:nestif // This code is still easy to understand
	if err != nil {
		if errors.Is(err, requestCtx.Err()) {
			return "", errors.WithStack(err)
		}

		errRes := client.CheckTendermintError(err, encodedTx)
		if !isTxInMempool(errRes) {
			return "", errors.WithStack(err)
		}
		txHash = errRes.TxHash
	} else {
		txHash = res.Hash.String()
		if res.Code != 0 {
			if err := checkSequence(res.Codespace, res.Code, res.Log); err != nil {
				return "", err
			}
			return "", errors.Errorf("node returned non-zero code for tx '%s' (code: %d, codespace: %s): %s",
				txHash, res.Code, res.Codespace, res.Log)
		}
	}

	txHashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		return "", errors.WithStack(err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, txTimeout)
	defer cancel()

	err = retry.Do(timeoutCtx, txStatusPollInterval, func() error {
		requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()

		resultTx, err := c.clientCtx.Client.Tx(requestCtx, txHashBytes, false)
		if err != nil {
			if errors.Is(err, requestCtx.Err()) {
				return retry.Retryable(errors.WithStack(err))
			}
			if errRes := client.CheckTendermintError(err, encodedTx); errRes != nil {
				if isTxInMempool(errRes) {
					return retry.Retryable(errors.WithStack(err))
				}
				return errors.WithStack(err)
			}
			return retry.Retryable(errors.WithStack(err))
		}
		if resultTx.TxResult.Code != 0 {
			res := resultTx.TxResult
			if err := checkSequence(res.Codespace, res.Code, res.Log); err != nil {
				return err
			}
			return errors.Errorf("node returned non-zero code for tx '%s' (code: %d, codespace: %s): %s",
				txHash, res.Code, res.Codespace, res.Log)
		}
		if resultTx.Height == 0 {
			return retry.Retryable(errors.Errorf("transaction '%s' hasn't been included in a block yet", txHash))
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return txHash, nil
}

// PrepareTxBankSend creates a transaction sending tokens from one wallet to another
func (c Client) PrepareTxBankSend(ctx context.Context, sender, receiver Wallet, balance Balance) ([]byte, error) {
	fromAddress, err := sdk.AccAddressFromBech32(sender.Key.Address())
	must.OK(err)
	toAddress, err := sdk.AccAddressFromBech32(receiver.Key.Address())
	must.OK(err)

	signedTx, err := c.Sign(ctx, sender, banktypes.NewMsgSend(fromAddress, toAddress, sdk.Coins{
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

func isTxInMempool(errRes *sdk.TxResponse) bool {
	if errRes == nil {
		return false
	}
	return isSDKErrorResult(errRes.Codespace, errRes.Code, cosmoserrors.ErrTxInMempoolCache)
}

func isSDKErrorResult(codespace string, code uint32, sdkErr *cosmoserrors.Error) bool {
	return codespace == sdkErr.Codespace() &&
		code == sdkErr.ABCICode()
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

type sequenceError struct {
	expectedSequence uint64
	message          string
}

func (e sequenceError) Error() string {
	return e.message
}

func checkSequence(codespace string, code uint32, log string) error {
	// Cosmos SDK doesn't return expected sequence number as a parameter from RPC call,
	// so we must parse the error message in a hacky way.

	if !isSDKErrorResult(codespace, code, cosmoserrors.ErrWrongSequence) {
		return nil
	}
	matches := expectedSequenceRegExp.FindStringSubmatch(log)
	if len(matches) != 2 {
		return errors.Errorf("cosmos sdk hasn't returned expected sequence number, log mesage received: %s", log)
	}
	expectedSequence, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return errors.Wrapf(err, "can't parse expected sequence number, log mesage received: %s", log)
	}
	return errors.WithStack(sequenceError{message: log, expectedSequence: expectedSequence})
}

// FetchSequenceFromError checks if error is related to account sequence mismatch, and returns expected account sequence
func FetchSequenceFromError(err error) (uint64, bool) {
	var seqErr sequenceError
	if errors.As(err, &seqErr) {
		return seqErr.expectedSequence, true
	}
	return 0, false
}
