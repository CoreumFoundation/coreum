package store

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const GlobalCodespace = "global"

var (
	// ErrInvalidKey is returned when the provided store key is invalid.
	ErrInvalidKey = sdkerrors.Register(GlobalCodespace, 1, "invalid key")
)
