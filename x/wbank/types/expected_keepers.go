package types

import sdk "github.com/cosmos/cosmos-sdk/types"

// FungibleTokenProvider defines an interface to interact with the fungible token functionality.
type FungibleTokenProvider interface {
	IsSendAllowed(ctx sdk.Context, fromAddress, toAddress sdk.AccAddress, coins sdk.Coins) error
}
