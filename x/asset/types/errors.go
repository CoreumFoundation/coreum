package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// ErrInvalidAsset defines the common error for the invalid asset state.
	ErrInvalidAsset = sdkerrors.Register(ModuleName, 1, "invalid asset")
	// ErrNotFound error for an entry not found in the store
	ErrNotFound = sdkerrors.Register(ModuleName, 2, "not found")
	// ErrInvalidState defines the common error for the invalid state.
	ErrInvalidState = sdkerrors.Register(ModuleName, 3, "invalid state")
)
