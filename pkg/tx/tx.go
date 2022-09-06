package tx

import (
	"context"
	"encoding/hex"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/mempool"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
)

var (
	txTimeout            = time.Minute
	txStatusPollInterval = 500 * time.Millisecond
	requestTimeout       = 10 * time.Second
)

// Factory is a re-export of the cosmos sdk tx.Factory type, to make usage of this package more convenient.
// It will help users by removing the need to import tx package from cosmos sdk and help avoid package name collision.
type Factory = tx.Factory

// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will also simulate gas requirements if necessary.
// It will return an error upon failure.
// NOTE: copied from the link below and made some changes
// https://github.com/cosmos/cosmos-sdk/blob/v0.45.2/client/tx/tx.go
// TODO: add test to check if client respects ctx.
func BroadcastTx(ctx context.Context, clientCtx client.Context, txf Factory, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	txf, err := prepareFactory(ctx, clientCtx, txf)
	if err != nil {
		return nil, err
	}

	unsignedTx, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	unsignedTx.SetFeeGranter(clientCtx.GetFeeGranterAddress())
	err = tx.Sign(txf, clientCtx.GetFromName(), unsignedTx, true)
	if err != nil {
		return nil, err
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(unsignedTx.GetTx())
	if err != nil {
		return nil, err
	}

	// broadcast to a Tendermint node
	switch clientCtx.BroadcastMode {
	case flags.BroadcastSync:
		res, err := clientCtx.Client.BroadcastTxSync(ctx, txBytes)
		if err != nil {
			return nil, err
		}
		return sdk.NewResponseFormatBroadcastTx(res), nil

	case flags.BroadcastAsync:
		res, err := clientCtx.Client.BroadcastTxAsync(ctx, txBytes)
		if err != nil {
			return nil, err
		}
		return sdk.NewResponseFormatBroadcastTx(res), nil

	case flags.BroadcastBlock:
		res, err := clientCtx.Client.BroadcastTxSync(ctx, txBytes)
		if err != nil {
			return nil, err
		}

		awaitRes, err := AwaitTx(ctx, clientCtx, res.Hash.String())
		if err != nil {
			return nil, err
		}

		return sdk.NewResponseResultTx(awaitRes, nil, ""), nil

	default:
		return nil, errors.Errorf("unsupported broadcast mode %s; supported types: sync, async, block", clientCtx.BroadcastMode)
	}
}

func prepareFactory(ctx context.Context, clientCtx client.Context, txf tx.Factory) (tx.Factory, error) {
	if txf.AccountNumber() == 0 || txf.Sequence() == 0 {
		acc, err := GetAccountInfo(ctx, clientCtx, clientCtx.GetFromAddress())
		if err != nil {
			return txf, err
		}
		txf = txf.
			WithAccountNumber(acc.GetAccountNumber()).
			WithSequence(acc.GetSequence())
	}

	return txf, nil
}

// GetAccountInfo returns account number and account sequence for provided address
func GetAccountInfo(
	ctx context.Context,
	clientCtx client.Context,
	address sdk.AccAddress,
) (authtypes.AccountI, error) {
	req := &authtypes.QueryAccountRequest{
		Address: address.String(),
	}
	authQueryClient := authtypes.NewQueryClient(clientCtx)
	res, err := authQueryClient.Account(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var acc authtypes.AccountI
	if err := clientCtx.InterfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
		return nil, errors.WithStack(err)
	}

	return acc, nil
}

// AwaitTx awaits until a signed transaction is included in a block, returning the result.
func AwaitTx(
	ctx context.Context,
	clientCtx client.Context,
	txHash string,
) (resultTx *coretypes.ResultTx, err error) {
	txHashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		err = errors.Wrap(err, "tx hash is not a valid hex")
		return nil, err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, txTimeout)
	defer cancel()

	if err = retry.Do(timeoutCtx, txStatusPollInterval, func() error {
		requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()

		var err error
		resultTx, err = clientCtx.Client.Tx(requestCtx, txHashBytes, false)
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

	return resultTx, nil
}

// checkTendermintError checks if the error returned from BroadcastTx is a
// Tendermint error that is returned before the tx is submitted due to
// precondition checks that failed. If an Tendermint error is detected, this
// function returns the correct code back in TxResponse.
//
// NOTE: copy paste from Cosmos SDK! To avoid hassle getting the tx hash.
// copied from https://github.com/cosmos/cosmos-sdk/blob/069514e4d607718995a4c8c4dc02785cbbccf752/client/broadcast.go#L49
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
