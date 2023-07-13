package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// ErrInvalidState is returned when state of the module is invalid.
	ErrInvalidState = sdkerrors.Register(ModuleName, 1, "invalid state")
)
