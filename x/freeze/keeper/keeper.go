package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/CoreumFoundation/coreum/x/freeze/types"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   sdk.StoreKey
		memKey     sdk.StoreKey
		paramstore paramtypes.Subspace
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey sdk.StoreKey,
	ps paramtypes.Subspace,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k *Keeper) IsFrozenCoin(ctx sdk.Context, holder sdk.AccAddress, denom string) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), holder.Bytes())
	return store.Has([]byte(denom))
}

func (k *Keeper) FreezeCoin(ctx sdk.Context, holder sdk.AccAddress, denom string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), holder.Bytes())

	if store.Has([]byte(denom)) {
		return fmt.Errorf("coin %s is already frozen on the given account", denom)
	}

	store.Set([]byte(denom), []byte("true"))

	return nil
}

func (k *Keeper) UnfreezeCoin(ctx sdk.Context, holder sdk.AccAddress, denom string) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), holder.Bytes())

	if !store.Has([]byte(denom)) {
		return fmt.Errorf("coin %s is not frozen on the given account", denom)
	}

	store.Delete([]byte(denom))

	return nil
}
