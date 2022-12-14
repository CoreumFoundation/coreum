package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum/pkg/store"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// IssueFungibleToken issues new fungible token and returns it's denom.
func (k Keeper) IssueFungibleToken(ctx sdk.Context, settings types.IssueFungibleTokenSettings) (string, error) {
	if err := types.ValidateSubunit(settings.Subunit); err != nil {
		return "", sdkerrors.Wrapf(err, "provided subunit: %s", settings.Subunit)
	}

	if err := types.ValidateBurnRate(settings.BurnRate); err != nil {
		return "", err
	}

	err := types.ValidateSymbol(settings.Symbol)
	if err != nil {
		return "", sdkerrors.Wrapf(err, "provided symbol: %s", settings.Symbol)
	}

	if err := k.StoreSymbol(ctx, settings.Symbol, settings.Issuer); err != nil {
		return "", sdkerrors.Wrapf(err, "provided symbol: %s", settings.Symbol)
	}

	denom := types.BuildFungibleTokenDenom(settings.Subunit, settings.Issuer)
	if _, found := k.bankKeeper.GetDenomMetaData(ctx, denom); found {
		return "", sdkerrors.Wrapf(
			types.ErrInvalidInput,
			"subunit %s already registered for the address %s",
			settings.Subunit,
			settings.Issuer.String(),
		)
	}

	k.SetFungibleTokenDenomMetadata(ctx, denom, settings.Symbol, settings.Description, settings.Precision)

	definition := types.FungibleTokenDefinition{
		Denom:    denom,
		Issuer:   settings.Issuer.String(),
		Features: settings.Features,
		BurnRate: settings.BurnRate,
	}
	k.SetFungibleTokenDefinition(ctx, definition)

	// TODO: Delete this once recipient is removed
	//nolint:nosnakecase
	if definition.IsFeatureEnabled(types.FungibleTokenFeature_whitelist) && settings.InitialAmount.IsPositive() {
		if err := k.SetWhitelistedBalance(ctx, settings.Issuer, settings.Recipient, sdk.NewCoin(denom, settings.InitialAmount)); err != nil {
			return "", err
		}
	}

	if err := k.mintFungibleToken(ctx, definition, settings.InitialAmount, settings.Recipient); err != nil {
		return "", err
	}

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
		BurnRate:      settings.BurnRate,
	}); err != nil {
		return "", sdkerrors.Wrap(err, "can't emit EventFungibleTokenIssued event")
	}

	k.Logger(ctx).Debug("issued new fungible token with denom %d", denom)

	return denom, nil
}

// IsSymbolDuplicate checks symbol exists in the store
func (k Keeper) IsSymbolDuplicate(ctx sdk.Context, symbol string, issuer sdk.AccAddress) bool {
	symbol = types.NormalizeSymbolForKey(symbol)
	compositeKey := store.JoinKeys(types.CreateSymbolPrefix(issuer), []byte(symbol))
	bytes := ctx.KVStore(k.storeKey).Get(compositeKey)
	return bytes != nil
}

// StoreSymbol saves the symbol to store
func (k Keeper) StoreSymbol(ctx sdk.Context, symbol string, issuer sdk.AccAddress) error {
	if k.IsSymbolDuplicate(ctx, symbol, issuer) {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "duplicate symbol %s", symbol)
	}

	compositeKey := store.JoinKeys(types.CreateSymbolPrefix(issuer), []byte(symbol))
	ctx.KVStore(k.storeKey).Set(compositeKey, []byte{0x01})
	return nil
}

// GetFungibleToken return the fungible token by its denom.
func (k Keeper) GetFungibleToken(ctx sdk.Context, denom string) (types.FungibleToken, error) {
	definition, err := k.GetFungibleTokenDefinition(ctx, denom)
	if err != nil {
		return types.FungibleToken{}, err
	}

	return k.getFungibleTokenFullInfo(ctx, definition)
}

