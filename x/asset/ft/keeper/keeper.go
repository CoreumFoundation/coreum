package keeper

import (
	"bytes"
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// globalFreezeEnabledStoreVal is the value used to store the globally frozen flag.
var globalFreezeEnabledStoreVal = []byte{0x01}

// ParamSubspace represents a subscope of methods exposed by param module to store and retrieve parameters.
type ParamSubspace interface {
	GetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
	SetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
}

// Keeper is the asset module keeper.
type Keeper struct {
	cdc           codec.BinaryCodec
	paramSubspace ParamSubspace
	storeKey      sdk.StoreKey
	bankKeeper    types.BankKeeper
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	paramSubspace ParamSubspace,
	storeKey sdk.StoreKey,
	bankKeeper types.BankKeeper,
) Keeper {
	return Keeper{
		cdc:           cdc,
		paramSubspace: paramSubspace,
		storeKey:      storeKey,
		bankKeeper:    bankKeeper,
	}
}

// GetParams gets the parameters of the model.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var params types.Params
	k.paramSubspace.GetParamSet(ctx, &params)
	return params
}

// SetParams sets the parameters of the model.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

// GetTokens returns all fungible tokens.
func (k Keeper) GetTokens(ctx sdk.Context, pagination *query.PageRequest) ([]types.Token, *query.PageResponse, error) {
	defs, pageResponse, err := k.getDefinitions(ctx, pagination)
	if err != nil {
		return nil, nil, err
	}

	tokens, err := k.getTokensByDefinitions(ctx, defs)
	if err != nil {
		return nil, nil, err
	}

	return tokens, pageResponse, nil
}

// GetIssuerTokens returns fungible tokens issued by the issuer.
func (k Keeper) GetIssuerTokens(ctx sdk.Context, issuer sdk.AccAddress, pagination *query.PageRequest) ([]types.Token, *query.PageResponse, error) {
	definitions, pageResponse, err := k.getIssuerDefinitions(ctx, issuer, pagination)
	if err != nil {
		return nil, nil, err
	}

	tokens, err := k.getTokensByDefinitions(ctx, definitions)
	if err != nil {
		return nil, nil, err
	}

	return tokens, pageResponse, nil
}

// IterateAllDefinitions iterates over all token definitions applies the provided callback.
// If true is returned from the callback, iteration is halted.
func (k Keeper) IterateAllDefinitions(ctx sdk.Context, cb func(types.Definition) bool) {
	iterator := prefix.NewStore(ctx.KVStore(k.storeKey), types.TokenKeyPrefix).Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var definition types.Definition
		k.cdc.MustUnmarshal(iterator.Value(), &definition)

		if cb(definition) {
			break
		}
	}
}

// GetDefinition returns the Definition by the denom.
func (k Keeper) GetDefinition(ctx sdk.Context, denom string) (types.Definition, error) {
	subunit, issuer, err := types.DeconstructDenom(denom)
	if err != nil {
		return types.Definition{}, err
	}
	bz := ctx.KVStore(k.storeKey).Get(types.CreateTokenKey(issuer, subunit))
	if bz == nil {
		return types.Definition{}, sdkerrors.Wrapf(types.ErrTokenNotFound, "denom: %s", denom)
	}
	var definition types.Definition
	k.cdc.MustUnmarshal(bz, &definition)

	return definition, nil
}

// GetToken return the fungible token by its denom.
func (k Keeper) GetToken(ctx sdk.Context, denom string) (types.Token, error) {
	def, err := k.GetDefinition(ctx, denom)
	if err != nil {
		return types.Token{}, err
	}

	return k.getTokenFullInfo(ctx, def)
}

