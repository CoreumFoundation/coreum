package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum/x/asset/types"
	snapshottypes "github.com/CoreumFoundation/coreum/x/snapshot/types"
)

func (k Keeper) AirdropFungibleToken(ctx sdk.Context, airdropInfo types.AirdropFungibleTokenInfo) error {
	// FIXME (wojtek): verify that denom exists
	// FIXME (wojtek): take control over funds required for airdrop

	store := ctx.KVStore(k.storeKey)

	id := sdk.ZeroInt()
	bz := store.Get(types.AirdropIDGeneratorKey)
	if bz != nil {
		must.OK(id.Unmarshal(bz))
		id = id.Add(sdk.OneInt())
	}
	store.Set(types.AirdropIDGeneratorKey, must.Bytes(id.Marshal()))

	snapshotID, err := k.snapshotKeeper.RequestSnapshot(ctx, snapshottypes.SnapshotRequestInfo{
		Prefix: snapshottypes.SnapshotPrefix{
			StoreName: banktypes.StoreKey,
			Name:      types.BalancesSnapshotName(airdropInfo.RequiredDenom),
		},
		Owner:           airdropInfo.Sender,
		Height:          airdropInfo.Height,
		Description:     fmt.Sprintf("snapshot of fungible token %s balances", airdropInfo.RequiredDenom),
		UserDescription: airdropInfo.Description,
	})
	if err != nil {
		return err
	}
	airdrop := &types.AirdropFungibleToken{
		Id:            id,
		Sender:        airdropInfo.Sender,
		SnapshotId:    snapshotID,
		Height:        airdropInfo.Height,
		Description:   airdropInfo.Description,
		RequiredDenom: airdropInfo.RequiredDenom,
		Offer:         airdropInfo.Offer,
	}

	store.Set(types.GetAirdropKey(airdrop.RequiredDenom, id), k.cdc.MustMarshal(airdrop))
	return nil
}

func (k Keeper) GetAirdropsFungibleToken(ctx sdk.Context, denom string) []types.AirdropFungibleToken {
	// FIXME (wojtek): add pagination

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GetAirdropDenomPrefix(denom))
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	var res []types.AirdropFungibleToken
	for ; iterator.Valid(); iterator.Next() {
		var request types.AirdropFungibleToken
		k.cdc.MustUnmarshal(iterator.Value(), &request)
		res = append(res, request)
	}

	return res
}
