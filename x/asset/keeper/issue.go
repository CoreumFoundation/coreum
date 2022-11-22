package keeper

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// IssueFungibleToken issues new fungible token and returns it's denom.
func (k Keeper) IssueFungibleToken(ctx sdk.Context, settings types.IssueFungibleTokenSettings) (string, error) {
	if err := types.ValidateSubunit(settings.Subunit); err != nil {
		return "", err
	}

	denom := types.BuildFungibleTokenDenom(settings.Subunit, settings.Issuer)
	if _, found := k.bankKeeper.GetDenomMetaData(ctx, denom); found {
		return "", sdkerrors.Wrapf(
			types.ErrInvalidFungibleToken,
			"subunit %s already registered for the address %s",
			settings.Subunit,
			settings.Issuer.String(),
		)
	}

	k.setFungibleTokenDenomMetadata(ctx, denom, settings)

	if err := k.mintFungibleToken(ctx, denom, settings.InitialAmount, settings.Recipient); err != nil {
		return "", err
	}

	definition := types.FungibleTokenDefinition{
		Denom:    denom,
		Issuer:   settings.Issuer.String(),
		Features: settings.Features,
	}
	k.SetFungibleTokenDefinition(ctx, definition)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventFungibleTokenIssued{
		Denom:         denom,
		Issuer:        settings.Issuer.String(),
		Symbol:        settings.Symbol,
		SubUnit:       settings.Subunit,
		Precision:     settings.Precision,
		Description:   settings.Description,
		Recipient:     settings.Recipient.String(),
		InitialAmount: settings.InitialAmount,
		Features:      settings.Features,
	}); err != nil {
		return "", sdkerrors.Wrap(err, "can't emit EventFungibleTokenIssued event")
	}

	k.Logger(ctx).Debug("issued new fungible token with denom %d", denom)

	return denom, nil
}

// GetFungibleToken return the fungible token by its denom.
func (k Keeper) GetFungibleToken(ctx sdk.Context, denom string) (types.FungibleToken, error) {
	definition, err := k.GetFungibleTokenDefinition(ctx, denom)
	if err != nil {
		return types.FungibleToken{}, err
	}

	denomParts := strings.Split(definition.Denom, "-")
	if len(denomParts) != 2 {
		return types.FungibleToken{}, sdkerrors.Wrap(types.ErrInvalidSubunit, "symbol must match format [subunit]-[issuer-address]")
	}

	metadata, found := k.bankKeeper.GetDenomMetaData(ctx, denom)
	if !found {
		return types.FungibleToken{}, sdkerrors.Wrapf(types.ErrFungibleTokenNotFound, "metadata for %s denom not found", denom)
	}

	var precision uint32
	for _, unit := range metadata.DenomUnits {
		if unit.Denom == metadata.Symbol {
			precision = unit.Exponent
		}
	}

	return types.FungibleToken{
		Denom:       definition.Denom,
		Issuer:      definition.Issuer,
		Symbol:      metadata.Symbol,
		Precision:   precision,
		SubUnit:     denomParts[0],
		Description: metadata.Description,
		Features:    definition.Features,
	}, nil
}

// GetFungibleTokenDefinitions returns all fungible token definitions.
func (k Keeper) GetFungibleTokenDefinitions(ctx sdk.Context, pagination *query.PageRequest) ([]types.FungibleTokenDefinition, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.FungibleTokenKeyPrefix)
	definitionsPointers, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
		pagination,
		// builder
		func(key []byte, definition *types.FungibleTokenDefinition) (*types.FungibleTokenDefinition, error) {
			return definition, nil
		},
		// constructor
		func() *types.FungibleTokenDefinition {
			return &types.FungibleTokenDefinition{}
		},
	)

	if err != nil {
		return nil, nil, err
	}

	definitions := make([]types.FungibleTokenDefinition, 0, len(definitionsPointers))
	for _, definition := range definitionsPointers {
		definitions = append(definitions, *definition)
	}

	return definitions, pageRes, err
}

// GetFungibleTokenDefinition returns the FungibleTokenDefinition by the denom.
func (k Keeper) GetFungibleTokenDefinition(ctx sdk.Context, denom string) (types.FungibleTokenDefinition, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetFungibleTokenKey(denom))
	if bz == nil {
		return types.FungibleTokenDefinition{}, sdkerrors.Wrapf(types.ErrFungibleTokenNotFound, "denom: %s", denom)
	}
	var definition types.FungibleTokenDefinition
	k.cdc.MustUnmarshal(bz, &definition)

	return definition, nil
}

// SetFungibleTokenDefinition stores the FungibleTokenDefinition.
func (k Keeper) SetFungibleTokenDefinition(ctx sdk.Context, definition types.FungibleTokenDefinition) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetFungibleTokenKey(definition.Denom), k.cdc.MustMarshal(&definition))
}

func (k Keeper) setFungibleTokenDenomMetadata(ctx sdk.Context, denom string, st types.IssueFungibleTokenSettings) {
	denomMetadata := banktypes.Metadata{
		Name:        st.Symbol,
		Symbol:      st.Symbol,
		Description: st.Description,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    st.Symbol,
				Exponent: st.Precision,
			},
			{
				Denom:    denom,
				Exponent: uint32(0),
			},
		},
		Base:    denom,
		Display: st.Symbol,
	}

	k.bankKeeper.SetDenomMetaData(ctx, denomMetadata)
}

func (k Keeper) mintFungibleToken(ctx sdk.Context, denom string, amount sdk.Int, recipient sdk.AccAddress) error {
	if !amount.IsPositive() {
		return nil
	}
	coinsToMint := sdk.NewCoins(sdk.NewCoin(denom, amount))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coinsToMint); err != nil {
		return sdkerrors.Wrapf(err, "can't mint %s for the module %s", coinsToMint.String(), types.ModuleName)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coinsToMint); err != nil {
		return sdkerrors.Wrapf(err, "can't send minted coins from module %s to account %s", types.ModuleName, recipient.String())
	}

	return nil
}
