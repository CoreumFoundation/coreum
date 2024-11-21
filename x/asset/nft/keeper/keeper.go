package keeper

import (
	"bytes"

	sdkstore "cosmossdk.io/core/store"
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/nft"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/pkg/errors"

	pkgstore "github.com/CoreumFoundation/coreum/v5/pkg/store"
	"github.com/CoreumFoundation/coreum/v5/x/asset/nft/types"
)

// Keeper is the asset module non-fungible token nftKeeper.
type Keeper struct {
	cdc          codec.BinaryCodec
	storeService sdkstore.KVStoreService
	nftKeeper    types.NFTKeeper
	bankKeeper   types.BankKeeper
	authority    string
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService sdkstore.KVStoreService,
	nftKeeper types.NFTKeeper,
	bankKeeper types.BankKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:          cdc,
		storeService: storeService,
		nftKeeper:    nftKeeper,
		bankKeeper:   bankKeeper,
		authority:    authority,
	}
}

// GetParams gets the parameters of the module.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	bz, _ := k.storeService.OpenKVStore(ctx).Get(types.ParamsKey)
	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the parameters of the module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.ParamsKey, bz)
}

// UpdateParams is a governance operation that sets parameters of the module.
func (k Keeper) UpdateParams(ctx sdk.Context, authority string, params types.Params) error {
	if k.authority != authority {
		return sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, authority)
	}

	return k.SetParams(ctx, params)
}

// GetClass reruns the Class.
func (k Keeper) GetClass(ctx sdk.Context, classID string) (types.Class, error) {
	definition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return types.Class{}, err
	}

	class, found := k.nftKeeper.GetClass(ctx, classID)
	if !found {
		return types.Class{}, sdkerrors.Wrapf(types.ErrClassNotFound, "nft class with ID:%s not found", classID)
	}

	return types.Class{
		Id:          class.Id,
		Issuer:      definition.Issuer,
		Name:        class.Name,
		Symbol:      class.Symbol,
		Description: class.Description,
		URI:         class.Uri,
		URIHash:     class.UriHash,
		Data:        class.Data,
		Features:    definition.Features,
		RoyaltyRate: definition.RoyaltyRate,
	}, nil
}

// GetClasses returns the classes list, argument issuer is optional.
func (k Keeper) GetClasses(
	ctx sdk.Context, issuer *sdk.AccAddress, pagination *query.PageRequest,
) ([]types.Class, *query.PageResponse, error) {
	definitions, pageRes, err := k.GetClassDefinitions(ctx, issuer, pagination)
	if err != nil {
		return nil, nil, err
	}

	classes := make([]types.Class, 0, len(definitions))
	for _, definition := range definitions {
		class, err := k.GetClass(ctx, definition.ID)
		if err != nil {
			return nil, nil, err
		}
		classes = append(classes, class)
	}

	return classes, pageRes, nil
}

// IssueClass issues new non-fungible token class and returns its id.
func (k Keeper) IssueClass(ctx sdk.Context, settings types.IssueClassSettings) (string, error) {
	if err := types.ValidateClassSymbol(settings.Symbol); err != nil {
		return "", err
	}

	if err := types.ValidateClassFeatures(settings.Features); err != nil {
		return "", err
	}

	if err := types.ValidateRoyaltyRate(settings.RoyaltyRate); err != nil {
		return "", err
	}

	id := types.BuildClassID(settings.Symbol, settings.Issuer)
	if err := types.ValidateClassData(settings.Data); err != nil {
		return "", sdkerrors.Wrap(types.ErrInvalidInput, err.Error())
	}

	found := k.nftKeeper.HasClass(ctx, id)
	if found {
		return "", sdkerrors.Wrapf(
			types.ErrInvalidInput,
			"symbol %q already used for the address %q",
			settings.Symbol,
			settings.Issuer,
		)
	}

	if err := k.nftKeeper.SaveClass(ctx, nft.Class{
		Id:          id,
		Symbol:      settings.Symbol,
		Name:        settings.Name,
		Description: settings.Description,
		Uri:         settings.URI,
		UriHash:     settings.URIHash,
		Data:        settings.Data,
	}); err != nil {
		return "", sdkerrors.Wrapf(types.ErrInvalidInput, "can't save non-fungible token: %s", err)
	}

	if err := k.SetClassDefinition(ctx, types.ClassDefinition{
		ID:          id,
		Issuer:      settings.Issuer.String(),
		Features:    settings.Features,
		RoyaltyRate: settings.RoyaltyRate,
	}); err != nil {
		return "", err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventClassIssued{
		ID:          id,
		Issuer:      settings.Issuer.String(),
		Symbol:      settings.Symbol,
		Name:        settings.Name,
		Description: settings.Description,
		URI:         settings.URI,
		URIHash:     settings.URIHash,
		Features:    settings.Features,
		RoyaltyRate: settings.RoyaltyRate,
	}); err != nil {
		return "", sdkerrors.Wrapf(types.ErrInvalidInput, "failed to emit event EventClassIssued: %s", err)
	}

	return id, nil
}

// GetClassDefinition reruns the ClassDefinition.
func (k Keeper) GetClassDefinition(ctx sdk.Context, classID string) (types.ClassDefinition, error) {
	if _, _, err := types.DeconstructClassID(classID); err != nil {
		return types.ClassDefinition{}, err
	}

	classKey, err := types.CreateClassKey(classID)
	if err != nil {
		return types.ClassDefinition{}, err
	}

	bz, _ := k.storeService.OpenKVStore(ctx).Get(classKey)
	if bz == nil {
		return types.ClassDefinition{}, sdkerrors.Wrapf(types.ErrClassNotFound, "classID: %s", classID)
	}
	var definition types.ClassDefinition
	k.cdc.MustUnmarshal(bz, &definition)

	return definition, nil
}

