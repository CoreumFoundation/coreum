package testing

import (
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/pkg/errors"
)

// IsErr returns true if error was caused by insufficient funds provided with the transaction.
func IsErr(err error, cosmosErr *cosmoserrors.Error) bool {
	return asSDKError(err, cosmosErr) != nil
}

func asSDKError(err error, expectedSDKErr *cosmoserrors.Error) *cosmoserrors.Error {
	var sdkErr *cosmoserrors.Error
	if !errors.As(err, &sdkErr) || !isSDKErrorResult(sdkErr.Codespace(), sdkErr.ABCICode(), expectedSDKErr) {
		return nil
	}
	return sdkErr
}

func isSDKErrorResult(codespace string, code uint32, expectedSDKError *cosmoserrors.Error) bool {
	return codespace == expectedSDKError.Codespace() &&
		code == expectedSDKError.ABCICode()
}
