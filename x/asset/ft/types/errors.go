package types

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	// ErrInvalidInput defines the common error for the invalid input.
	ErrInvalidInput = sdkerrors.Register(ModuleName, 1, "invalid input")
	// ErrTokenNotFound error for a fungible token not found in the store.
	ErrTokenNotFound = sdkerrors.Register(ModuleName, 2, "token not found")
	// ErrInvalidKey is returned when the provided store key is invalid.
	ErrInvalidKey = sdkerrors.Register(ModuleName, 3, "invalid key")
	// ErrFeatureDisabled is returned when used disabled feature.
	ErrFeatureDisabled = sdkerrors.Register(ModuleName, 4, "feature disabled")
	// ErrInvalidDenom is returned when the provided denom is invalid fungible token denom.
	ErrInvalidDenom = sdkerrors.Register(ModuleName, 5, "invalid denom")
	// ErrGloballyFrozen is returned when token is globally frozen so all operations with it are blocked.
	ErrGloballyFrozen = sdkerrors.Register(ModuleName, 6, "token is globally frozen")
	// ErrWhitelistedLimitExceeded is returned when new balance after receiving coins exceeds the whitelisted limit.
	ErrWhitelistedLimitExceeded = sdkerrors.Register(ModuleName, 7, "whitelisted limit exceeded")
	// ErrInvalidState is returned when state of the module is invalid.
	ErrInvalidState = sdkerrors.Register(ModuleName, 8, "invalid state")
	// ErrExtensionCallFailed is returned when the execution of the asset extensino fails.
	ErrExtensionCallFailed = sdkerrors.Register(ModuleName, 9, "call to asset extension failed")
	// ErrDEXSettingsNotFound error for a DEX settings not found in the store.
	ErrDEXSettingsNotFound = sdkerrors.Register(ModuleName, 10, "DEX settings not found")
	// ErrDEXLockFailed is returned when DEX lock is failed.
	ErrDEXLockFailed = sdkerrors.Register(ModuleName, 11, "DEX lock is failed")
)