// IterateAllClassDefinitions iterates over all class definitions and applies the provided callback.
// If true is returned from the callback, iteration is halted.
func (k Keeper) IterateAllClassDefinitions(ctx sdk.Context, cb func(types.ClassDefinition) (bool, error)) error {
	store := k.storeService.OpenKVStore(ctx)
	iterator := storetypes.KVStorePrefixIterator(runtime.KVStoreAdapter(store), types.NFTClassKeyPrefix)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var definition types.ClassDefinition
		k.cdc.MustUnmarshal(iterator.Value(), &definition)

		stop, err := cb(definition)
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}
	return nil
}

// GetClassDefinitions returns all non-fungible class token definitions.
func (k Keeper) GetClassDefinitions(
	ctx sdk.Context, issuer *sdk.AccAddress, pagination *query.PageRequest,
) ([]types.ClassDefinition, *query.PageResponse, error) {
	store := k.storeService.OpenKVStore(ctx)
	fetchingKey := types.NFTClassKeyPrefix
	if issuer != nil {
		var err error
		fetchingKey, err = types.CreateIssuerClassPrefix(*issuer)
		if err != nil {
			return nil, nil, err
		}
	}
	definitionsPointers, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		prefix.NewStore(runtime.KVStoreAdapter(store), fetchingKey),
		pagination,
		// builder
		func(key []byte, definition *types.ClassDefinition) (*types.ClassDefinition, error) {
			return definition, nil
		},
		// constructor
		func() *types.ClassDefinition {
			return &types.ClassDefinition{}
		},
	)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(types.ErrInvalidInput, "failed to paginate, err: %s", err)
	}

	definitions := make([]types.ClassDefinition, 0, len(definitionsPointers))
	for _, definition := range definitionsPointers {
		definitions = append(definitions, *definition)
	}

	return definitions, pageRes, nil
}

// SetClassDefinition stores the ClassDefinition.
func (k Keeper) SetClassDefinition(ctx sdk.Context, definition types.ClassDefinition) error {
	classKey, err := types.CreateClassKey(definition.ID)
	if err != nil {
		return err
	}

	return k.storeService.OpenKVStore(ctx).Set(classKey, k.cdc.MustMarshal(&definition))
}

// Mint mints new non-fungible token.
func (k Keeper) Mint(ctx sdk.Context, settings types.MintSettings) error {
	if err := types.ValidateTokenID(settings.ID); err != nil {
		return sdkerrors.Wrap(types.ErrInvalidInput, err.Error())
	}

	if err := types.ValidateNFTData(settings.Data); err != nil {
		return sdkerrors.Wrap(types.ErrInvalidInput, err.Error())
	}

	definition, err := k.GetClassDefinition(ctx, settings.ClassID)
	if err != nil {
		return err
	}

	if !definition.IsIssuer(settings.Sender) {
		return sdkerrors.Wrapf(
			cosmoserrors.ErrUnauthorized,
			"address %q is unauthorized to perform the mint operation",
			settings.Sender.String(),
		)
	}

	if definition.IsFeatureEnabled(types.ClassFeature_whitelisting) && !definition.IsIssuer(settings.Recipient) {
		isWhitelisted, err := k.isClassWhitelisted(ctx, settings.ClassID, settings.Recipient)
		if err != nil {
			return err
		}
		if !isWhitelisted {
			return sdkerrors.Wrapf(
				cosmoserrors.ErrUnauthorized,
				"due to enabled whitelisting only the issuer can receive minted NFT, %s is not the issuer",
				settings.Recipient.String(),
			)
		}
	}

	if !k.nftKeeper.HasClass(ctx, settings.ClassID) {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "classID %q not found", settings.ClassID)
	}

	if k.nftKeeper.HasNFT(ctx, settings.ClassID, settings.ID) {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "ID %q already defined for the class", settings.ID)
	}

	burnt, err := k.IsBurnt(ctx, settings.ClassID, settings.ID)
	if err != nil {
		return err
	}
	if burnt {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "ID %q has been burnt for the class", settings.ID)
	}

	params := k.GetParams(ctx)
	if params.MintFee.IsPositive() {
		coinsToBurn := sdk.NewCoins(params.MintFee)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(
			ctx, settings.Sender, types.ModuleName, coinsToBurn,
		); err != nil {
			return sdkerrors.Wrapf(
				err,
				"can't send coins from account %s to module %s",
				settings.Sender.String(),
				types.ModuleName,
			)
		}
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coinsToBurn); err != nil {
			return sdkerrors.Wrapf(err, "can't burn %s for the module %s", coinsToBurn.String(), types.ModuleName)
		}
	}

	if err := k.nftKeeper.Mint(ctx, nft.NFT{
		ClassId: settings.ClassID,
		Id:      settings.ID,
		Uri:     settings.URI,
		UriHash: settings.URIHash,
		Data:    settings.Data,
	}, settings.Recipient); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "can't save non-fungible token: %s", err)
	}

	return nil
}

