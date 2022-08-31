package tx

import (
	"strconv"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/pkg/errors"
	"github.com/tendermint/tendermint/mempool"
)

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
		return 0, false, errors.Errorf("cosmos sdk hasn't returned expected sequence number, log message received: %s", log)
	}
	expectedSequence, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, false, errors.Wrapf(err, "can't parse expected sequence number, log message received: %s", log)
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
