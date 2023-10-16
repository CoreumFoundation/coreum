package keeper

import (
	"github.com/CoreumFoundation/coreum/v3/x/nft"
	wnftkeeper "github.com/CoreumFoundation/coreum/v3/x/wnft/keeper"
)

// StoreKey is the store key string for nft.
const StoreKey = nft.ModuleName

// Keeper of the nft store.
type Keeper struct {
	wkeeper wnftkeeper.Wrapper
}

// NewKeeper creates a new nft Keeper instance.
func NewKeeper(
	wkeeper wnftkeeper.Wrapper,
) Keeper {
	return Keeper{
		wkeeper: wkeeper,
	}
}
