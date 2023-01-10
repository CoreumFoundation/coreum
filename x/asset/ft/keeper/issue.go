package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// Issue issues new fungible token and returns it's denom.
func (k Keeper) Issue(ctx sdk.Context, settings types.IssueSettings) (string, error) {
	if err := types.ValidateSubunit(settings.Subunit); err != nil {
		return "", sdkerrors.Wrapf(err, "provided subunit: %s", settings.Subunit)
	}

	if err := types.ValidateBurnRate(settings.BurnRate); err != nil {
		return "", err
	}
	if err := types.ValidateSendCommissionRate(settings.SendCommissionRate); err != nil {
		return "", err
	}

	err := types.ValidateSymbol(settings.Symbol)
	if err != nil {
		return "", sdkerrors.Wrapf(err, "provided symbol: %s", settings.Symbol)
	}

	denom := types.BuildDenom(settings.Subunit, settings.Issuer)
	if _, found := k.bankKeeper.GetDenomMetaData(ctx, denom); found {
		return "", sdkerrors.Wrapf(
			types.ErrInvalidInput,
			"subunit %s already registered for the address %s",
			settings.Subunit,
			settings.Issuer.String(),
		)
	}

	params := k.GetParams(ctx)
	if params.IssueFee.IsPositive() {
		coinsToBurn := sdk.NewCoins(params.IssueFee)
		if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, settings.Issuer, types.ModuleName, coinsToBurn); err != nil {
			return "", sdkerrors.Wrapf(err, "can't send coins from account %s to module %s", settings.Issuer.String(), types.ModuleName)
		}
		if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coinsToBurn); err != nil {
			return "", sdkerrors.Wrapf(err, "can't burn %s for the module %s", coinsToBurn.String(), types.ModuleName)
		}
	}

	if err := k.SetSymbol(ctx, settings.Symbol, settings.Issuer); err != nil {
		return "", sdkerrors.Wrapf(err, "provided symbol: %s", settings.Symbol)
	}

	definition := types.FTDefinition{
		Denom:              denom,
		Issuer:             settings.Issuer.String(),
		Features:           settings.Features,
		BurnRate:           settings.BurnRate,
		SendCommissionRate: settings.SendCommissionRate,
	}

	k.SetDenomMetadata(ctx, denom, settings.Symbol, settings.Description, settings.Precision)
	k.SetTokenDefinition(ctx, settings.Issuer, settings.Subunit, definition)

	if err := k.mint(ctx, definition, settings.InitialAmount, settings.Issuer); err != nil {
		return "", err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventTokenIssued{
		Denom:              denom,
		Issuer:             settings.Issuer.String(),
		Symbol:             settings.Symbol,
		Subunit:            settings.Subunit,
		Precision:          settings.Precision,
		Description:        settings.Description,
		InitialAmount:      settings.InitialAmount,
		Features:           settings.Features,
		BurnRate:           settings.BurnRate,
		SendCommissionRate: settings.SendCommissionRate,
	}); err != nil {
		return "", sdkerrors.Wrap(err, "can't emit EventTokenIssued event")
	}

	k.Logger(ctx).Debug("issued new fungible token with denom %d", denom)

	return denom, nil
}

// IsSymbolDuplicate checks symbol exists in the store
func (k Keeper) IsSymbolDuplicate(ctx sdk.Context, symbol string, issuer sdk.AccAddress) bool {
	symbol = types.NormalizeSymbolForKey(symbol)
	compositeKey := types.CreateSymbolKey(issuer, symbol)
	bytes := ctx.KVStore(k.storeKey).Get(compositeKey)
	return bytes != nil
}

// SetSymbol saves the symbol to store
func (k Keeper) SetSymbol(ctx sdk.Context, symbol string, issuer sdk.AccAddress) error {
	if k.IsSymbolDuplicate(ctx, symbol, issuer) {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "duplicate symbol %s", symbol)
	}

	compositeKey := types.CreateSymbolKey(issuer, symbol)
	ctx.KVStore(k.storeKey).Set(compositeKey, []byte{0x01})
	return nil
}

// GetToken return the fungible token by its denom.
func (k Keeper) GetToken(ctx sdk.Context, denom string) (types.FT, error) {
	definition, err := k.GetTokenDefinition(ctx, denom)
	if err != nil {
		return types.FT{}, err
	}

	return k.getTokenFullInfo(ctx, definition)
}

// getTokenFullInfo return the fungible token info from bank, given its definition.
func (k Keeper) getTokenFullInfo(ctx sdk.Context, definition types.FTDefinition) (types.FT, error) {
	subunit, _, err := types.DeconstructDenom(definition.Denom)
	if err != nil {
		return types.FT{}, err
	}

	metadata, found := k.bankKeeper.GetDenomMetaData(ctx, definition.Denom)
	if !found {
		return types.FT{}, sdkerrors.Wrapf(types.ErrFTNotFound, "metadata for %s denom not found", definition.Denom)
	}

	precision := -1
	for _, unit := range metadata.DenomUnits {
		if unit.Denom == metadata.Symbol {
			precision = int(unit.Exponent)
			break
		}
	}

	if precision < 0 {
		return types.FT{}, sdkerrors.Wrap(types.ErrInvalidInput, "precision not found")
	}

	return types.FT{
		Denom:              definition.Denom,
		Issuer:             definition.Issuer,
		Symbol:             metadata.Symbol,
		Precision:          uint32(precision),
		Subunit:            subunit,
		Description:        metadata.Description,
		Features:           definition.Features,
		BurnRate:           definition.BurnRate,
		SendCommissionRate: definition.SendCommissionRate,
		GloballyFrozen:     k.isGloballyFrozen(ctx, definition.Denom),
	}, nil
}

