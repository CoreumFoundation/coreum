package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
)

// Keeper is the dex module keeper.
type Keeper struct {
	cdc      codec.BinaryCodec
	storeKey storetypes.StoreKey
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
) Keeper {
	return Keeper{
		cdc:      cdc,
		storeKey: storeKey,
	}
}
