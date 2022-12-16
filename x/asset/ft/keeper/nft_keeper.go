package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/CoreumFoundation/coreum/x/asset/types"
	"github.com/CoreumFoundation/coreum/x/nft"
)

// NonFungibleTokenKeeper is the asset module non-fungible token keeper.
type NonFungibleTokenKeeper struct {
	cdc       codec.BinaryCodec
	storeKey  sdk.StoreKey
	nftKeeper types.NFTKeeper
}

// NewNonFungibleTokenKeeper creates a new instance of the NonFungibleTokenKeeper.
func NewNonFungibleTokenKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, nftKeeper types.NFTKeeper) NonFungibleTokenKeeper {
	return NonFungibleTokenKeeper{
		cdc:       cdc,
		storeKey:  storeKey,
		nftKeeper: nftKeeper,
	}
}

// IssueClass issues new non-fungible token class and returns its id.
func (k NonFungibleTokenKeeper) IssueClass(ctx sdk.Context, settings types.IssueNonFungibleTokenClassSettings) (string, error) {
	if err := types.ValidateNonFungibleTokenClassSymbol(settings.Symbol); err != nil {
		return "", err
	}

	id := types.BuildNonFungibleTokenClassID(settings.Symbol, settings.Issuer)
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

	if err := ctx.EventManager().EmitTypedEvent(&types.EventNonFungibleTokenClassIssued{
		ID:          id,
		Issuer:      settings.Issuer.String(),
		Symbol:      settings.Symbol,
		Name:        settings.Name,
		Description: settings.Description,
		URI:         settings.URI,
		URIHash:     settings.URIHash,
	}); err != nil {
		return "", sdkerrors.Wrapf(types.ErrInvalidInput, "can't emit event EventNonFungibleTokenClassIssued: %s", err)
	}

	return id, nil
}

// Mint mints new non-fungible token.
func (k NonFungibleTokenKeeper) Mint(ctx sdk.Context, settings types.MintNonFungibleTokenSettings) error {
	if err := types.ValidateNonFungibleTokenID(settings.ID); err != nil {
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
	issuer, err := types.DeconstructNonFungibleTokenClassID(classID)
	if err != nil {
		return false, err
	}

	return issuer.String() == sender.String(), nil
}
