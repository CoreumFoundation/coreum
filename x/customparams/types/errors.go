package types

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	// ErrInvalidState is returned when state of the module is invalid.
	ErrInvalidState = sdkerrors.Register(ModuleName, 1, "invalid state")
)
