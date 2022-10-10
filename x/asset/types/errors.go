package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// ErrInvalidAsset defines the common error for the invalid asset state.
	ErrInvalidAsset = sdkerrors.Register(ModuleName, 1, "Invalid asset")
)