// Issue issues new fungible token and returns it's denom.
func (k Keeper) Issue(ctx sdk.Context, settings types.IssueSettings) (string, error) {
	if err := types.ValidateSubunit(settings.Subunit); err != nil {
		return "", sdkerrors.Wrapf(err, "provided subunit: %s", settings.Subunit)
	}

	if err := types.ValidatePrecision(settings.Precision); err != nil {
		return "", sdkerrors.Wrapf(err, "provided precision: %d", settings.Precision)
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
		if err := k.burn(ctx, settings.Issuer, sdk.NewCoins(params.IssueFee)); err != nil {
			return "", err
		}
	}

	if err := k.SetSymbol(ctx, settings.Symbol, settings.Issuer); err != nil {
		return "", sdkerrors.Wrapf(err, "provided symbol: %s", settings.Symbol)
	}

	definition := types.Definition{
		Denom:              denom,
		Issuer:             settings.Issuer.String(),
		Features:           settings.Features,
		BurnRate:           settings.BurnRate,
		SendCommissionRate: settings.SendCommissionRate,
	}

	if err := k.SetDenomMetadata(ctx, denom, settings.Symbol, settings.Description, settings.Precision); err != nil {
		return "", err
	}
	k.SetDefinition(ctx, settings.Issuer, settings.Subunit, definition)

	if err := k.mintIfReceivable(ctx, definition, settings.InitialAmount, settings.Issuer); err != nil {
		return "", err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventIssued{
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
		return "", sdkerrors.Wrap(err, "can't emit EventIssued event")
	}

	k.logger(ctx).Debug("issued new fungible token with denom %d", denom)

	return denom, nil
}

// SetSymbol saves the symbol to store.
func (k Keeper) SetSymbol(ctx sdk.Context, symbol string, issuer sdk.AccAddress) error {
	symbol = types.NormalizeSymbolForKey(symbol)
	if k.isSymbolDuplicated(ctx, symbol, issuer) {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "duplicate symbol %s", symbol)
	}

	compositeKey := types.CreateSymbolKey(issuer, symbol)
	ctx.KVStore(k.storeKey).Set(compositeKey, []byte{0x01})
	return nil
}

// SetDefinition stores the Definition.
func (k Keeper) SetDefinition(ctx sdk.Context, issuer sdk.AccAddress, subunit string, definition types.Definition) {
	ctx.KVStore(k.storeKey).Set(types.CreateTokenKey(issuer, subunit), k.cdc.MustMarshal(&definition))
}

// SetDenomMetadata registers denom metadata on the bank keeper.
func (k Keeper) SetDenomMetadata(ctx sdk.Context, denom, symbol, description string, precision uint32) error {
	denomMetadata := banktypes.Metadata{
		Name:        symbol,
		Symbol:      symbol,
		Description: description,

		// This is a cosmos sdk requirement that the first denomination unit MUST be the base
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denom,
				Exponent: uint32(0),
			},
			{
				Denom:    symbol,
				Exponent: precision,
			},
		},
		// here take subunit provided by the user, generate the denom and used it as base,
		// and we take the symbol provided by the user and use it as symbol
		Base:    denom,
		Display: symbol,
	}

	if err := denomMetadata.Validate(); err != nil {
		return err
	}

	k.bankKeeper.SetDenomMetaData(ctx, denomMetadata)
	return nil
}

// Mint mints new fungible token.
func (k Keeper) Mint(ctx sdk.Context, sender sdk.AccAddress, coin sdk.Coin) error {
	def, err := k.GetDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	if err = def.CheckFeatureAllowed(sender, types.Feature_minting); err != nil {
		return err
	}

	return k.mintIfReceivable(ctx, def, coin.Amount, sender)
}

// Burn burns fungible token.
func (k Keeper) Burn(ctx sdk.Context, sender sdk.AccAddress, coin sdk.Coin) error {
	def, err := k.GetDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	err = def.CheckFeatureAllowed(sender, types.Feature_burning)
	if err != nil {
		return err
	}

	return k.burnIfSpendable(ctx, sender, def, coin.Amount)
}

