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

	k.SetClassDefinition(ctx, types.NFTClassDefinition{
		ID:       id,
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

	if err := validateMintingAllowed(settings.Sender, settings.ClassID); err != nil {
		return err
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

	err = checkFeatureAllowed(classID, definition, types.ClassFeature_burn) //nolint:nosnakecase // generated variable
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

// SetClassDefinition stores the NFTClassDefinition.
func (k Keeper) SetClassDefinition(ctx sdk.Context, definition types.NFTClassDefinition) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetClassKey(definition.ID), k.cdc.MustMarshal(&definition))
}

// GetNFTClass reruns the NFTClass.
func (k Keeper) GetNFTClass(ctx sdk.Context, classID string) (types.NFTClass, error) {
	definition, err := k.GetClassDefinition(ctx, classID)
	if err != nil {
		return types.NFTClass{}, err
	}

	class, _ := k.nftKeeper.GetClass(ctx, classID)

	return types.NFTClass{
		Id:          class.Id,
		Name:        class.Name,
		Symbol:      class.Symbol,
		Description: class.Description,
		URI:         class.Uri,
		URIHash:     class.UriHash,
		Data:        class.Data,
		Features:    definition.Features,
	}, nil
}

// GetClassDefinition reruns the NFTClassDefinition.
func (k Keeper) GetClassDefinition(ctx sdk.Context, classID string) (types.NFTClassDefinition, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetClassKey(classID))
	if bz == nil {
		return types.NFTClassDefinition{}, sdkerrors.Wrapf(types.ErrClassNotFound, "classID: %s", classID)
	}
	var definition types.NFTClassDefinition
	k.cdc.MustUnmarshal(bz, &definition)

	return definition, nil
}

// GetClassDefinitions returns all non-fungible class token definitions.
func (k Keeper) GetClassDefinitions(ctx sdk.Context, pagination *query.PageRequest) ([]types.NFTClassDefinition, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.NFTClassKeyPrefix)
	definitionsPointers, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
		pagination,
		// builder
		func(key []byte, definition *types.NFTClassDefinition) (*types.NFTClassDefinition, error) {
			return definition, nil
		},
		// constructor
		func() *types.NFTClassDefinition {
			return &types.NFTClassDefinition{}
		},
	)
	if err != nil {
		return nil, nil, err
	}

	definitions := make([]types.NFTClassDefinition, 0, len(definitionsPointers))
	for _, definition := range definitionsPointers {
		definitions = append(definitions, *definition)
	}

	return definitions, pageRes, err
}

func validateMintingAllowed(sender sdk.AccAddress, classID string) error {
	isIssuer, err := isIssuer(sender, classID)
	if err != nil {
		return err
	}

	if !isIssuer {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "address %q is unauthorized to perform the mint operation", sender.String())
	}

	return nil
}

func isIssuer(sender sdk.AccAddress, classID string) (bool, error) {
	issuer, err := types.DeconstructClassID(classID)
	if err != nil {
		return false, err
	}

	return issuer.String() == sender.String(), nil
}

func checkFeatureAllowed(classID string, definition types.NFTClassDefinition, feature types.ClassFeature) error {
	if !definition.IsFeatureEnabled(feature) {
		return sdkerrors.Wrapf(types.ErrFeatureNotActive, "classID:%s, feature:%s", classID, feature)
	}

	return nil
}
