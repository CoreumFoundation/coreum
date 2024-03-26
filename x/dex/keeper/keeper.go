package keeper

import (
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

// Order defines order methods required by the keeper.
type Order interface {
	proto.Message
	DenomOffered() string
	DenomRequested() string
}

// Keeper is the dex module keeper.
type Keeper struct {
	cdc               codec.BinaryCodec
	storeKey          storetypes.StoreKey
	transientStoreKey storetypes.StoreKey
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	transientStoreKey storetypes.StoreKey,
) Keeper {
	return Keeper{
		cdc:               cdc,
		storeKey:          storeKey,
		transientStoreKey: transientStoreKey,
	}
}

// StoreTransientOrder stores order in the transient queue to be processed by the end blocker.
func (k Keeper) StoreTransientOrder(ctx sdk.Context, order Order) error {
	denom1Seq, denom2Seq := k.denomTransientSequences(ctx, order.DenomOffered(), order.DenomRequested())
	orderSeq := k.nextOrderTransientSequence(ctx)

	orderAny, err := codectypes.NewAnyWithValue(order)
	if err != nil {
		return err
	}

	bz, err := orderAny.Marshal()
	if err != nil {
		return err
	}

	tStore := ctx.TransientStore(k.transientStoreKey)
	tStore.Set(types.CreateOrderTransientQueueKey(denom1Seq, denom2Seq, orderSeq), bz)

	return nil
}

func (k Keeper) nextOrderTransientSequence(ctx sdk.Context) uint64 {
	tStore := ctx.TransientStore(k.transientStoreKey)

	seq := sdkmath.ZeroUint()
	bz := tStore.Get(types.OrderTransientSequenceKey)

	if bz != nil {
		if err := seq.Unmarshal(bz); err != nil {
			panic(err)
		}
	}
	seq.Incr()
	bz, err := seq.Marshal()
	if err != nil {
		panic(err)
	}

	tStore.Set(types.OrderTransientSequenceKey, bz)

	return seq.Uint64()
}

func (k Keeper) nextDenomSequence(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)

	seq := sdkmath.ZeroUint()
	bz := store.Get(types.DenomSequenceKey)

	if bz != nil {
		if err := seq.Unmarshal(bz); err != nil {
			panic(err)
		}
	}
	seq.Incr()
	bz, err := seq.Marshal()
	if err != nil {
		panic(err)
	}

	store.Set(types.DenomSequenceKey, bz)

	return seq.Uint64()
}

func (k Keeper) denomTransientSequences(ctx sdk.Context, denom1, denom2 string) (uint64, uint64) {
	if strings.Compare(denom1, denom2) > 0 {
		denom1, denom2 = denom2, denom1
	}

	return k.denomSequence(ctx, denom1), k.denomSequence(ctx, denom2)
}

func (k Keeper) denomSequence(ctx sdk.Context, denom string) uint64 {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateDenomMappingKey(denom)
	bz := store.Get(key)
	if bz == nil {
		seqRaw := k.nextDenomSequence(ctx)
		seq := sdkmath.NewUint(seqRaw)

		bz, err := seq.Marshal()
		if err != nil {
			panic(err)
		}
		store.Set(key, bz)

		return seqRaw
	}

	seq := sdkmath.ZeroUint()
	if err := seq.Unmarshal(bz); err != nil {
		panic(err)
	}
	return seq.Uint64()
}
