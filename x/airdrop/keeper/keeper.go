package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/airdrop/types"
	coreumbanktypes "github.com/CoreumFoundation/coreum/x/bank/types"
	snapshottypes "github.com/CoreumFoundation/coreum/x/snapshot/types"
)

type SnapshotKeeper interface {
	RequestSnapshot(ctx sdk.Context, request snapshottypes.SnapshotRequestInfo) (uint64, error)
	GetSnapshot(ctx sdk.Context, owner sdk.AccAddress, snapshotID uint64) (snapshottypes.Snapshot, error)
	GetValueFromSnapshot(ctx sdk.Context, snapshotKey snapshottypes.SnapshotKey, key []byte) ([]byte, error)
	ClaimFromSnapshot(ctx sdk.Context, snapshotKey snapshottypes.SnapshotKey, key []byte) error
}

type BankKeeper interface {
	MintCoins(ctx sdk.Context, moduleName string, amounts sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// Keeper is the asset module keeper.
type Keeper struct {
	cdc            codec.BinaryCodec
	storeKey       sdk.StoreKey
	snapshotKeeper SnapshotKeeper
	bankKeeper     BankKeeper
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey sdk.StoreKey,
	snapshotKeeper SnapshotKeeper,
	bankKeeper BankKeeper,
) Keeper {
	return Keeper{
		cdc:            cdc,
		storeKey:       storeKey,
		snapshotKeeper: snapshotKeeper,
		bankKeeper:     bankKeeper,
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

func (k Keeper) Claim(ctx sdk.Context, denom string, airdropID uint64, recipient sdk.AccAddress) error {
	// FIXME (wojtek): DANGER!!! This method, together with the rest of airdrop module, is a free ATM which might produce
	// any amount of any token (ucore too). Don't merge it into master. It is only for PoC purposes.

	airdrop, err := k.airdrop(ctx, denom, airdropID)
	if err != nil {
		return err
	}

	snapshot, err := k.snapshotKeeper.GetSnapshot(ctx, sdk.MustAccAddressFromBech32(airdrop.Sender), airdrop.Id)
	if err != nil {
		return err
	}

	bz, err := k.snapshotKeeper.GetValueFromSnapshot(ctx, snapshot.Key, coreumbanktypes.AccountKey(recipient))
	if err != nil {
		return err
	}
	if bz == nil {
		return errors.New("balance does not exist in snapshot")
	}
	var requiredCoin sdk.Coin
	k.cdc.MustUnmarshal(bz, &requiredCoin)
	amount := requiredCoin.Amount.ToDec()

	coinsToClaim := make(sdk.Coins, 0, len(airdrop.Offer))
	for _, coin := range airdrop.Offer {
		coinsToClaim = append(coinsToClaim, sdk.NewCoin(coin.Denom, coin.Amount.Mul(amount).TruncateInt()))
	}

	err = k.snapshotKeeper.ClaimFromSnapshot(ctx, snapshot.Key, coreumbanktypes.AccountKey(recipient))
	if err != nil {
		return err
	}

	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coinsToClaim); err != nil {
		return err
	}
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coinsToClaim)
}

func (k Keeper) airdrop(ctx sdk.Context, denom string, airdropID uint64) (types.Airdrop, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetAirdropKey(denom, airdropID))
	if bz == nil {
		return types.Airdrop{}, errors.New("airdrop does not exist")
	}

	var res types.Airdrop
	k.cdc.MustUnmarshal(bz, &res)
	return res, nil
}