// UpdateData updates non-fungible token data.
func (k Keeper) UpdateData(
	ctx sdk.Context,
	sender sdk.AccAddress,
	classID, id string,
	itemsToUpdate []types.DataDynamicIndexedItem,
) error {
	if err := k.validateNFTNotFrozen(ctx, classID, id); err != nil {
		return err
	}

	storedNFT, found := k.nftKeeper.GetNFT(ctx, classID, id)
	if !found {
		return sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, id)
	}
	if storedNFT.Data == nil {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "no data to update classID:%s, ID:%s", classID, id)
	}
	if storedNFT.Data.TypeUrl != "/"+proto.MessageName((*types.DataDynamic)(nil)) {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "nft data type %s is not updatable", storedNFT.Data.TypeUrl)
	}
	var dataDynamic types.DataDynamic
	if err := dataDynamic.Unmarshal(storedNFT.Data.Value); err != nil {
		return sdkerrors.Wrap(types.ErrInvalidInput, "failed to unmarshal DataDynamic data")
	}

	owner := k.nftKeeper.GetOwner(ctx, classID, id)
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return err
	}

	// update dynamic items
	for _, itemToUpdate := range itemsToUpdate {
		if int(itemToUpdate.Index) > len(dataDynamic.Items)-1 {
			return sdkerrors.Wrapf(
				types.ErrInvalidInput, "invalid item, index %d out or range", itemToUpdate.Index,
			)
		}
		storedItem := dataDynamic.Items[int(itemToUpdate.Index)]
		if len(storedItem.Editors) == 0 {
			return sdkerrors.Wrapf(types.ErrInvalidInput, "the item with index %d is not updatable", itemToUpdate.Index)
		}
		updateAllowed, err := isDataDynamicItemUpdateAllowed(sender, owner, classDefinition, storedItem)
		if err != nil {
			return err
		}
		if !updateAllowed {
			return sdkerrors.Wrapf(
				cosmoserrors.ErrUnauthorized,
				"sender is not authorized to update the item with index %d",
				itemToUpdate.Index,
			)
		}

		dataDynamic.Items[int(itemToUpdate.Index)].Data = itemToUpdate.Data
	}
	data, err := codectypes.NewAnyWithValue(&dataDynamic)
	if err != nil {
		return sdkerrors.Wrap(types.ErrInvalidState, "failed to pack to Any type")
	}
	storedNFT.Data = data

	// validate that final data after update is still valid
	if err := types.ValidateNFTData(storedNFT.Data); err != nil {
		return err
	}

	return k.nftKeeper.Update(ctx, storedNFT)
}

// Burn burns non-fungible token.
func (k Keeper) Burn(ctx sdk.Context, owner sdk.AccAddress, classID, id string) error {
	ndfd, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return err
	}

	if err = ndfd.CheckFeatureAllowed(owner, types.ClassFeature_burning); err != nil {
		return err
	}

	if !k.nftKeeper.HasNFT(ctx, classID, id) {
		return sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, id)
	}

	if k.nftKeeper.GetOwner(ctx, classID, id).String() != owner.String() {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "only owner can burn the nft")
	}

	if err := k.checkBurnable(ctx, owner, ndfd, classID, id); err != nil {
		return err
	}

	// If the token is burnt the storage needs to be cleaned up.
	// We clean freezing because it's a single record only.
	// We don't clean whitelisting because potential number of records is unlimited.
	if err := k.SetFrozen(ctx, classID, id, false); err != nil {
		return err
	}

	if err := k.nftKeeper.Burn(ctx, classID, id); err != nil {
		return err
	}

	return k.SetBurnt(ctx, classID, id)
}

func (k Keeper) checkBurnable(
	ctx sdk.Context, owner sdk.AccAddress, ndfd types.ClassDefinition, classID, nftID string,
) error {
	frozen, err := k.IsFrozen(ctx, classID, nftID)
	if err != nil && !errors.Is(err, types.ErrFeatureDisabled) {
		return err
	}

	// non issuer is not allowed to burn frozen NFT, but the issuer can
	if frozen && owner.String() != ndfd.Issuer {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "frozen token cannot be burnt")
	}

	return nil
}

// IsBurnt return whether a non-fungible token is burnt or not.
func (k Keeper) IsBurnt(ctx sdk.Context, classID, nftID string) (bool, error) {
	key, err := types.CreateBurningKey(classID, nftID)
	if err != nil {
		return false, err
	}

	isBurnt, _ := k.storeService.OpenKVStore(ctx).Get(key)
	return bytes.Equal(isBurnt, types.StoreTrue), nil
}

// GetBurntByClass return the list of burnt NFTs in class.
func (k Keeper) GetBurntByClass(
	ctx sdk.Context, classID string, q *query.PageRequest,
) (*query.PageResponse, []string, error) {
	store := k.storeService.OpenKVStore(ctx)
	key, err := types.CreateClassBurningKey(classID)
	if err != nil {
		return nil, nil, err
	}

	nfts := []string{}
	pageRes, err := query.Paginate(prefix.NewStore(runtime.KVStoreAdapter(store), key), q,
		func(key, value []byte) error {
			if !bytes.Equal(value, types.StoreTrue) {
				return sdkerrors.Wrapf(
					types.ErrInvalidState,
					"value stored in burnt store is not %x, value %x",
					types.StoreTrue,
					value,
				)
			}

			nft := string(key[1:]) // the first byte contains the length prefix
			nfts = append(nfts, nft)
			return nil
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return pageRes, nfts, nil
}

// SetBurnt marks the nft burnt, but does not make any checks
// should not be used directly outside the module except for genesis.
func (k Keeper) SetBurnt(ctx sdk.Context, classID, nftID string) error {
	if k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "nft with classID:%s and ID:%s exists", classID, nftID)
	}
	burnt, err := k.IsBurnt(ctx, classID, nftID)
	if err != nil {
		return err
	}
	if burnt {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "nft with classID:%s and ID:%s has been already burnt", classID, nftID)
	}

	key, err := types.CreateBurningKey(classID, nftID)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(key, types.StoreTrue)
}

