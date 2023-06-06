package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// ImportPendingTokenUpgrades imports pending version upgrades from genesis state.
func (k Keeper) ImportPendingTokenUpgrades(ctx sdk.Context, versions []types.GenesisTokenVersion) error {
	for _, v := range versions {
		if err := k.setPendingVersion(ctx, v.Denom, v.Version); err != nil {
			return err
		}
	}
	return nil
}

// ExportPendingTokenUpgrades exports pending version upgrades.
func (k Keeper) ExportPendingTokenUpgrades(ctx sdk.Context) ([]types.GenesisTokenVersion, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PendingVersionUpgradeKeyPrefix)
	versions := []types.GenesisTokenVersion{}
	_, err := query.Paginate(store, &query.PageRequest{Limit: query.MaxLimit}, func(key []byte, value []byte) error {
		version, _ := binary.Uvarint(value)
		versions = append(versions, types.GenesisTokenVersion{
			Denom:   string(key),
			Version: uint32(version),
		})

		return nil
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return versions, nil
}

// setPendingVersion sets pending vrsion for token upgrade.
func (k Keeper) setPendingVersion(ctx sdk.Context, denom string, version uint32) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PendingVersionUpgradeKeyPrefix)
	key := []byte(denom)
	if store.Has(key) {
		return errors.Errorf("upgrade is already pending for denom %q", denom)
	}

	value := make([]byte, binary.MaxVarintLen32)
	n := binary.PutUvarint(value, uint64(version))
	store.Set(key, value[:n])

	return nil
}

// clearPendingVersion clears pending version marker.
func (k Keeper) clearPendingVersion(ctx sdk.Context, denom string) {
	prefix.NewStore(ctx.KVStore(k.storeKey), types.PendingVersionUpgradeKeyPrefix).Delete([]byte(denom))
}
