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
	// ErrInvalidDenom is returned when the provided denom is not valid
	ErrInvalidDenom = sdkerrors.Register(ModuleName, 4, "denom is not valid")
	// ErrInvalidKey is returned when the provided store key is invalid
	ErrInvalidKey = sdkerrors.Register(ModuleName, 5, "invalid key")
	// ErrNotEnoughBalance is returned when there is not enough
	ErrNotEnoughBalance = sdkerrors.Register(ModuleName, 6, "not enough balance")
	// ErrInvalidSymbol is returned when the provided symbol is not of valid format
	ErrInvalidSymbol = sdkerrors.Register(ModuleName, 7, "symbol format is not valid")
	// ErrGloballyFrozen is returned when token is globally frozen so all operations with it are blocked
	ErrGloballyFrozen = sdkerrors.Register(ModuleName, 8, "token is globally frozen")
)