// GetBurntNFTs return paginated burnt NFTs.
//
//nolint:dupl
func (k Keeper) GetBurntNFTs(
	ctx sdk.Context, q *query.PageRequest,
) ([]types.BurntNFT, *query.PageResponse, error) {
	store := k.storeService.OpenKVStore(ctx)
	burnt := make([]types.BurntNFT, 0)
	classIDToBurntNFTIdx := make(map[string]int)
	pageRes, err := query.Paginate(prefix.NewStore(runtime.KVStoreAdapter(store), types.NFTBurningKeyPrefix),
		q, func(key, value []byte) error {
			if !bytes.Equal(value, types.StoreTrue) {
				return sdkerrors.Wrapf(
					types.ErrInvalidState,
					"value stored in burning store is not %x, value %x",
					types.StoreTrue,
					value,
				)
			}
			classID, nftID, err := types.ParseBurningKey(key)
			if err != nil {
				return err
			}

			idx, ok := classIDToBurntNFTIdx[classID]
			if ok {
				burnt[idx].NftIDs = append(burnt[idx].NftIDs, nftID)
				return nil
			}

			burnt = append(burnt, types.BurntNFT{
				ClassID: classID,
				NftIDs:  []string{nftID},
			})
			classIDToBurntNFTIdx[classID] = len(burnt) - 1
			return nil
		})
	if err != nil {
		return nil, nil, err
	}

	return burnt, pageRes, nil
}

// Freeze freezes a non-fungible token.
func (k Keeper) Freeze(ctx sdk.Context, sender sdk.AccAddress, classID, nftID string) error {
	return k.freezeOrUnfreeze(ctx, sender, classID, nftID, true)
}

// Unfreeze unfreezes a non-fungible token.
func (k Keeper) Unfreeze(ctx sdk.Context, sender sdk.AccAddress, classID, nftID string) error {
	return k.freezeOrUnfreeze(ctx, sender, classID, nftID, false)
}

// SetFrozen marks the nft frozen, but does not make any checks
// should not be used directly outside the module except for genesis.
func (k Keeper) SetFrozen(ctx sdk.Context, classID, nftID string, frozen bool) error {
	key, err := types.CreateFreezingKey(classID, nftID)
	if err != nil {
		return err
	}
	store := k.storeService.OpenKVStore(ctx)
	if frozen {
		return store.Set(key, types.StoreTrue)
	}
	return store.Delete(key)
}

// ClassFreeze freezes a non-fungible token.
func (k Keeper) ClassFreeze(ctx sdk.Context, sender, account sdk.AccAddress, classID string) error {
	return k.classFreezeOrUnfreeze(ctx, sender, account, classID, true)
}

// ClassUnfreeze unfreezes a non-fungible token.
func (k Keeper) ClassUnfreeze(ctx sdk.Context, sender, account sdk.AccAddress, classID string) error {
	return k.classFreezeOrUnfreeze(ctx, sender, account, classID, false)
}

// SetClassFrozen marks the nft class as for an account, but does not make any checks
// should not be used directly outside the module except for genesis.
func (k Keeper) SetClassFrozen(ctx sdk.Context, classID string, account sdk.AccAddress, frozen bool) error {
	key, err := types.CreateClassFreezingKey(classID, account)
	if err != nil {
		return err
	}
	store := k.storeService.OpenKVStore(ctx)
	if frozen {
		return store.Set(key, types.StoreTrue)
	}
	return store.Delete(key)
}

// IsFrozen return whether a non-fungible token is frozen or not.
func (k Keeper) IsFrozen(ctx sdk.Context, classID, nftID string) (bool, error) {
	store := k.storeService.OpenKVStore(ctx)
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return false, err
	}

	if !classDefinition.IsFeatureEnabled(types.ClassFeature_freezing) {
		return false, sdkerrors.Wrapf(types.ErrFeatureDisabled, `feature "freezing" is disabled`)
	}

	if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return false, sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, nftID)
	}

	key, err := types.CreateFreezingKey(classID, nftID)
	if err != nil {
		return false, err
	}

	val, _ := store.Get(key)
	if bytes.Equal(val, types.StoreTrue) {
		return true, nil
	}

	owner := k.nftKeeper.GetOwner(ctx, classID, nftID)
	key, err = types.CreateClassFreezingKey(classID, owner)
	if err != nil {
		return false, err
	}

	val, _ = store.Get(key)
	return bytes.Equal(val, types.StoreTrue), nil
}

// IsClassFrozen return whether an account is frozen for an NFT class  .
func (k Keeper) IsClassFrozen(ctx sdk.Context, classID string, account sdk.AccAddress) (bool, error) {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return false, err
	}

	if !classDefinition.IsFeatureEnabled(types.ClassFeature_freezing) {
		return false, sdkerrors.Wrapf(types.ErrFeatureDisabled, `feature "freezing" is disabled`)
	}

	if !k.nftKeeper.HasClass(ctx, classID) {
		return false, sdkerrors.Wrapf(types.ErrNFTNotFound, "class with ID:%s not found", classID)
	}

	key, err := types.CreateClassFreezingKey(classID, account)
	if err != nil {
		return false, err
	}

	val, _ := k.storeService.OpenKVStore(ctx).Get(key)
	return bytes.Equal(val, types.StoreTrue), nil
}

