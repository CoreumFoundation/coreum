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
		return "", sdkerrors.Wrapf(err, "provided subunit: %s", settings.Subunit)
	}

	if err := k.checkAndStoreSymbol(ctx, settings.Symbol, settings.Issuer); err != nil {
		return "", sdkerrors.Wrapf(err, "provided symbol: %s", settings.Symbol)
	}

	denom := types.BuildFungibleTokenDenom(settings.Subunit, settings.Issuer)
	if _, found := k.bankKeeper.GetDenomMetaData(ctx, denom); found {
		return "", sdkerrors.Wrapf(
			types.ErrInvalidSubunit,
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
		Subunit:       settings.Subunit,
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

func (k Keeper) checkAndStoreSymbol(ctx sdk.Context, symbol string, issuer sdk.AccAddress) error {
	err := types.ValidateSymbol(symbol)
	if err != nil {
		return sdkerrors.Wrapf(err, "provided symbol: %s", symbol)
	}

	symbol = strings.ToLower(symbol)
	symbolStore := prefix.NewStore(ctx.KVStore(k.storeKey), types.CreateSymbolPrefix(issuer))
	bytes := symbolStore.Get([]byte(symbol))
	if bytes != nil {
		return sdkerrors.Wrapf(types.ErrInvalidSymbol, "duplicate symbol")
	}

	symbolStore.Set([]byte(symbol), []byte{0x01})
	return nil
}

// GetFungibleToken return the fungible token by its denom.
func (k Keeper) GetFungibleToken(ctx sdk.Context, denom string) (types.FungibleToken, error) {
	definition, err := k.GetFungibleTokenDefinition(ctx, denom)
	if err != nil {
		return types.FungibleToken{}, err
	}

	subunit, _, err := types.DeconstructFungibleTokenDenom(definition.Denom)
	if err != nil {
		return types.FungibleToken{}, err
	}

	metadata, found := k.bankKeeper.GetDenomMetaData(ctx, denom)
	if !found {
		return types.FungibleToken{}, sdkerrors.Wrapf(types.ErrFungibleTokenNotFound, "metadata for %s denom not found", denom)
	}

	var precision uint32
	precisionFound := false
	for _, unit := range metadata.DenomUnits {
		if unit.Denom == metadata.Symbol {
			precision = unit.Exponent
			precisionFound = true
			break
		}
	}

	if !precisionFound {
		return types.FungibleToken{}, sdkerrors.Wrap(types.ErrInvalidFungibleToken, "precision not found")
	}

	return types.FungibleToken{
		Denom:       definition.Denom,
		Issuer:      definition.Issuer,
		Symbol:      metadata.Symbol,
		Precision:   precision,
		Subunit:     subunit,
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

func (k Keeper) burnFungibleToken(ctx sdk.Context, coin sdk.Coin, account sdk.AccAddress) error {
	if err := coin.Validate(); err != nil {
		return err
	}

	coinsToBurn := sdk.NewCoins(coin)
	if err := k.areCoinsSpendable(ctx, account, coinsToBurn); err != nil {
		return sdkerrors.Wrapf(err, "coins are not spendable")
	}

	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, coinsToBurn); err != nil {
		return sdkerrors.Wrapf(err, "can't send  coins from account %s to module %s", account.String(), types.ModuleName)
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coinsToBurn); err != nil {
		return sdkerrors.Wrapf(err, "can't burn %s for the module %s", coinsToBurn.String(), types.ModuleName)
	}

	return nil
}