// Freeze freezes specified token from the specified account.
func (k Keeper) Freeze(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error {
	if !coin.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "freeze amount should be positive")
	}

	def, err := k.GetDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	if def.IsIssuer(addr) {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "issuer's balance can't be frozen")
	}

	if err = def.CheckFeatureAllowed(sender, types.Feature_freezing); err != nil {
		return err
	}

	frozenStore := k.frozenAccountBalanceStore(ctx, addr)
	frozenBalance := frozenStore.Balance(coin.Denom)
	newFrozenBalance := frozenBalance.Add(coin)
	frozenStore.SetBalance(newFrozenBalance)

	return ctx.EventManager().EmitTypedEvent(&types.EventFrozenAmountChanged{
		Account:        addr.String(),
		Denom:          coin.Denom,
		PreviousAmount: frozenBalance.Amount,
		CurrentAmount:  newFrozenBalance.Amount,
	})
}

// Unfreeze unfreezes specified tokens from the specified account.
func (k Keeper) Unfreeze(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error {
	if !coin.IsPositive() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "freeze amount should be positive")
	}

	def, err := k.GetDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	if err = def.CheckFeatureAllowed(sender, types.Feature_freezing); err != nil {
		return err
	}

	frozenStore := k.frozenAccountBalanceStore(ctx, addr)
	frozenBalance := frozenStore.Balance(coin.Denom)
	if !frozenBalance.IsGTE(coin) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds,
			"unfreeze request %s is greater than the available frozen balance %s",
			coin.String(),
			frozenBalance.String(),
		)
	}

	newFrozenBalance := frozenBalance.Sub(coin)
	frozenStore.SetBalance(newFrozenBalance)

	return ctx.EventManager().EmitTypedEvent(&types.EventFrozenAmountChanged{
		Account:        addr.String(),
		Denom:          coin.Denom,
		PreviousAmount: frozenBalance.Amount,
		CurrentAmount:  newFrozenBalance.Amount,
	})
}

// GloballyFreeze enables global freeze on a fungible token. This function is idempotent.
func (k Keeper) GloballyFreeze(ctx sdk.Context, sender sdk.AccAddress, denom string) error {
	def, err := k.GetDefinition(ctx, denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", denom)
	}

	if err = def.CheckFeatureAllowed(sender, types.Feature_freezing); err != nil {
		return err
	}

	k.SetGlobalFreeze(ctx, denom, true)
	return nil
}

// GloballyUnfreeze disables global freeze on a fungible token. This function is idempotent.
func (k Keeper) GloballyUnfreeze(ctx sdk.Context, sender sdk.AccAddress, denom string) error {
	def, err := k.GetDefinition(ctx, denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", denom)
	}

	if err = def.CheckFeatureAllowed(sender, types.Feature_freezing); err != nil {
		return err
	}

	k.SetGlobalFreeze(ctx, denom, false)
	return nil
}

// GetAccountsFrozenBalances returns the frozen balance on all the account.
func (k Keeper) GetAccountsFrozenBalances(ctx sdk.Context, pagination *query.PageRequest) ([]types.Balance, *query.PageResponse, error) {
	return collectBalances(k.cdc, k.frozenBalancesStore(ctx), pagination)
}

// IterateAccountsFrozenBalances iterates over all frozen balances of all accounts and applies the provided callback.
// If true is returned from the callback, iteration is stopped.
func (k Keeper) IterateAccountsFrozenBalances(ctx sdk.Context, cb func(sdk.AccAddress, sdk.Coin) bool) error {
	return k.frozenAccountsBalanceStore(ctx).IterateAllBalances(cb)
}

// GetFrozenBalances returns the frozen balance of an account.
func (k Keeper) GetFrozenBalances(ctx sdk.Context, addr sdk.AccAddress, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error) {
	return k.frozenAccountBalanceStore(ctx, addr).Balances(pagination)
}

// GetFrozenBalance returns the frozen balance of a denom and account.
func (k Keeper) GetFrozenBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return k.frozenAccountBalanceStore(ctx, addr).Balance(denom)
}

// SetFrozenBalances sets the frozen balances of a specified account.
// Pay attention that the sdk.NewCoins() sanitizes/removes the empty coins, hence if you need set zero amount use the slice []sdk.Coins.
func (k Keeper) SetFrozenBalances(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	frozenStore := k.frozenAccountBalanceStore(ctx, addr)
	for _, coin := range coins {
		frozenStore.SetBalance(coin)
	}
}

