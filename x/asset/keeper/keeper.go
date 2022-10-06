package keeper

import (
	"encoding/binary"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// Keeper is the asset module keeper.
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   sdk.StoreKey
	bankKeeper types.BankKeeper
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, bankKeeper types.BankKeeper) Keeper {
	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		bankKeeper: bankKeeper,
	}
}

// IssueAsset issues new asset.
func (k Keeper) IssueAsset(ctx sdk.Context, definition types.AssetDefinition) (uint64, error) {
	recipient, err := sdk.AccAddressFromBech32(definition.Recipient)
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "can't decode %s recipient address to AccAddress", definition.Recipient)
	}

	var id uint64
	switch definition.Type {
	case types.AssetType_FT: //nolint:nosnakecase // protogen
		id, err = k.issueFTAsset(ctx, definition, recipient)
		if err != nil {
			return 0, sdkerrors.Wrap(err, "can't issue FT asset")
		}
	case types.AssetType_NFT: //nolint:nosnakecase // protogen
		return 0, sdkerrors.Wrapf(types.ErrInvalidAsset, "asset module doesn't support the NFT issuance yet")
	}

	if err = ctx.EventManager().EmitTypedEvent(&types.EventAssetIssued{
		Id: id,
	}); err != nil {
		return 0, sdkerrors.Wrap(err, "can't emit EventAssetIssued event")
	}

	return id, nil
}

// GetAsset return the asset by its id.
func (k Keeper) GetAsset(ctx sdk.Context, id uint64) (types.Asset, error) {
	store := ctx.KVStore(k.storeKey)
	store.Get(types.GetAssetFTKey(id))

	bz := store.Get(types.GetAssetFTKey(id))
	if bz == nil {
		return types.Asset{}, sdkerrors.Wrap(types.ErrNotFound, "asset")
	}
	var definition types.AssetDefinition
	k.cdc.MustUnmarshal(bz, &definition)

	return types.Asset{
		Id:         id,
		Definition: &definition,
	}, nil
}

// Logger returns the Keeper logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) issueFTAsset(ctx sdk.Context, definition types.AssetDefinition, recipient sdk.AccAddress) (uint64, error) {
	id := k.incrementAssetID(ctx)
	// register the denom metadata in the bank module
	denomName, denomBaseName, err := k.createFTDenomMetadata(ctx, definition.Code, definition.Description, definition.Ft.Precision, id)
	if err != nil {
		return 0, err
	}
	// set the denom names to be stored in the keeper
	definition.Ft.DenomName = denomName
	definition.Ft.DenomBaseName = denomBaseName

	// mint the initial amount
	if definition.Ft.InitialAmount.IsPositive() {
		if err := k.mintFTWithPrecision(ctx, denomBaseName, definition.Ft.InitialAmount, definition.Ft.Precision, recipient); err != nil {
			return 0, err
		}
	}
	// store the new asset
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetAssetFTKey(id), k.cdc.MustMarshal(&definition))

	k.Logger(ctx).Debug("issued new asset %s with id %d", denomBaseName, id)

	return id, nil
}

func (k Keeper) incrementAssetID(ctx sdk.Context) uint64 {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.AssetSequenceKey)
	id := uint64(1)
	if bz != nil {
		id = binary.BigEndian.Uint64(bz)
	}
	bz = sdk.Uint64ToBigEndian(id + 1)
	store.Set(types.AssetSequenceKey, bz)
	return id
}

func (k Keeper) createFTDenomMetadata(ctx sdk.Context, code, description string, precision uint32, assetID uint64) (string, string, error) {
	denomName := fmt.Sprintf("%s%s%d", types.ModuleName, code, assetID)
	denomBaseName := fmt.Sprintf("b%s", denomName)
	// in case the precision is zero the name is equal the base name
	if precision == 0 {
		denomBaseName = denomName
	}

	if _, found := k.bankKeeper.GetDenomMetaData(ctx, denomBaseName); found {
		return "", "", sdkerrors.Wrapf(types.ErrInvalidState, "found unexpected denom metadata %s", denomBaseName)
	}
	denomMetadata := banktypes.Metadata{
		Name:        denomName,
		Symbol:      denomName,
		Description: description,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denomBaseName,
				Exponent: uint32(0),
			},
		},
		Base:    denomBaseName,
		Display: denomName,
	}
	// Add additional denom unit in case the precision is not zero
	if precision > 0 {
		denomMetadata.DenomUnits = append(denomMetadata.DenomUnits, &banktypes.DenomUnit{
			Denom:    denomName,
			Exponent: precision,
		})
	}

	k.bankKeeper.SetDenomMetaData(ctx, denomMetadata)

	return denomName, denomBaseName, nil
}

func (k Keeper) mintFTWithPrecision(ctx sdk.Context, denom string, amount sdk.Int, precision uint32, recipient sdk.AccAddress) error {
	initialAmount := amount.Mul(sdk.NewIntWithDecimal(1, int(precision)))
	coinsToMint := sdk.NewCoins(sdk.NewCoin(denom, initialAmount))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coinsToMint); err != nil {
		return sdkerrors.Wrapf(err, "can't mint %s for the module %s", coinsToMint.String(), types.ModuleName)
	}

	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coinsToMint)
}