// GetFrozenNFTs return paginated frozen NFTs.
//
//nolint:dupl
func (k Keeper) GetFrozenNFTs(ctx sdk.Context, q *query.PageRequest) ([]types.FrozenNFT, *query.PageResponse, error) {
	store := k.storeService.OpenKVStore(ctx)
	frozen := make([]types.FrozenNFT, 0)
	classIDToFrozenNFTIdx := make(map[string]int)
	pageRes, err := query.Paginate(prefix.NewStore(runtime.KVStoreAdapter(store), types.NFTFreezingKeyPrefix),
		q, func(key, value []byte) error {
			if !bytes.Equal(value, types.StoreTrue) {
				return sdkerrors.Wrapf(
					types.ErrInvalidState,
					"value stored in freezing store is not %x, value %x",
					types.StoreTrue,
					value,
				)
			}
			classID, nftID, err := types.ParseFreezingKey(key)
			if err != nil {
				return err
			}

			idx, ok := classIDToFrozenNFTIdx[classID]
			if ok {
				frozen[idx].NftIDs = append(frozen[idx].NftIDs, nftID)
				return nil
			}

			frozen = append(frozen, types.FrozenNFT{
				ClassID: classID,
				NftIDs:  []string{nftID},
			})
			classIDToFrozenNFTIdx[classID] = len(frozen) - 1
			return nil
		})
	if err != nil {
		return nil, nil, err
	}

	return frozen, pageRes, nil
}

// GetAllClassFrozenAccounts returns all frozen accounts for all NFTs.
//
//nolint:dupl // merging the code under a common abstraction will make it less maintainable.
func (k Keeper) GetAllClassFrozenAccounts(
	ctx sdk.Context, q *query.PageRequest,
) ([]types.ClassFrozenAccounts, *query.PageResponse, error) {
	store := k.storeService.OpenKVStore(ctx)
	frozen := make([]types.ClassFrozenAccounts, 0)
	classIDToFrozenIdx := make(map[string]int)
	pageRes, err := query.Paginate(prefix.NewStore(runtime.KVStoreAdapter(store), types.NFTClassFreezingKeyPrefix),
		q, func(key, value []byte) error {
			if !bytes.Equal(value, types.StoreTrue) {
				return sdkerrors.Wrapf(
					types.ErrInvalidState,
					"value stored in whitelisting store is not %x, value %x",
					types.StoreTrue,
					value,
				)
			}
			classID, account, err := types.ParseClassFreezingKey(key)
			if err != nil {
				return err
			}
			if !k.nftKeeper.HasClass(ctx, classID) {
				return nil
			}

			idx, ok := classIDToFrozenIdx[classID]
			if ok {
				frozen[idx].Accounts = append(frozen[idx].Accounts, account.String())
				return nil
			}

			frozen = append(frozen, types.ClassFrozenAccounts{
				ClassID:  classID,
				Accounts: []string{account.String()},
			})
			classIDToFrozenIdx[classID] = len(frozen) - 1
			return nil
		})
	if err != nil {
		return nil, nil, err
	}

	return frozen, pageRes, nil
}

// GetClassFrozenAccounts returns all class frozen accounts for the class.
func (k Keeper) GetClassFrozenAccounts(
	ctx sdk.Context, classID string, q *query.PageRequest,
) ([]string, *query.PageResponse, error) {
	store := k.storeService.OpenKVStore(ctx)
	compositeKey, err := pkgstore.JoinKeysWithLength([]byte(classID))
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(types.ErrInvalidKey, "failed to create a composite key for nft, err: %s", err)
	}
	key := pkgstore.JoinKeys(types.NFTClassFreezingKeyPrefix, compositeKey)
	accounts := []string{}
	pageRes, err := query.Paginate(prefix.NewStore(runtime.KVStoreAdapter(store), key),
		q, func(key, value []byte) error {
			if !bytes.Equal(value, types.StoreTrue) {
				return sdkerrors.Wrapf(
					types.ErrInvalidState,
					"value stored in whitelisting store is not %x, value %x",
					types.StoreTrue,
					value,
				)
			}

			account := sdk.AccAddress(key[1:]) // the first byte contains the length prefix
			accounts = append(accounts, account.String())
			return nil
		})
	if err != nil {
		return nil, nil, err
	}

	return accounts, pageRes, nil
}

// IsWhitelisted checks to see if an account is whitelisted for an NFT.
func (k Keeper) IsWhitelisted(ctx sdk.Context, classID, nftID string, account sdk.AccAddress) (bool, error) {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return false, err
	}

	if !classDefinition.IsFeatureEnabled(types.ClassFeature_whitelisting) {
		return false, sdkerrors.Wrapf(types.ErrFeatureDisabled, `feature "whitelisting" is disabled`)
	}

	isWhitelisted, err := k.isTokenWhitelisted(ctx, classID, nftID, account)
	if err != nil {
		return false, err
	}
	if isWhitelisted {
		return true, nil
	}

	return k.isClassWhitelisted(ctx, classID, account)
}

func isDataDynamicItemUpdateAllowed(
	sender sdk.AccAddress,
	owner sdk.AccAddress,
	classDefinition types.ClassDefinition,
	item types.DataDynamicItem,
) (bool, error) {
	for _, editor := range item.Editors {
		switch editor {
		case types.DataEditor_admin:
			// TODO(v5) use admin instead of issuer once the admin is introduced
			if classDefinition.IsIssuer(sender) {
				return true, nil
			}
		case types.DataEditor_owner:
			if owner.String() == sender.String() {
				return true, nil
			}
		default:
			return false, sdkerrors.Wrapf(types.ErrInvalidState, "unsupported editor %d", editor)
		}
	}
	return false, nil
}