// SetGlobalFreeze enables/disables global freeze on a fungible token depending on frozen arg.
func (k Keeper) SetGlobalFreeze(ctx sdk.Context, denom string, frozen bool) {
	if frozen {
		ctx.KVStore(k.storeKey).Set(types.CreateGlobalFreezeKey(denom), globalFreezeEnabledStoreVal)
		return
	}
	ctx.KVStore(k.storeKey).Delete(types.CreateGlobalFreezeKey(denom))
}

// SetWhitelistedBalance sets whitelisted limit for the account.
func (k Keeper) SetWhitelistedBalance(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error {
	if coin.IsNil() || coin.IsNegative() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidCoins, "whitelisted limit amount should be greater than or equal to 0")
	}

	def, err := k.GetDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	if def.IsIssuer(addr) {
		return sdkerrors.Wrap(sdkerrors.ErrUnauthorized, "issuer's balance can't be whitelisted")
	}

	if err = def.CheckFeatureAllowed(sender, types.Feature_whitelisting); err != nil {
		return err
	}

	whitelistedStore := k.whitelistedAccountBalanceStore(ctx, addr)
	previousWhitelistedBalance := whitelistedStore.Balance(coin.Denom)
	whitelistedStore.SetBalance(coin)

	return ctx.EventManager().EmitTypedEvent(&types.EventWhitelistedAmountChanged{
		Account:        addr.String(),
		Denom:          coin.Denom,
		PreviousAmount: previousWhitelistedBalance.Amount,
		CurrentAmount:  coin.Amount,
	})
}

// GetAccountsWhitelistedBalances returns the whitelisted balance of all the account.
func (k Keeper) GetAccountsWhitelistedBalances(ctx sdk.Context, pagination *query.PageRequest) ([]types.Balance, *query.PageResponse, error) {
	return collectBalances(k.cdc, prefix.NewStore(ctx.KVStore(k.storeKey), types.WhitelistedBalancesKeyPrefix), pagination)
}

// IterateAccountsWhitelistedBalances iterates over all whitelisted balances of all accounts and applies the provided callback.
// If true is returned from the callback, iteration is halted.
func (k Keeper) IterateAccountsWhitelistedBalances(ctx sdk.Context, cb func(sdk.AccAddress, sdk.Coin) bool) error {
	return newBalanceStore(k.cdc, ctx.KVStore(k.storeKey), types.WhitelistedBalancesKeyPrefix).IterateAllBalances(cb)
}

// GetWhitelistedBalances returns the whitelisted balance of an account.
func (k Keeper) GetWhitelistedBalances(ctx sdk.Context, addr sdk.AccAddress, pagination *query.PageRequest) (sdk.Coins, *query.PageResponse, error) {
	return k.whitelistedAccountBalanceStore(ctx, addr).Balances(pagination)
}

// GetWhitelistedBalance returns the whitelisted balance of a denom and account.
func (k Keeper) GetWhitelistedBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return k.whitelistedAccountBalanceStore(ctx, addr).Balance(denom)
}

// SetWhitelistedBalances sets the whitelisted balances of a specified account.
// Pay attention that the sdk.NewCoins() sanitizes/removes the empty coins, hence if you need set zero amount use the slice []sdk.Coins.
func (k Keeper) SetWhitelistedBalances(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	whitelistedStore := k.whitelistedAccountBalanceStore(ctx, addr)
	for _, coin := range coins {
		whitelistedStore.SetBalance(coin)
	}
}

