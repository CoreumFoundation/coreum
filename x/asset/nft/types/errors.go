package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// ErrInvalidInput defines the common error for the invalid input.
	ErrInvalidInput = sdkerrors.Register(ModuleName, 1, "invalid input")
	// ErrInvalidID is returned when the provided id is not of valid format
	ErrInvalidID = sdkerrors.Register(ModuleName, 2, "id format is not valid")
)
