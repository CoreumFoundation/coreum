package keeper

import (
	sdkstore "cosmossdk.io/core/store"
	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/CoreumFoundation/coreum/v6/x/customparams/types"
)

// Keeper is customparams module Keeper.
type Keeper struct {
	storeService sdkstore.KVStoreService
	cdc          codec.BinaryCodec
	authority    string
}

// NewKeeper returns a new Keeper instance.
func NewKeeper(
	storeService sdkstore.KVStoreService,
	cdc codec.BinaryCodec,
	authority string,
) Keeper {
	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		authority:    authority,
	}
}

// GetStakingParams returns the set of staking parameters.
func (k Keeper) GetStakingParams(ctx sdk.Context) (types.StakingParams, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.StakingParamsKey)
	if err != nil {
		return types.StakingParams{}, err
	}
	var params types.StakingParams
	k.cdc.MustUnmarshal(bz, &params)
	return params, nil
}

// SetStakingParams sets the module staking parameters to the param space.
func (k Keeper) SetStakingParams(ctx sdk.Context, params types.StakingParams) error {
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.StakingParamsKey, bz)
}

// UpdateStakingParams is a governance operation that sets the staking parameters of the module.
func (k Keeper) UpdateStakingParams(ctx sdk.Context, authority string, params types.StakingParams) error {
	if k.authority != authority {
		return sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, authority)
	}

	return k.SetStakingParams(ctx, params)
}
