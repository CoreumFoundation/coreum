package types

import (
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	// ErrInvalidInput defines the common error for the invalid input.
	ErrInvalidInput = sdkerrors.Register(ModuleName, 1, "invalid input")
	// ErrInvalidID is returned when the provided id is not of valid format
	ErrInvalidID = sdkerrors.Register(ModuleName, 2, "id format is not valid")
	// ErrClassNotFound is returned when token class not found in the store.
	ErrClassNotFound = sdkerrors.Register(ModuleName, 3, "non-fungible token class not found")
	// ErrFeatureDisabled is returned when used disabled feature.
	ErrFeatureDisabled = sdkerrors.Register(ModuleName, 4, "feature disabled")
	// ErrNFTNotFound is returned if a non-fungible token not found in the store.
	ErrNFTNotFound = sdkerrors.Register(ModuleName, 5, "non-fungible token not found")
	// ErrInvalidKey is returned when the provided store key is invalid
	ErrInvalidKey = sdkerrors.Register(ModuleName, 6, "invalid key")
)
