package types

import (
	sdkerrors "cosmossdk.io/errors"
)

var (
	// ErrInvalidInput defines the common error for the invalid input.
	ErrInvalidInput = sdkerrors.Register(ModuleName, 1, "invalid input")

	// ErrInvalidCoin defines the error for the invalid coin.
	ErrInvalidCoin = sdkerrors.Register(ModuleName, 2, "invalid coin")

	// ErrInvalidPrice defines the error for the invalid price.
	ErrInvalidPrice = sdkerrors.Register(ModuleName, 3, "invalid price")
)
