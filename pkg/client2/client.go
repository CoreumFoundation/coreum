package client2

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/mempool"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/CoreumFoundation/coreum/app"
)

const (
	requestTimeout       = 10 * time.Second
	txTimeout            = time.Minute
	txStatusPollInterval = 500 * time.Millisecond
)

var expectedSequenceRegExp = regexp.MustCompile(`account sequence mismatch, expected (\d+), got \d+`)

// Client is the client for cored blockchain
type Client interface {
	ClientContext() client.Context
	GetNumberSequence(ctx context.Context, address sdk.AccAddress) (uint64, uint64, error)
	EstimateGas(ctx context.Context, config TxConfig, msgs ...sdk.Msg) (int64, error)
	BroadcastAsync(ctx context.Context, config TxConfig, msgs ...sdk.Msg) (string, error)
	BroadcastSync(ctx context.Context, config TxConfig, msgs ...sdk.Msg) (*TxResult, error)
	TxResult(ctx context.Context, txHash string) (*TxResult, error)
}

// TxResult contains results of transaction broadcast
type TxResult struct {
	TxHash  string
	GasUsed int64
}

// New creates new client for cored
func New(chainID app.ChainID, addr string) (Client, error) {
	parsedURL, err := url.Parse(addr)
	if err != nil {
		return coreClient{}, errors.WithStack(err)
	}
	switch parsedURL.Scheme {
	case "tcp", "http", "https":
	default:
		return coreClient{}, errors.Errorf("unknown scheme '%s' in address", parsedURL.Scheme)
	}
	rpcClient, err := client.NewClientFromNode(addr)
	if err != nil {
		return coreClient{}, errors.WithStack(err)
	}

	clientCtx := app.
		NewDefaultClientContext().
		WithChainID(string(chainID)).
		WithClient(rpcClient)

	return coreClient{
		clientCtx:       clientCtx,
		authQueryClient: authtypes.NewQueryClient(clientCtx),
		txServiceClient: txtypes.NewServiceClient(clientCtx),
	}, nil
}

type coreClient struct {
	clientCtx       client.Context
	authQueryClient authtypes.QueryClient
	txServiceClient txtypes.ServiceClient
}

func (c coreClient) ClientContext() client.Context {
	return c.clientCtx
}

// GetNumberSequence returns account number and account sequence for provided address
func (c coreClient) GetNumberSequence(
	ctx context.Context,
	address sdk.AccAddress,
) (num uint64, seq uint64, err error) {
	addr, err := sdk.AccAddressFromBech32(string(address))
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}

	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req := &authtypes.QueryAccountRequest{
		Address: addr.String(),
	}
	res, err := c.authQueryClient.Account(requestCtx, req)
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}

	var acc authtypes.AccountI
	if err := c.clientCtx.InterfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
		return 0, 0, errors.WithStack(err)
	}

	return acc.GetAccountNumber(), acc.GetSequence(), nil
}

// EstimateGas runs the transaction cost estimation and returns new suggested gas limit,
// in contrast with the default Cosmos SDK gas estimation logic, this method returns unadjusted gas used.
func (c coreClient) EstimateGas(ctx context.Context, config TxConfig, msgs ...sdk.Msg) (int64, error) {
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	encodedTx, err := buildSimTx(c.clientCtx, config, msgs...)
	if err != nil {
		err = errors.Wrap(err, "failed to build Tx for simulation")
		return 0, err
	}

	simRes, err := c.txServiceClient.Simulate(requestCtx, &txtypes.SimulateRequest{
		TxBytes: encodedTx,
	})
	if err != nil {
		return 0, errors.Wrap(err, "failed to simulate the transaction execution")
	}

	// usually gas has to be multiplied by some adjustment coefficient: e.g. *1.5
	// but in this case we return unadjusted, so every module can decide the adjustment value
	return int64(simRes.GasInfo.GasUsed), nil
}

// BroadcastAsync sends transaction to chain, ensuring it passeses CheckTx.
// Doesn't await for Tx being included in a block.
func (c coreClient) BroadcastAsync(ctx context.Context, config TxConfig, msgs ...sdk.Msg) (txHash string, err error) {
	encodedTx, err := c.prepareTx(ctx, config, msgs...)
	if err != nil {
		return "", err
	}

	txHash = fmt.Sprintf("%X", tmtypes.Tx(encodedTx).Hash())

	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	res, err := c.clientCtx.Client.BroadcastTxSync(requestCtx, encodedTx)
	if err != nil {
		if errors.Is(err, requestCtx.Err()) {
			return txHash, errors.WithStack(err)
		}

		errRes := checkTendermintError(err, txHash)
		if !isTxInMempool(errRes) {
			return txHash, errors.WithStack(err)
		}
	} else if res.Code != 0 {
		err := errors.Wrapf(sdkerrors.New(res.Codespace, res.Code, res.Log),
			"transaction '%s' failed", txHash)
		return txHash, err
	}

	return res.Hash.String(), nil
}

