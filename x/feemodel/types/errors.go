package types

import (
	sdkerrors "cosmossdk.io/errors"
)

// ErrInvalidState is returned when state of the module is invalid.
var ErrInvalidState = sdkerrors.Register(ModuleName, 1, "invalid state")
