package store

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const globalCodespace = "global"

var (
	// ErrInvalidKey is returned when the provided store key is invalid.
	ErrInvalidKey = sdkerrors.Register(globalCodespace, 1, "invalid key")
)
