package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/x/nft"
)

// Keeper is the asset module non-fungible token nftKeeper.
type Keeper struct {
	cdc       codec.BinaryCodec
	storeKey  sdk.StoreKey
	nftKeeper types.NFTKeeper
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, nftKeeper types.NFTKeeper) Keeper {
	return Keeper{
		cdc:       cdc,
		storeKey:  storeKey,
		nftKeeper: nftKeeper,
	}
}

// IssueClass issues new non-fungible token class and returns its id.
func (k Keeper) IssueClass(ctx sdk.Context, settings types.IssueClassSettings) (string, error) {
	if err := types.ValidateClassSymbol(settings.Symbol); err != nil {
		return "", err
	}

	id := types.BuildClassID(settings.Symbol, settings.Issuer)
	if err := nft.ValidateClassID(id); err != nil {
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

	k.SetClassDefinition(ctx, types.ClassDefinition{
		ID:       id,
		Issuer:   settings.Issuer.String(),
		Features: settings.Features,
	})

	if err := ctx.EventManager().EmitTypedEvent(&types.EventClassIssued{
		ID:          id,
		Issuer:      settings.Issuer.String(),
		Symbol:      settings.Symbol,
		Name:        settings.Name,
		Description: settings.Description,
		URI:         settings.URI,
		URIHash:     settings.URIHash,
		Features:    settings.Features,
	}); err != nil {
		return "", sdkerrors.Wrapf(types.ErrInvalidInput, "can't emit event EventClassIssued: %s", err)
	}

	return id, nil
}

// Mint mints new non-fungible token.
func (k Keeper) Mint(ctx sdk.Context, settings types.MintSettings) error {
	if err := types.ValidateTokenID(settings.ID); err != nil {
		return sdkerrors.Wrap(types.ErrInvalidInput, err.Error())
	}

	definition, err := k.GetClassDefinition(ctx, settings.ClassID)
	if err != nil {
		return err
	}

	if !definition.IsIssuer(settings.Sender) {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "address %q is unauthorized to perform the mint operation", settings.Sender.String())
	}

	if !k.nftKeeper.HasClass(ctx, settings.ClassID) {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "classID %q not found", settings.ClassID)
	}

	if nftFound := k.nftKeeper.HasNFT(ctx, settings.ClassID, settings.ID); nftFound {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "ID %q already defined for the class", settings.ID)
	}

	if err := k.nftKeeper.Mint(ctx, nft.NFT{
		ClassId: settings.ClassID,
		Id:      settings.ID,
		Uri:     settings.URI,
		UriHash: settings.URIHash,
		Data:    settings.Data,
	}, settings.Sender); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "can't save non-fungible token: %s", err)
	}

	return nil
}

// Burn burns non-fungible token.
func (k Keeper) Burn(ctx sdk.Context, owner sdk.AccAddress, classID, id string) error {
	definition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return err
	}

	err = checkFeatureAllowed(owner, definition, types.ClassFeature_burn) //nolint:nosnakecase // generated variable
	if err != nil {
		return err
	}

	if !k.nftKeeper.HasNFT(ctx, classID, id) {
		return sdkerrors.Wrapf(types.ErrNFTNotFound, "nft with classID:%s and ID:%s not found", classID, id)
	}

	if k.nftKeeper.GetOwner(ctx, classID, id).String() != owner.String() {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "only owner can burn the nft")
	}

	return k.nftKeeper.Burn(ctx, classID, id)
}

// SetClassDefinition stores the ClassDefinition.
func (k Keeper) SetClassDefinition(ctx sdk.Context, definition types.ClassDefinition) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetClassKey(definition.ID), k.cdc.MustMarshal(&definition))
}

// GetClass reruns the Class.
func (k Keeper) GetClass(ctx sdk.Context, classID string) (types.Class, error) {
	definition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return types.Class{}, err
	}

	class, found := k.nftKeeper.GetClass(ctx, classID)
	if !found {
		return types.Class{}, sdkerrors.Wrapf(types.ErrNFTNotFound, "nft class with ID:%s not found", classID)
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
	}, nil
}

// GetClassDefinition reruns the ClassDefinition.
func (k Keeper) GetClassDefinition(ctx sdk.Context, classID string) (types.ClassDefinition, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetClassKey(classID))
	if bz == nil {
		return types.ClassDefinition{}, sdkerrors.Wrapf(types.ErrClassNotFound, "classID: %s", classID)
	}
	var definition types.ClassDefinition
	k.cdc.MustUnmarshal(bz, &definition)

	return definition, nil
}

// GetClassDefinitions returns all non-fungible class token definitions.
func (k Keeper) GetClassDefinitions(ctx sdk.Context, pagination *query.PageRequest) ([]types.ClassDefinition, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.NFTClassKeyPrefix)
	definitionsPointers, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
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
		return nil, nil, err
	}

	definitions := make([]types.ClassDefinition, 0, len(definitionsPointers))
	for _, definition := range definitionsPointers {
		definitions = append(definitions, *definition)
	}

	return definitions, pageRes, err
}

//nolint:nosnakecase
func checkFeatureAllowed(addr sdk.AccAddress, def types.ClassDefinition, feature types.ClassFeature) error {
	featureEnabled := def.IsFeatureEnabled(feature)

	// issuer can use any enabled feature and burning even if it is disabled
	if def.IsIssuer(addr) {
		if featureEnabled || feature == types.ClassFeature_burn {
			return nil
		}
	}

	// non-issuer can use only burning and only if it is enabled
	if featureEnabled && feature == types.ClassFeature_burn {
		return nil
	}

	return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "address %s is unauthorized to perform %q related operations", addr.String(), feature.String())
}
