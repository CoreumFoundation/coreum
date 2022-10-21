package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// ErrInvalidFungibleToken defines the common error for the invalid fungible tokens.
	ErrInvalidFungibleToken = sdkerrors.Register(ModuleName, 1, "invalid fungible token")
	// ErrFungibleTokenNotFound error for a fungible token not found in the store.
	ErrFungibleTokenNotFound = sdkerrors.Register(ModuleName, 2, "fungible token not found")
)