func (k Keeper) mintIfReceivable(ctx sdk.Context, def types.Definition, amount sdk.Int, recipient sdk.AccAddress) error {
	if !amount.IsPositive() {
		return nil
	}
	if err := k.isCoinReceivable(ctx, recipient, def, amount); err != nil {
		return sdkerrors.Wrapf(err, "coins are not receivable")
	}

	coinsToMint := sdk.NewCoins(sdk.NewCoin(def.Denom, amount))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coinsToMint); err != nil {
		return sdkerrors.Wrapf(err, "can't mint %s for the module %s", coinsToMint.String(), types.ModuleName)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coinsToMint); err != nil {
		return sdkerrors.Wrapf(err, "can't send minted coins from module %s to account %s", types.ModuleName, recipient.String())
	}

	return nil
}

func (k Keeper) burnIfSpendable(ctx sdk.Context, account sdk.AccAddress, def types.Definition, amount sdk.Int) error {
	if err := k.isCoinSpendable(ctx, account, def, amount); err != nil {
		return sdkerrors.Wrapf(err, "coins are not spendable")
	}

	return k.burn(ctx, account, sdk.NewCoins(sdk.NewCoin(def.Denom, amount)))
}

func (k Keeper) burn(ctx sdk.Context, account sdk.AccAddress, coinsToBurn sdk.Coins) error {
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, account, types.ModuleName, coinsToBurn); err != nil {
		return sdkerrors.Wrapf(err, "can't send coins from account %s to module %s", account.String(), types.ModuleName)
	}

	if err := k.bankKeeper.BurnCoins(ctx, types.ModuleName, coinsToBurn); err != nil {
		return sdkerrors.Wrapf(err, "can't burn %s for the module %s", coinsToBurn.String(), types.ModuleName)
	}

	return nil
}

func (k Keeper) isCoinSpendable(ctx sdk.Context, addr sdk.AccAddress, def types.Definition, amount sdk.Int) error {
	if !def.IsFeatureEnabled(types.Feature_freezing) || def.IsIssuer(addr) {
		return nil
	}

	if k.isGloballyFrozen(ctx, def.Denom) {
		return sdkerrors.Wrapf(types.ErrGloballyFrozen, "%s is globally frozen", def.Denom)
	}

	availableBalance := k.availableBalance(ctx, addr, def.Denom)
	if !availableBalance.Amount.GTE(amount) {
		return sdkerrors.Wrapf(sdkerrors.ErrInsufficientFunds, "%s is not available, available %s",
			sdk.NewCoin(def.Denom, amount), availableBalance)
	}
	return nil
}

func (k Keeper) isCoinReceivable(ctx sdk.Context, addr sdk.AccAddress, def types.Definition, amount sdk.Int) error {
	if !def.IsFeatureEnabled(types.Feature_whitelisting) || def.IsIssuer(addr) {
		return nil
	}

	balance := k.bankKeeper.GetBalance(ctx, addr, def.Denom)
	whitelistedBalance := k.GetWhitelistedBalance(ctx, addr, def.Denom)

	finalBalance := balance.Amount.Add(amount)
	if finalBalance.GT(whitelistedBalance.Amount) {
		return sdkerrors.Wrapf(types.ErrWhitelistedLimitExceeded, "balance whitelisted for %s is not enough to receive %s, current whitelisted balance: %s",
			addr, sdk.NewCoin(def.Denom, amount), whitelistedBalance)
	}
	return nil
}

func (k Keeper) isSymbolDuplicated(ctx sdk.Context, symbol string, issuer sdk.AccAddress) bool {
	compositeKey := types.CreateSymbolKey(issuer, symbol)
	rawBytes := ctx.KVStore(k.storeKey).Get(compositeKey)
	return rawBytes != nil
}

func (k Keeper) availableBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	balance := k.bankKeeper.GetBalance(ctx, addr, denom)
	if balance.IsZero() {
		return balance
	}

	frozenBalance := k.GetFrozenBalance(ctx, addr, denom)
	if frozenBalance.IsGTE(balance) {
		return sdk.NewCoin(denom, sdk.ZeroInt())
	}
	return balance.Sub(frozenBalance)
}

func (k Keeper) getDefinitions(ctx sdk.Context, pagination *query.PageRequest) ([]types.Definition, *query.PageResponse, error) {
	return k.getDefinitionsFromStore(prefix.NewStore(ctx.KVStore(k.storeKey), types.TokenKeyPrefix), pagination)
}

