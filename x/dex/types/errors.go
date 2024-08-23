package types

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	// ErrInvalidInput defines the common error for the invalid input.
	ErrInvalidInput = sdkerrors.Register(ModuleName, 1, "invalid input")
	// ErrInvalidKey is returned when the provided store key is invalid.
	ErrInvalidKey = sdkerrors.Register(ModuleName, 2, "invalid key")
	// ErrInvalidState is returned when state of the module is invalid.
	ErrInvalidState = sdkerrors.Register(ModuleName, 3, "invalid state")
	// ErrRecordNotFound is returned when record is not found in the store.
	ErrRecordNotFound = sdkerrors.Register(ModuleName, 4, "record not found")
	// ErrFailedToLockCoin is returned when the coin for the order locking is failed.
	ErrFailedToLockCoin = sdkerrors.Register(ModuleName, 5, "failed to lock coin")
)
