package v1

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/v3/x/asset/nft/types"
)

// MigrateStore migrates asset nft module state from v1 to v2.
func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey) error {
	// old key format:
	// prefix (0x02) || classID-addrBytes
	// new key is of format
	// prefix (0x02) || addrBytes || classID
	moduleStore := ctx.KVStore(storeKey)
	oldStore := prefix.NewStore(moduleStore, types.NFTClassKeyPrefix)

	oldStoreIter := oldStore.Iterator(nil, nil)
	defer oldStoreIter.Close()

	for ; oldStoreIter.Valid(); oldStoreIter.Next() {
		oldKey := oldStoreIter.Key()
		newKey, err := types.CreateClassKey(string(oldKey))
		if err != nil {
			return errors.Errorf("can't re-create asset NFT class store key from %s, err:%s", string(oldKey), err)
		}

		moduleStore.Set(newKey, oldStoreIter.Value())
		oldStore.Delete(oldStoreIter.Key())
	}

	return nil
}
