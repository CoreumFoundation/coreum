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
func (k Keeper) ImportPendingTokenUpgrades(ctx sdk.Context, versions []types.PendingTokenUpgrade) error {
	for _, v := range versions {
		if err := k.SetPendingVersion(ctx, v.Denom, v.Version); err != nil {
			return err
		}
	}
	return nil
}

// ExportPendingTokenUpgrades exports pending version upgrades.
func (k Keeper) ExportPendingTokenUpgrades(ctx sdk.Context) ([]types.PendingTokenUpgrade, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PendingTokenUpgradeKeyPrefix)
	versions := []types.PendingTokenUpgrade{}
	_, err := query.Paginate(store, &query.PageRequest{Limit: query.MaxLimit}, func(key []byte, value []byte) error {
		version, n := binary.Uvarint(value)
		if n <= 0 {
			return errors.New("unmarshaling varint failed")
		}
		versions = append(versions, types.PendingTokenUpgrade{
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

// SetPendingVersion sets pending vrsion for token upgrade.
func (k Keeper) SetPendingVersion(ctx sdk.Context, denom string, version uint32) error {
	store := ctx.KVStore(k.storeKey)
	key := types.CreatePendingTokenUpgradeKey(denom)
	if store.Has(key) {
		return errors.Errorf("token upgrade is already pending for denom %q", denom)
	}

	value := make([]byte, binary.MaxVarintLen32)
	n := binary.PutUvarint(value, uint64(version))
	store.Set(key, value[:n])

	return nil
}

// ClearPendingVersion clears pending version marker.
func (k Keeper) ClearPendingVersion(ctx sdk.Context, denom string) {
	prefix.NewStore(ctx.KVStore(k.storeKey), types.PendingTokenUpgradeKeyPrefix).Delete([]byte(denom))
}
