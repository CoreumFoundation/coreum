package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

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

// UpdateStakingParams is a governance operation that sets the staking parameters of the module.
func (k Keeper) UpdateStakingParams(ctx sdk.Context, authority string, params types.StakingParams) error {
	if k.authority != authority {
		return sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, authority)
	}

	return k.SetStakingParams(ctx, params)
}