// getFungibleTokenFullInfo return the fungible token info from bank, given its definition.
func (k Keeper) getFungibleTokenFullInfo(ctx sdk.Context, definition types.FungibleTokenDefinition) (types.FungibleToken, error) {
	subunit, _, err := types.DeconstructFungibleTokenDenom(definition.Denom)
	if err != nil {
		return types.FungibleToken{}, err
	}

	metadata, found := k.bankKeeper.GetDenomMetaData(ctx, definition.Denom)
	if !found {
		return types.FungibleToken{}, sdkerrors.Wrapf(types.ErrFungibleTokenNotFound, "metadata for %s denom not found", definition.Denom)
	}

	precision := -1
	for _, unit := range metadata.DenomUnits {
		if unit.Denom == metadata.Symbol {
			precision = int(unit.Exponent)
			break
		}
	}

	if precision < 0 {
		return types.FungibleToken{}, sdkerrors.Wrap(types.ErrInvalidInput, "precision not found")
	}

	return types.FungibleToken{
		Denom:          definition.Denom,
		Issuer:         definition.Issuer,
		Symbol:         metadata.Symbol,
		Precision:      uint32(precision),
		Subunit:        subunit,
		Description:    metadata.Description,
		Features:       definition.Features,
		BurnRate:       definition.BurnRate,
		GloballyFrozen: k.isGloballyFrozen(ctx, definition.Denom),
	}, nil
}

// GetFungibleTokens returns all fungible tokens.
func (k Keeper) GetFungibleTokens(ctx sdk.Context, pagination *query.PageRequest) ([]types.FungibleToken, *query.PageResponse, error) {
	definitions, pageResponse, err := k.GetFungibleTokenDefinitions(ctx, pagination)
	if err != nil {
		return nil, nil, err
	}

	var tokens []types.FungibleToken
	for _, definition := range definitions {
		token, err := k.getFungibleTokenFullInfo(ctx, definition)
		if err != nil {
			return nil, nil, err
		}

		tokens = append(tokens, token)
	}

	return tokens, pageResponse, nil
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

// SetFungibleTokenDenomMetadata registers denom metadata on the bank keeper
func (k Keeper) SetFungibleTokenDenomMetadata(ctx sdk.Context, denom, symbol, description string, precision uint32) {
	denomMetadata := banktypes.Metadata{
		Name:        symbol,
		Symbol:      symbol,
		Description: description,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    symbol,
				Exponent: precision,
			},
			{
				Denom:    denom,
				Exponent: uint32(0),
			},
		},
		// here take subunit provided by the user, generate the denom and used it as base,
		// and we take the symbol provided by the user and use it as symbol
		Base:    denom,
		Display: symbol,
	}

	k.bankKeeper.SetDenomMetaData(ctx, denomMetadata)
}

func (k Keeper) mintFungibleToken(ctx sdk.Context, ft types.FungibleTokenDefinition, amount sdk.Int, recipient sdk.AccAddress) error {
	if !amount.IsPositive() {
		return nil
	}
	if err := k.isCoinReceivable(ctx, recipient, ft, amount); err != nil {
		return sdkerrors.Wrapf(err, "coins are not receivable")
	}

	coinsToMint := sdk.NewCoins(sdk.NewCoin(ft.Denom, amount))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coinsToMint); err != nil {
		return sdkerrors.Wrapf(err, "can't mint %s for the module %s", coinsToMint.String(), types.ModuleName)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coinsToMint); err != nil {
		return sdkerrors.Wrapf(err, "can't send minted coins from module %s to account %s", types.ModuleName, recipient.String())
	}

	return nil
}

func (k Keeper) burnFungibleToken(ctx sdk.Context, account sdk.AccAddress, ft types.FungibleTokenDefinition, amount sdk.Int) error {
	if err := k.isCoinSpendable(ctx, account, ft, amount); err != nil {
		return sdkerrors.Wrapf(err, "coins are not spendable")
	}

	coinsToBurn := sdk.NewCoins(sdk.NewCoin(ft.Denom, amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, coinsToBurn); err != nil {
		return sdkerrors.Wrapf(err, "can't send  coins from account %s to module %s", account.String(), types.ModuleName)
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coinsToBurn); err != nil {
		return sdkerrors.Wrapf(err, "can't burn %s for the module %s", coinsToBurn.String(), types.ModuleName)
	}

	return nil
}
