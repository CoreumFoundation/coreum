package keeper

import (
	"encoding/binary"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
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
	moduleStore := k.storeService.OpenKVStore(ctx)
	store := prefix.NewStore(runtime.KVStoreAdapter(moduleStore), types.PendingTokenUpgradeKeyPrefix)
	versions := []types.PendingTokenUpgrade{}
	_, err := query.Paginate(store, &query.PageRequest{Limit: query.PaginationMaxLimit}, func(key, value []byte) error {
		version, n := binary.Uvarint(value)
		if n <= 0 {
			return sdkerrors.Wrap(types.ErrInvalidState, "unmarshalling varint failed")
		}
		versions = append(versions, types.PendingTokenUpgrade{
			Denom:   string(key),
			Version: uint32(version),
		})

		return nil
	})
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidInput, "failed to paginate: %s", err)
	}

	return versions, nil
}

// SetPendingVersion sets pending version for token upgrade.
func (k Keeper) SetPendingVersion(ctx sdk.Context, denom string, version uint32) error {
	store := k.storeService.OpenKVStore(ctx)
	key := types.CreatePendingTokenUpgradeKey(denom)
	val, err := store.Has(key)
	if err != nil {
		return err
	}
	if val {
		return sdkerrors.Wrapf(cosmoserrors.ErrUnauthorized, "token upgrade is already pending for denom %q", denom)
	}

	value := make([]byte, binary.MaxVarintLen32)
	n := binary.PutUvarint(value, uint64(version))
	return store.Set(key, value[:n])
}

// ClearPendingVersion clears pending version marker.
func (k Keeper) ClearPendingVersion(ctx sdk.Context, denom string) error {
	return k.storeService.OpenKVStore(ctx).Delete(types.CreatePendingTokenUpgradeKey(denom))
}

// GetTokenUpgradeStatuses returns the token upgrade statuses of a specified denom.
func (k Keeper) GetTokenUpgradeStatuses(ctx sdk.Context, denom string) (types.TokenUpgradeStatuses, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.CreateTokenUpgradeStatusesKey(denom))
	if err != nil {
		return types.TokenUpgradeStatuses{}, err
	}
	if bz == nil {
		return types.TokenUpgradeStatuses{}, nil
	}
	var tokenUpgradeStatuses types.TokenUpgradeStatuses
	k.cdc.MustUnmarshal(bz, &tokenUpgradeStatuses)

	return tokenUpgradeStatuses, nil
}

// SetTokenUpgradeStatuses sets the token upgrade statuses of a specified denom.
func (k Keeper) SetTokenUpgradeStatuses(
	ctx sdk.Context,
	denom string,
	tokenUpgradeStatuses types.TokenUpgradeStatuses,
) error {
	return k.storeService.OpenKVStore(ctx).Set(
		types.CreateTokenUpgradeStatusesKey(denom),
		k.cdc.MustMarshal(&tokenUpgradeStatuses),
	)
}
