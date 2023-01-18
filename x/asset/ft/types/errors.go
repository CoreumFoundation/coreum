package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// ErrInvalidInput defines the common error for the invalid input.
	ErrInvalidInput = sdkerrors.Register(ModuleName, 1, "invalid input")
	// ErrTokenNotFound error for a fungible token not found in the store.
	ErrTokenNotFound = sdkerrors.Register(ModuleName, 2, "token not found")
	// ErrInvalidKey is returned when the provided store key is invalid
	ErrInvalidKey = sdkerrors.Register(ModuleName, 3, "invalid key")
	// ErrFeatureDisabled is returned when used disabled feature.
	ErrFeatureDisabled = sdkerrors.Register(ModuleName, 4, "feature disabled")
	// ErrInvalidDenom is returned when the provided denom is invalid fungible token denom
	ErrInvalidDenom = sdkerrors.Register(ModuleName, 5, "invalid denom")
	// ErrGloballyFrozen is returned when token is globally frozen so all operations with it are blocked
	ErrGloballyFrozen = sdkerrors.Register(ModuleName, 6, "token is globally frozen")
	// ErrWhitelistedLimitExceeded is returned when new balance after receiving coins exceeds the whitelisted limit
	ErrWhitelistedLimitExceeded = sdkerrors.Register(ModuleName, 7, "whitelisted limit exceeded")
)
