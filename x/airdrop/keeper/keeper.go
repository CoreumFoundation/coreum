package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum/x/airdrop/types"
	coreumbanktypes "github.com/CoreumFoundation/coreum/x/bank/types"
	snapshottypes "github.com/CoreumFoundation/coreum/x/snapshot/types"
)

type SnapshotKeeper interface {
	RequestSnapshot(ctx sdk.Context, request snapshottypes.SnapshotRequestInfo) (uint64, error)
}

// Keeper is the asset module keeper.
type Keeper struct {
	cdc            codec.BinaryCodec
	storeKey       sdk.StoreKey
	snapshotKeeper SnapshotKeeper
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	snapshotKeeper SnapshotKeeper,
) Keeper {
	return Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		snapshotKeeper: snapshotKeeper,
	}
}

func (k Keeper) Create(ctx sdk.Context, airdropInfo types.AirdropInfo) error {
	// FIXME (wojtek): verify that denom exists
	// FIXME (wojtek): take control over funds required for airdrop

	store := ctx.KVStore(k.storeKey)

	id, err := k.snapshotKeeper.RequestSnapshot(ctx, snapshottypes.SnapshotRequestInfo{
		Prefix: snapshottypes.SnapshotPrefix{
			StoreName: banktypes.StoreKey,
			Name:      coreumbanktypes.SnapshotName(airdropInfo.RequiredDenom),
		},
		Owner:           airdropInfo.Sender,
		Height:          airdropInfo.Height,
		Description:     fmt.Sprintf("snapshot of fungible token %s balances", airdropInfo.RequiredDenom),
		UserDescription: airdropInfo.Description,
	})
	if err != nil {
		return err
	}
	airdrop := &types.Airdrop{
		Id:            id,
		Sender:        airdropInfo.Sender,
		Height:        airdropInfo.Height,
		Description:   airdropInfo.Description,
		RequiredDenom: airdropInfo.RequiredDenom,
		Offer:         airdropInfo.Offer,
	}

	store.Set(types.GetAirdropKey(airdrop.RequiredDenom, id), k.cdc.MustMarshal(airdrop))
	return nil
}

func (k Keeper) List(ctx sdk.Context, denom string) []types.Airdrop {
	// FIXME (wojtek): add pagination

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetAirdropDenomPrefix(denom))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var res []types.Airdrop
	for ; iterator.Valid(); iterator.Next() {
		var request types.Airdrop
		k.cdc.MustUnmarshal(iterator.Value(), &request)
		res = append(res, request)
	}

	return res
}