func (k Keeper) isClassWhitelisted(ctx sdk.Context, classID string, account sdk.AccAddress) (bool, error) {
	if !k.nftKeeper.HasClass(ctx, classID) {
		return false, sdkerrors.Wrapf(types.ErrNFTNotFound, "nft class with classID:%s not found", classID)
	}

	classKey, err := types.CreateClassWhitelistingKey(classID, account)
	if err != nil {
		return false, err
	}

	val, _ := k.storeService.OpenKVStore(ctx).Get(classKey)
	return bytes.Equal(val, types.StoreTrue), nil
}

func (k Keeper) isTokenWhitelisted(ctx sdk.Context, classID, nftID string, account sdk.AccAddress) (bool, error) {
	if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return false, sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, nftID)
	}

	key, err := types.CreateWhitelistingKey(classID, nftID, account)
	if err != nil {
		return false, err
	}

	val, _ := k.storeService.OpenKVStore(ctx).Get(key)
	return bytes.Equal(val, types.StoreTrue), nil
}

// GetWhitelistedAccountsForNFT returns all whitelisted accounts for all NFTs.
func (k Keeper) GetWhitelistedAccountsForNFT(
	ctx sdk.Context, classID, nftID string, q *query.PageRequest,
) ([]string, *query.PageResponse, error) {
	store := k.storeService.OpenKVStore(ctx)
	if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return nil, nil, sdkerrors.Wrapf(
			types.ErrNFTNotFound,
			"nft with classID:%s and ID:%s not found",
			classID,
			nftID,
		)
	}

	compositeKey, err := pkgstore.JoinKeysWithLength([]byte(classID), []byte(nftID))
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(
			types.ErrInvalidKey,
			"failed to create a composite key for nft, err: %s",
			err,
		)
	}
	key := pkgstore.JoinKeys(types.NFTWhitelistingKeyPrefix, compositeKey)
	accounts := []string{}
	pageRes, err := query.Paginate(prefix.NewStore(runtime.KVStoreAdapter(store), key),
		q, func(key, value []byte) error {
			if !bytes.Equal(value, types.StoreTrue) {
				return sdkerrors.Wrapf(
					types.ErrInvalidState,
					"value stored in whitelisting store is not %x, value %x",
					types.StoreTrue,
					value,
				)
			}

			account := sdk.AccAddress(key[1:]) // the first byte contains the length prefix
			accounts = append(accounts, account.String())
			return nil
		})
	if err != nil {
		return nil, nil, err
	}

	return accounts, pageRes, nil
}

// GetWhitelistedAccounts returns all whitelisted accounts for all NFTs.
func (k Keeper) GetWhitelistedAccounts(
	ctx sdk.Context, q *query.PageRequest,
) ([]types.WhitelistedNFTAccounts, *query.PageResponse, error) {
	store := k.storeService.OpenKVStore(ctx)
	type nftUniqueID struct {
		classID string
		nftID   string
	}
	whitelisted := make([]types.WhitelistedNFTAccounts, 0)
	nftUniqueIDToWhitelistIdx := make(map[nftUniqueID]int)
	pageRes, err := query.Paginate(prefix.NewStore(runtime.KVStoreAdapter(store), types.NFTWhitelistingKeyPrefix),
		q, func(key, value []byte) error {
			if !bytes.Equal(value, types.StoreTrue) {
				return sdkerrors.Wrapf(
					types.ErrInvalidState,
					"value stored in whitelisting store is not %x, value %x",
					types.StoreTrue,
					value,
				)
			}
			classID, nftID, account, err := types.ParseWhitelistingKey(key)
			if err != nil {
				return err
			}
			if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
				return nil
			}

			uniqueID := nftUniqueID{
				classID: classID,
				nftID:   nftID,
			}

			idx, ok := nftUniqueIDToWhitelistIdx[uniqueID]
			if ok {
				whitelisted[idx].Accounts = append(whitelisted[idx].Accounts, account.String())
				return nil
			}

			whitelisted = append(whitelisted, types.WhitelistedNFTAccounts{
				ClassID:  classID,
				NftID:    nftID,
				Accounts: []string{account.String()},
			})
			nftUniqueIDToWhitelistIdx[uniqueID] = len(whitelisted) - 1
			return nil
		})
	if err != nil {
		return nil, nil, err
	}

	return whitelisted, pageRes, nil
}

// GetAllClassWhitelistedAccounts returns all whitelisted accounts for all NFTs.
//
//nolint:dupl // merging the code under a common abstraction will make it less maintainable.
func (k Keeper) GetAllClassWhitelistedAccounts(
	ctx sdk.Context, q *query.PageRequest,
) ([]types.ClassWhitelistedAccounts, *query.PageResponse, error) {
	store := k.storeService.OpenKVStore(ctx)
	whitelisted := make([]types.ClassWhitelistedAccounts, 0)
	classIDToWhitelistedIdx := make(map[string]int)
	pageRes, err := query.Paginate(prefix.NewStore(runtime.KVStoreAdapter(store), types.NFTClassWhitelistingKeyPrefix),
		q, func(key, value []byte) error {
			if !bytes.Equal(value, types.StoreTrue) {
				return sdkerrors.Wrapf(
					types.ErrInvalidState,
					"value stored in whitelisting store is not %x, value %x",
					types.StoreTrue,
					value,
				)
			}
			classID, account, err := types.ParseClassWhitelistingKey(key)
			if err != nil {
				return err
			}
			if !k.nftKeeper.HasClass(ctx, classID) {
				return nil
			}

			idx, ok := classIDToWhitelistedIdx[classID]
			if ok {
				whitelisted[idx].Accounts = append(whitelisted[idx].Accounts, account.String())
				return nil
			}

			whitelisted = append(whitelisted, types.ClassWhitelistedAccounts{
				ClassID:  classID,
				Accounts: []string{account.String()},
			})
			classIDToWhitelistedIdx[classID] = len(whitelisted) - 1
			return nil
		})
	if err != nil {
		return nil, nil, err
	}

	return whitelisted, pageRes, nil
}

