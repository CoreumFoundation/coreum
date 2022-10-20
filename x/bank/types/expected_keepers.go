package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// FungibleTokenProvider defines an interface to interact with the fungible token functionality.
type FungibleTokenProvider interface {
	GetLockedCoins(ctx sdk.Context, address sdk.AccAddress) sdk.Coins
}
