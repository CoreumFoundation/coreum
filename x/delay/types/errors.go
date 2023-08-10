package types

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	// ErrInvalidData defines the error for invalid delayed data items.
	ErrInvalidData = sdkerrors.Register(ModuleName, 1, "invalid data")
	// ErrInvalidInput error is returned if input data are invalid.
	ErrInvalidInput = sdkerrors.Register(ModuleName, 2, "invalid input")
	// ErrInvalidConfiguration is returned when something is wrong with the configuration.
	ErrInvalidConfiguration = sdkerrors.Register(ModuleName, 3, "invalid configuration")
)