// GetTokens returns all fungible tokens.
func (k Keeper) GetTokens(ctx sdk.Context, pagination *query.PageRequest) ([]types.FT, *query.PageResponse, error) {
	definitions, pageResponse, err := k.getTokenDefinitions(ctx, pagination)
	if err != nil {
		return nil, nil, err
	}

	tokens, err := k.getTokensByDefinitions(ctx, definitions)
	if err != nil {
		return nil, nil, err
	}

	return tokens, pageResponse, nil
}

// GetIssuerTokens returns fungible tokens issued by the issuer.
func (k Keeper) GetIssuerTokens(ctx sdk.Context, issuer sdk.AccAddress, pagination *query.PageRequest) ([]types.FT, *query.PageResponse, error) {
	definitions, pageResponse, err := k.getIssuerTokenDefinitions(ctx, issuer, pagination)
	if err != nil {
		return nil, nil, err
	}

	tokens, err := k.getTokensByDefinitions(ctx, definitions)
	if err != nil {
		return nil, nil, err
	}

	return tokens, pageResponse, nil
}

// GetTokenDefinition returns the TokenDefinition by the denom.
func (k Keeper) GetTokenDefinition(ctx sdk.Context, denom string) (types.FTDefinition, error) {
	subunit, issuer, err := types.DeconstructDenom(denom)
	if err != nil {
		return types.FTDefinition{}, err
	}
	bz := ctx.KVStore(k.storeKey).Get(types.CreateTokenKey(issuer, subunit))
	if bz == nil {
		return types.FTDefinition{}, sdkerrors.Wrapf(types.ErrFTNotFound, "denom: %s", denom)
	}
	var definition types.FTDefinition
	k.cdc.MustUnmarshal(bz, &definition)

	return definition, nil
}

// IterateAllTokenDefinitions iterates over all token definitions applies the provided callback.
// If true is returned from the callback, iteration is halted.
func (k Keeper) IterateAllTokenDefinitions(ctx sdk.Context, cb func(types.FTDefinition) bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.FTKeyPrefix)
	iterator := store.Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var definition types.FTDefinition
		k.cdc.MustUnmarshal(iterator.Value(), &definition)

		if cb(definition) {
			break
		}
	}
}

// SetTokenDefinition stores the TokenDefinition.
func (k Keeper) SetTokenDefinition(ctx sdk.Context, issuer sdk.AccAddress, subunit string, definition types.FTDefinition) {
	ctx.KVStore(k.storeKey).Set(types.CreateTokenKey(issuer, subunit), k.cdc.MustMarshal(&definition))
}

// SetDenomMetadata registers denom metadata on the bank keeper
func (k Keeper) SetDenomMetadata(ctx sdk.Context, denom, symbol, description string, precision uint32) {
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

func (k Keeper) getTokenDefinitions(ctx sdk.Context, pagination *query.PageRequest) ([]types.FTDefinition, *query.PageResponse, error) {
	return k.getTokenDefinitionsFromStore(prefix.NewStore(ctx.KVStore(k.storeKey), types.FTKeyPrefix), pagination)
}

func (k Keeper) getIssuerTokenDefinitions(ctx sdk.Context, issuer sdk.AccAddress, pagination *query.PageRequest) ([]types.FTDefinition, *query.PageResponse, error) {
	return k.getTokenDefinitionsFromStore(prefix.NewStore(ctx.KVStore(k.storeKey), types.CreateIssuerTokensPrefix(issuer)), pagination)
}

// getTokenDefinitionsFromStore queries the FTDefinitions form the provided store.
// It is used to either get all definitions or only those which starts from particular prefix.
func (k Keeper) getTokenDefinitionsFromStore(store prefix.Store, pagination *query.PageRequest) ([]types.FTDefinition, *query.PageResponse, error) {
	definitionsPointers, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
		pagination,
		// builder
		func(key []byte, definition *types.FTDefinition) (*types.FTDefinition, error) {
			return definition, nil
		},
		// constructor
		func() *types.FTDefinition {
			return &types.FTDefinition{}
		},
	)
	if err != nil {
		return nil, nil, err
	}

	definitions := make([]types.FTDefinition, 0, len(definitionsPointers))
	for _, definition := range definitionsPointers {
		definitions = append(definitions, *definition)
	}

	return definitions, pageRes, err
}

func (k Keeper) getTokensByDefinitions(ctx sdk.Context, definitions []types.FTDefinition) ([]types.FT, error) {
	var tokens []types.FT
	for _, definition := range definitions {
		token, err := k.getTokenFullInfo(ctx, definition)
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
	}
	return tokens, nil
}

func (k Keeper) mint(ctx sdk.Context, ft types.FTDefinition, amount sdk.Int, recipient sdk.AccAddress) error {
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

func (k Keeper) burn(ctx sdk.Context, account sdk.AccAddress, ft types.FTDefinition, amount sdk.Int) error {
	if err := k.isCoinSpendable(ctx, account, ft, amount); err != nil {
		return sdkerrors.Wrapf(err, "coins are not spendable")
	}

	coinsToBurn := sdk.NewCoins(sdk.NewCoin(ft.Denom, amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, coinsToBurn); err != nil {
		return sdkerrors.Wrapf(err, "can't send coins from account %s to module %s", account.String(), types.ModuleName)
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coinsToBurn); err != nil {
		return sdkerrors.Wrapf(err, "can't burn %s for the module %s", coinsToBurn.String(), types.ModuleName)
	}

	return nil
}
