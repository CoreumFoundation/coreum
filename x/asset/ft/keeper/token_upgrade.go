package keeper

import (
	"encoding/binary"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
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
	_, err := query.Paginate(store, &query.PageRequest{Limit: query.MaxLimit}, func(key, value []byte) error {
		version, n := binary.Uvarint(value)
		if n <= 0 {
			return sdkerrors.Wrap(types.ErrInvalidState, "unmarshaling varint failed")
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
	store := ctx.KVStore(k.storeKey)
	key := types.CreatePendingTokenUpgradeKey(denom)
	if store.Has(key) {
		return sdkerrors.Wrapf(cosmoserrors.ErrUnauthorized, "token upgrade is already pending for denom %q", denom)
	}

	value := make([]byte, binary.MaxVarintLen32)
	n := binary.PutUvarint(value, uint64(version))
	store.Set(key, value[:n])

	return nil
}

// ClearPendingVersion clears pending version marker.
func (k Keeper) ClearPendingVersion(ctx sdk.Context, denom string) {
	ctx.KVStore(k.storeKey).Delete(types.CreatePendingTokenUpgradeKey(denom))
}

// GetTokenUpgradeStatuses returns the token upgrade statuses of a specified denom.
func (k Keeper) GetTokenUpgradeStatuses(ctx sdk.Context, denom string) types.TokenUpgradeStatuses {
	bz := ctx.KVStore(k.storeKey).Get(types.CreateTokenUpgradeStatusesKey(denom))
	if bz == nil {
		return types.TokenUpgradeStatuses{}
	}
	var tokenUpgradeStatuses types.TokenUpgradeStatuses
	k.cdc.MustUnmarshal(bz, &tokenUpgradeStatuses)

	return tokenUpgradeStatuses
}

// SetTokenUpgradeStatuses sets the token upgrade statuses of a specified denom.
func (k Keeper) SetTokenUpgradeStatuses(ctx sdk.Context, denom string, tokenUpgradeStatuses types.TokenUpgradeStatuses) {
	ctx.KVStore(k.storeKey).Set(types.CreateTokenUpgradeStatusesKey(denom), k.cdc.MustMarshal(&tokenUpgradeStatuses))
}
