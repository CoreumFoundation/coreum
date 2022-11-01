package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// ErrInvalidFungibleToken defines the common error for the invalid fungible tokens.
	ErrInvalidFungibleToken = sdkerrors.Register(ModuleName, 1, "invalid fungible token")
	// ErrFungibleTokenNotFound error for a fungible token not found in the store.
	ErrFungibleTokenNotFound = sdkerrors.Register(ModuleName, 2, "fungible token not found")
	// ErrFeatureNotActive is returned when an operation is performed on a token which is missing a required feature
	ErrFeatureNotActive = sdkerrors.Register(ModuleName, 3, "token feature is not active")
	// ErrInvalidDenomChecksum is returned when the checksum on the denom is not correct
	ErrInvalidDenomChecksum = sdkerrors.Register(ModuleName, 4, "denom checksum is not valid")
	// ErrInvalidDenomFormat is returned when the provided denom does not match required format
	ErrInvalidDenomFormat = sdkerrors.Register(ModuleName, 5, "denom format is not valid")
)