// TxResult awaits until a signed transaction is included in a block, returing the TxResult.
func (c coreClient) TxResult(ctx context.Context, txHash string) (*TxResult, error) {
	txHashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		err = errors.Wrap(err, "tx hash is not a valid hex")
		return nil, err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, txTimeout)
	defer cancel()

	var resultTx *coretypes.ResultTx
	if err = retry.Do(timeoutCtx, txStatusPollInterval, func() error {
		requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()

		var err error
		resultTx, err = c.clientCtx.Client.Tx(requestCtx, txHashBytes, false)
		if err != nil {
			if errors.Is(err, requestCtx.Err()) {
				return retry.Retryable(errors.WithStack(err))
			}

			if errRes := checkTendermintError(err, txHash); errRes != nil {
				if isTxInMempool(errRes) {
					return retry.Retryable(errors.WithStack(err))
				}
				return errors.WithStack(err)
			}

			return retry.Retryable(errors.WithStack(err))
		}

		if resultTx.TxResult.Code != 0 {
			res := resultTx.TxResult
			return errors.Wrapf(sdkerrors.New(res.Codespace, res.Code, res.Log), "transaction '%s' failed", txHash)
		}

		if resultTx.Height == 0 {
			return retry.Retryable(errors.Errorf("transaction '%s' hasn't been included in a block yet", txHash))
		}

		return nil
	}); err != nil {
		return nil, err
	}

	txResult := &TxResult{
		TxHash:  txHash,
		GasUsed: resultTx.TxResult.GasUsed,
	}

	return txResult, nil
}

// BroadcastSync is a shortcut for broadcasting the Tx and awaiting for inclusion in a block.
func (c coreClient) BroadcastSync(
	ctx context.Context,
	config TxConfig,
	msgs ...sdk.Msg,
) (txResult *TxResult, err error) {
	txHash, err := c.BroadcastAsync(ctx, config, msgs...)
	if err != nil {
		return nil, err
	}

	return c.TxResult(ctx, txHash)
}

// prepareTx encodes messages in a new transaction then signs and encodes it
func (c coreClient) prepareTx(ctx context.Context, config TxConfig, msgs ...sdk.Msg) ([]byte, error) {
	if config.Keyring == nil {
		err := errors.New("prepareTx is required but no keyring provided")
		return nil, err
	}

	if config.FromAccount == nil {
		num, seq, err := c.GetNumberSequence(ctx, config.From)
		if err != nil {
			return nil, err
		}

		config.SetAccountNumber(num)
		config.SetAccountSequence(seq)
	}

	signedTx, err := signTx(c.clientCtx, config, msgs...)
	if err != nil {
		return nil, err
	}

	signedBytes, err := c.clientCtx.TxConfig.TxEncoder()(signedTx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return signedBytes, err
}

func isTxInMempool(errRes *sdk.TxResponse) bool {
	if errRes == nil {
		return false
	}
	return isSDKErrorResult(errRes.Codespace, errRes.Code, sdkerrors.ErrTxInMempoolCache)
}

func isSDKErrorResult(codespace string, code uint32, expectedSDKError *sdkerrors.Error) bool {
	return codespace == expectedSDKError.Codespace() &&
		code == expectedSDKError.ABCICode()
}

func asSDKError(err error, expectedSDKErr *sdkerrors.Error) *sdkerrors.Error {
	var sdkErr *sdkerrors.Error
	if !errors.As(err, &sdkErr) || !isSDKErrorResult(sdkErr.Codespace(), sdkErr.ABCICode(), expectedSDKErr) {
		return nil
	}
	return sdkErr
}

// ExpectedSequenceFromError checks if error is related to account sequence mismatch, and returns expected account sequence
func ExpectedSequenceFromError(err error) (uint64, bool, error) {
	sdkErr := asSDKError(err, sdkerrors.ErrWrongSequence)
	if sdkErr == nil {
		return 0, false, nil
	}

	log := sdkErr.Error()
	matches := expectedSequenceRegExp.FindStringSubmatch(log)
	if len(matches) != 2 {
		return 0, false, errors.Errorf("cosmos sdk hasn't returned expected sequence number, log mesage received: %s", log)
	}
	expectedSequence, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, false, errors.Wrapf(err, "can't parse expected sequence number, log mesage received: %s", log)
	}
	return expectedSequence, true, nil
}

// IsInsufficientFeeError returns true if error was caused by insufficient fee provided with the transaction
func IsInsufficientFeeError(err error) bool {
	return asSDKError(err, sdkerrors.ErrInsufficientFee) != nil
}

// checkTendermintError checks if the error returned from BroadcastTx is a
// Tendermint error that is returned before the tx is submitted due to
// precondition checks that failed. If an Tendermint error is detected, this
// function returns the correct code back in TxResponse.
//
// NOTE: Copypasta from Cosmos SDK! To avoid hassle getting the tx hash.
func checkTendermintError(err error, txHash string) *sdk.TxResponse {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, strings.ToLower(mempool.ErrTxInCache.Error())):
		return &sdk.TxResponse{
			Code:      sdkerrors.ErrTxInMempoolCache.ABCICode(),
			Codespace: sdkerrors.ErrTxInMempoolCache.Codespace(),
			TxHash:    txHash,
		}

	case strings.Contains(errStr, "mempool is full"):
		return &sdk.TxResponse{
			Code:      sdkerrors.ErrMempoolIsFull.ABCICode(),
			Codespace: sdkerrors.ErrMempoolIsFull.Codespace(),
			TxHash:    txHash,
		}

	case strings.Contains(errStr, "tx too large"):
		return &sdk.TxResponse{
			Code:      sdkerrors.ErrTxTooLarge.ABCICode(),
			Codespace: sdkerrors.ErrTxTooLarge.Codespace(),
			TxHash:    txHash,
		}

	default:
		return nil
	}
}
