package types

import sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

// ErrInvalidData defines the error for invalid delayed data items.
var ErrInvalidData = sdkerrors.Register(ModuleName, 1, "invalid data")
