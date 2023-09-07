package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v2/x/customparams/types"
)

// Keeper is customparams module Keeper.
type Keeper struct {
	storeKey  storetypes.StoreKey
	cdc       codec.BinaryCodec
	authority string
}

// NewKeeper returns a new Keeper instance.
func NewKeeper(
	storeKey storetypes.StoreKey,
	cdc codec.BinaryCodec,
	authority string,
) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		authority: authority,
	}
}

// GetStakingParams returns the set of staking parameters.
func (k Keeper) GetStakingParams(ctx sdk.Context) types.StakingParams {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.StakingParamsKey)
	var params types.StakingParams
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetStakingParams sets the module staking parameters to the param space.
func (k Keeper) SetStakingParams(ctx sdk.Context, params types.StakingParams) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(types.StakingParamsKey, bz)
	return nil
}

// GetAuthority return the module's authority.
func (k Keeper) GetAuthority() string {
	return k.authority
}