// GetClassWhitelistedAccounts returns all whitelisted accounts for the class.
func (k Keeper) GetClassWhitelistedAccounts(
	ctx sdk.Context, classID string, q *query.PageRequest,
) ([]string, *query.PageResponse, error) {
	store := k.storeService.OpenKVStore(ctx)
	compositeKey, err := pkgstore.JoinKeysWithLength([]byte(classID))
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(
			types.ErrInvalidKey, "failed to create a composite key for nft, err: %s", err,
		)
	}
	key := pkgstore.JoinKeys(types.NFTClassWhitelistingKeyPrefix, compositeKey)
	accounts := []string{}
	pageRes, err := query.Paginate(prefix.NewStore(runtime.KVStoreAdapter(store), key),
		q, func(key, value []byte) error {
			if !bytes.Equal(value, types.StoreTrue) {
				return sdkerrors.Wrapf(
					types.ErrInvalidState,
					"value stored in whitelisting store is not %x, value %x",
					types.StoreTrue,
					value,
				)
			}

			account := sdk.AccAddress(key[1:]) // the first byte contains the length prefix
			accounts = append(accounts, account.String())
			return nil
		})
	if err != nil {
		return nil, nil, err
	}

	return accounts, pageRes, nil
}

// AddToWhitelist adds an account to the whitelisted list of accounts for the NFT.
func (k Keeper) AddToWhitelist(ctx sdk.Context, classID, nftID string, sender, account sdk.AccAddress) error {
	return k.addToWhitelistOrRemoveFromWhitelist(ctx, classID, nftID, sender, account, true)
}

// RemoveFromWhitelist removes an account from the whitelisted list of accounts for the NFT.
func (k Keeper) RemoveFromWhitelist(ctx sdk.Context, classID, nftID string, sender, account sdk.AccAddress) error {
	return k.addToWhitelistOrRemoveFromWhitelist(ctx, classID, nftID, sender, account, false)
}

// AddToClassWhitelist adds an account to the whitelisted list of accounts for the entire class.
func (k Keeper) AddToClassWhitelist(ctx sdk.Context, classID string, sender, account sdk.AccAddress) error {
	return k.addToWhitelistOrRemoveFromWhitelistClass(ctx, classID, sender, account, true)
}

// RemoveFromClassWhitelist removes an account from the whitelisted list of accounts for the entire class.
func (k Keeper) RemoveFromClassWhitelist(ctx sdk.Context, classID string, sender, account sdk.AccAddress) error {
	return k.addToWhitelistOrRemoveFromWhitelistClass(ctx, classID, sender, account, false)
}

// SetWhitelisting adds an account to the whitelisting of the NFT, if whitelisting is true
// and removes it, if whitelisting is false.
func (k Keeper) SetWhitelisting(
	ctx sdk.Context, classID, nftID string, account sdk.AccAddress, whitelisting bool,
) error {
	key, err := types.CreateWhitelistingKey(classID, nftID, account)
	if err != nil {
		return err
	}
	store := k.storeService.OpenKVStore(ctx)
	if whitelisting {
		return store.Set(key, types.StoreTrue)
	}
	return store.Delete(key)
}

// SetClassWhitelisting adds an account to the whitelisting of the Class, if whitelisting is true
// and removes it, if whitelisting is false.
func (k Keeper) SetClassWhitelisting(
	ctx sdk.Context, classID string, account sdk.AccAddress, whitelisting bool,
) error {
	key, err := types.CreateClassWhitelistingKey(classID, account)
	if err != nil {
		return err
	}
	store := k.storeService.OpenKVStore(ctx)
	if whitelisting {
		return store.Set(key, types.StoreTrue)
	}
	return store.Delete(key)
}

func (k Keeper) isNFTSendable(ctx sdk.Context, classID, nftID string) error {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	// we return nil here, since we want the original tests of the nft module to pass, but they
	// fail if we return errors for unregistered NFTs on asset. Also, the original nft module
	// does not have access to the asset module to register the NFTs
	if types.ErrClassNotFound.Is(err) {
		return nil
	}
	if err != nil {
		return err
	}

	// always allow issuer to send NFTs issued by them.
	owner := k.nftKeeper.GetOwner(ctx, classID, nftID)
	if classDefinition.Issuer == owner.String() {
		return nil
	}

	if classDefinition.IsFeatureEnabled(types.ClassFeature_disable_sending) {
		return sdkerrors.Wrapf(
			cosmoserrors.ErrUnauthorized,
			"nft with classID:%s and ID:%s has sending disabled",
			classID,
			nftID,
		)
	}

	// we check for soulbound only after the check for issuer, since the issuer should be able to send the token.
	if classDefinition.IsFeatureEnabled(types.ClassFeature_soulbound) {
		return sdkerrors.Wrapf(
			cosmoserrors.ErrUnauthorized,
			"nft with classID:%s and ID:%s is soulbound and cannot be sent",
			classID,
			nftID,
		)
	}

	return k.validateNFTNotFrozen(ctx, classID, nftID)
}

func (k Keeper) validateNFTNotFrozen(ctx sdk.Context, classID, nftID string) error {
	// the IsFrozen includes both class and NFT freezing check
	isFrozen, err := k.IsFrozen(ctx, classID, nftID)
	if err != nil {
		if errors.Is(err, types.ErrFeatureDisabled) {
			return nil
		}
		return err
	}
	if isFrozen {
		return sdkerrors.Wrapf(cosmoserrors.ErrUnauthorized, "nft with classID:%s and ID:%s is frozen", classID, nftID)
	}

	return nil
}