func (k Keeper) getIssuerDefinitions(ctx sdk.Context, issuer sdk.AccAddress, pagination *query.PageRequest) ([]types.Definition, *query.PageResponse, error) {
	return k.getDefinitionsFromStore(prefix.NewStore(ctx.KVStore(k.storeKey), types.CreateIssuerTokensPrefix(issuer)), pagination)
}

func (k Keeper) getTokenFullInfo(ctx sdk.Context, definition types.Definition) (types.Token, error) {
	subunit, _, err := types.DeconstructDenom(definition.Denom)
	if err != nil {
		return types.Token{}, err
	}

	metadata, found := k.bankKeeper.GetDenomMetaData(ctx, definition.Denom)
	if !found {
		return types.Token{}, sdkerrors.Wrapf(types.ErrTokenNotFound, "metadata for %s denom not found", definition.Denom)
	}

	precision := -1
	for _, unit := range metadata.DenomUnits {
		if unit.Denom == metadata.Symbol {
			precision = int(unit.Exponent)
			break
		}
	}

	if precision < 0 {
		return types.Token{}, sdkerrors.Wrap(types.ErrInvalidInput, "precision not found")
	}

	return types.Token{
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

func (k Keeper) getDefinitionsFromStore(store prefix.Store, pagination *query.PageRequest) ([]types.Definition, *query.PageResponse, error) {
	definitionsPointers, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
		pagination,
		// builder
		func(key []byte, definition *types.Definition) (*types.Definition, error) {
			return definition, nil
		},
		// constructor
		func() *types.Definition {
			return &types.Definition{}
		},
	)
	if err != nil {
		return nil, nil, err
	}

	definitions := make([]types.Definition, 0, len(definitionsPointers))
	for _, definition := range definitionsPointers {
		definitions = append(definitions, *definition)
	}

	return definitions, pageRes, err
}

func (k Keeper) getTokensByDefinitions(ctx sdk.Context, defs []types.Definition) ([]types.Token, error) {
	tokens := make([]types.Token, 0, len(defs))
	for _, definition := range defs {
		token, err := k.getTokenFullInfo(ctx, definition)
		if err != nil {
			return nil, err
		}

		tokens = append(tokens, token)
	}
	return tokens, nil
}

// frozenBalancesStore get the store for the frozen balances of all accounts.
func (k Keeper) frozenBalancesStore(ctx sdk.Context) prefix.Store {
	return prefix.NewStore(ctx.KVStore(k.storeKey), types.FrozenBalancesKeyPrefix)
}

// frozenAccountBalanceStore gets the store for the frozen balances of an account.
func (k Keeper) frozenAccountBalanceStore(ctx sdk.Context, addr sdk.AccAddress) balanceStore {
	return newBalanceStore(k.cdc, ctx.KVStore(k.storeKey), types.CreateFrozenBalancesKey(addr))
}

// frozenAccountBalanceStore gets the store for the frozen balances of an account.
func (k Keeper) frozenAccountsBalanceStore(ctx sdk.Context) balanceStore {
	return newBalanceStore(k.cdc, ctx.KVStore(k.storeKey), types.FrozenBalancesKeyPrefix)
}

func (k Keeper) isGloballyFrozen(ctx sdk.Context, denom string) bool {
	globFreezeVal := ctx.KVStore(k.storeKey).Get(types.CreateGlobalFreezeKey(denom))
	return bytes.Equal(globFreezeVal, globalFreezeEnabledStoreVal)
}

// whitelistedAccountBalanceStore gets the store for the whitelisted balances of an account.
func (k Keeper) whitelistedAccountBalanceStore(ctx sdk.Context, addr sdk.AccAddress) balanceStore {
	return newBalanceStore(k.cdc, ctx.KVStore(k.storeKey), types.CreateWhitelistedBalancesKey(addr))
}

// logger returns the Keeper logger.
func (k Keeper) logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