// TODO: probably we should path naming `is` -> `validate` here and for all similar.
func (k Keeper) isNFTReceivable(ctx sdk.Context, classID, nftID string, receiver sdk.AccAddress) error {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	// we return nil here, since we want the original tests of the nft module to pass, but they
	// fail if we return errors for unregistered NFTs on asset. Also the original nft module
	// does not have access to the asset module to register the NFTs
	if types.ErrClassNotFound.Is(err) {
		return nil
	}
	if err != nil {
		return err
	}

	if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, nftID)
	}

	// always allow issuer to receive NFTs issued by them.
	if classDefinition.IsIssuer(receiver) {
		return nil
	}

	whitelisted, err := k.IsWhitelisted(ctx, classID, nftID, receiver)
	if err != nil {
		if errors.Is(err, types.ErrFeatureDisabled) {
			return nil
		}
		return err
	}
	if !whitelisted {
		return sdkerrors.Wrapf(
			cosmoserrors.ErrUnauthorized,
			"nft with classID:%s and ID:%s is not whitelisted for account %s",
			classID, nftID, receiver,
		)
	}
	return nil
}

func (k Keeper) freezeOrUnfreeze(ctx sdk.Context, sender sdk.AccAddress, classID, nftID string, setFrozen bool) error {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return err
	}

	if err = classDefinition.CheckFeatureAllowed(sender, types.ClassFeature_freezing); err != nil {
		return err
	}

	if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, nftID)
	}

	if err := k.SetFrozen(ctx, classID, nftID, setFrozen); err != nil {
		return err
	}

	owner := k.nftKeeper.GetOwner(ctx, classID, nftID)

	var event proto.Message
	if setFrozen {
		event = &types.EventFrozen{
			ClassId: classID,
			Id:      nftID,
			Owner:   owner.String(),
		}
	} else {
		event = &types.EventUnfrozen{
			ClassId: classID,
			Id:      nftID,
			Owner:   owner.String(),
		}
	}

	if err = ctx.EventManager().EmitTypedEvent(event); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit event: %v, err: %s", event, err)
	}

	return nil
}

func (k Keeper) classFreezeOrUnfreeze(
	ctx sdk.Context, sender, account sdk.AccAddress, classID string, setFrozen bool,
) error {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return err
	}

	if err = classDefinition.CheckFeatureAllowed(sender, types.ClassFeature_freezing); err != nil {
		return err
	}

	if classDefinition.Issuer == account.String() {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "setting class-freezing for the nft class issuer is forbidden")
	}

	if !k.nftKeeper.HasClass(ctx, classID) {
		return sdkerrors.Wrapf(types.ErrClassNotFound, "classID:%s not found", classID)
	}

	if err := k.SetClassFrozen(ctx, classID, account, setFrozen); err != nil {
		return err
	}

	var event proto.Message
	if setFrozen {
		event = &types.EventClassFrozen{
			ClassId: classID,
			Account: account.String(),
		}
	} else {
		event = &types.EventClassUnfrozen{
			ClassId: classID,
			Account: account.String(),
		}
	}

	if err = ctx.EventManager().EmitTypedEvent(event); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit event: %v, err: %s", event, err)
	}

	return nil
}

func (k Keeper) addToWhitelistOrRemoveFromWhitelistClass(
	ctx sdk.Context, classID string, sender, account sdk.AccAddress, setWhitelisted bool,
) error {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return err
	}

	if err = classDefinition.CheckFeatureAllowed(sender, types.ClassFeature_whitelisting); err != nil {
		return err
	}

	if classDefinition.Issuer == account.String() {
		return sdkerrors.Wrap(
			cosmoserrors.ErrUnauthorized, "setting class whitelisting for the nft class issuer is forbidden",
		)
	}

	if err := k.SetClassWhitelisting(ctx, classID, account, setWhitelisted); err != nil {
		return err
	}

	var event proto.Message
	if setWhitelisted {
		event = &types.EventAddedToClassWhitelist{
			ClassId: classID,
			Account: account.String(),
		}
	} else {
		event = &types.EventRemovedFromClassWhitelist{
			ClassId: classID,
			Account: account.String(),
		}
	}

	if err = ctx.EventManager().EmitTypedEvent(event); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit event: %v, err: %s", event, err)
	}

	return nil
}

func (k Keeper) addToWhitelistOrRemoveFromWhitelist(
	ctx sdk.Context, classID, nftID string, sender, account sdk.AccAddress, setWhitelisted bool,
) error {
	classDefinition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return err
	}

	if err = classDefinition.CheckFeatureAllowed(sender, types.ClassFeature_whitelisting); err != nil {
		return err
	}

	if !k.nftKeeper.HasNFT(ctx, classID, nftID) {
		return sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, nftID)
	}

	if classDefinition.Issuer == account.String() {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "setting whitelisting for the nft class issuer is forbidden")
	}

	if err := k.SetWhitelisting(ctx, classID, nftID, account, setWhitelisted); err != nil {
		return err
	}

	var event proto.Message
	if setWhitelisted {
		event = &types.EventAddedToWhitelist{
			ClassId: classID,
			Id:      nftID,
			Account: account.String(),
		}
	} else {
		event = &types.EventRemovedFromWhitelist{
			ClassId: classID,
			Id:      nftID,
			Account: account.String(),
		}
	}

	if err = ctx.EventManager().EmitTypedEvent(event); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit event: %v, err: %s", event, err)
	}

	return nil
}
