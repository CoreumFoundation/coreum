package keeper

import (
	"bytes"
	"encoding/json"
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v4/x/wasm"
	cwasmtypes "github.com/CoreumFoundation/coreum/v4/x/wasm/types"
	wibctransfertypes "github.com/CoreumFoundation/coreum/v4/x/wibctransfer/types"
)

// ExtensionInstantiateMsg is the message passed to the extension cosmwasm contract.
// The contract must be able to properly process this message.
//
//nolint:tagliatelle // these will be exposed to rust and must be snake case.
type ExtensionInstantiateMsg struct {
	Denom       string                       `json:"denom"`
	IssuanceMsg wasmtypes.RawContractMessage `json:"issuance_msg"`
}

// Keeper is the asset module keeper.
type Keeper struct {
	cdc                    codec.BinaryCodec
	storeKey               storetypes.StoreKey
	bankKeeper             types.BankKeeper
	delayKeeper            types.DelayKeeper
	wasmKeeper             cwasmtypes.WasmKeeper
	wasmPermissionedKeeper types.WasmPermissionedKeeper
	accountKeeper          types.AccountKeeper
	authority              string
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	bankKeeper types.BankKeeper,
	delayKeeper types.DelayKeeper,
	wasmKeeper cwasmtypes.WasmKeeper,
	wasmPermissionedKeeper types.WasmPermissionedKeeper,
	accountKeeper types.AccountKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:                    cdc,
		storeKey:               storeKey,
		bankKeeper:             bankKeeper,
		delayKeeper:            delayKeeper,
		wasmKeeper:             wasmKeeper,
		wasmPermissionedKeeper: wasmPermissionedKeeper,
		accountKeeper:          accountKeeper,
		authority:              authority,
	}
}

// GetParams gets the parameters of the module.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params
}

// SetParams sets the parameters of the module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(types.ParamsKey, bz)
	return nil
}

// UpdateParams is a governance operation that sets parameters of the module.
func (k Keeper) UpdateParams(ctx sdk.Context, authority string, params types.Params) error {
	if k.authority != authority {
		return sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, authority)
	}

	return k.SetParams(ctx, params)
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
func (k Keeper) GetIssuerTokens(
	ctx sdk.Context,
	issuer sdk.AccAddress,
	pagination *query.PageRequest,
) ([]types.Token, *query.PageResponse, error) {
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

// IterateAllDefinitions iterates over all token definitions and applies the provided callback.
// If true is returned from the callback, iteration is halted.
func (k Keeper) IterateAllDefinitions(ctx sdk.Context, cb func(types.Definition) (bool, error)) error {
	iterator := prefix.NewStore(ctx.KVStore(k.storeKey), types.TokenKeyPrefix).Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var definition types.Definition
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

// GetToken returns the fungible token by it's denom.
func (k Keeper) GetToken(ctx sdk.Context, denom string) (types.Token, error) {
	def, err := k.GetDefinition(ctx, denom)
	if err != nil {
		return types.Token{}, err
	}

	return k.getTokenFullInfo(ctx, def)
}

// Issue issues new fungible token and returns it's denom.
func (k Keeper) Issue(ctx sdk.Context, settings types.IssueSettings) (string, error) {
	return k.IssueVersioned(ctx, settings, types.CurrentTokenVersion)
}

// IssueVersioned issues new fungible token and sets its version.
// To be used only in unit tests !!!
//
//nolint:funlen // breaking down this function will make it less readable.
func (k Keeper) IssueVersioned(ctx sdk.Context, settings types.IssueSettings, version uint32) (string, error) {
	if err := types.ValidateSubunit(settings.Subunit); err != nil {
		return "", sdkerrors.Wrapf(err, "provided subunit: %s", settings.Subunit)
	}

	if err := types.ValidatePrecision(settings.Precision); err != nil {
		return "", sdkerrors.Wrapf(err, "provided precision: %d", settings.Precision)
	}

	if err := types.ValidateFeatures(settings.Features); err != nil {
		return "", err
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
		Version:            version,
		URI:                settings.URI,
		URIHash:            settings.URIHash,
		Admin:              settings.Issuer.String(),
	}

	if err := k.mintIfReceivable(ctx, definition, settings.InitialAmount, settings.Issuer); err != nil {
		return "", err
	}

	if definition.IsFeatureEnabled(types.Feature_extension) {
		if settings.ExtensionSettings == nil {
			return "", types.ErrInvalidInput.Wrap("extension settings must be provided")
		}

		if len(settings.ExtensionSettings.IssuanceMsg) == 0 {
			settings.ExtensionSettings.IssuanceMsg = []byte("{}")
		}

		instantiateMsgBytes, err := json.Marshal(ExtensionInstantiateMsg{
			Denom:       denom,
			IssuanceMsg: settings.ExtensionSettings.IssuanceMsg,
		})
		if err != nil {
			return "", types.ErrInvalidInput.Wrapf("error marshalling ExtensionInstantiateMsg (%s)", err)
		}

		contractAddress, _, err := k.wasmPermissionedKeeper.Instantiate2(
			ctx,
			settings.ExtensionSettings.CodeId,
			settings.Issuer,
			settings.Issuer,
			instantiateMsgBytes,
			settings.ExtensionSettings.Label,
			settings.ExtensionSettings.Funds,
			ctx.BlockHeader().AppHash,
			true,
		)
		if err != nil {
			return "", sdkerrors.Wrapf(err, "error instantiating cw contract")
		}

		definition.ExtensionCWAddress = contractAddress.String()
	}

	if err := k.SetDenomMetadata(
		ctx,
		denom,
		settings.Symbol,
		settings.Description,
		settings.URI,
		settings.URIHash,
		settings.Precision,
	); err != nil {
		return "", err
	}

	k.SetDefinition(ctx, settings.Issuer, settings.Subunit, definition)

	if settings.DEXSettings != nil {
		if err := types.ValidateDEXSettings(*settings.DEXSettings); err != nil {
			return "", err
		}
		k.SetDEXSettings(ctx, denom, *settings.DEXSettings)
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
		URI:                settings.URI,
		URIHash:            settings.URIHash,
		Admin:              settings.Issuer.String(),
		DEXSettings:        settings.DEXSettings,
	}); err != nil {
		return "", sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit EventIssued event: %s", err)
	}

	k.logger(ctx).Debug(
		"issued new fungible token",
		"denom", denom,
		"settings", settings,
	)

	return denom, nil
}

// SetSymbol saves the symbol to store.
func (k Keeper) SetSymbol(ctx sdk.Context, symbol string, issuer sdk.AccAddress) error {
	symbol = types.NormalizeSymbolForKey(symbol)
	if k.isSymbolDuplicated(ctx, symbol, issuer) {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "duplicate symbol %s", symbol)
	}

	ctx.KVStore(k.storeKey).Set(types.CreateSymbolKey(issuer, symbol), types.StoreTrue)
	return nil
}

// SetDefinition stores the Definition.
func (k Keeper) SetDefinition(ctx sdk.Context, issuer sdk.AccAddress, subunit string, definition types.Definition) {
	ctx.KVStore(k.storeKey).Set(types.CreateTokenKey(issuer, subunit), k.cdc.MustMarshal(&definition))
}

// SetDenomMetadata registers denom metadata on the bank keeper.
func (k Keeper) SetDenomMetadata(
	ctx sdk.Context,
	denom, symbol, description, uri, uriHash string,
	precision uint32,
) error {
	denomMetadata := banktypes.Metadata{
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
		Name:    symbol,
		Symbol:  symbol,
		URI:     uri,
		URIHash: uriHash,
	}

	// in case the precision is zero, we cannot 2 zero exponents in denom units, so
	// we are force to have single entry in denom units and also Display must be the
	// same as Base.
	if precision == 0 {
		denomMetadata.DenomUnits = []*banktypes.DenomUnit{
			{
				Denom:    denom,
				Exponent: uint32(0),
			},
		}
		denomMetadata.Display = denom
	}

	if err := denomMetadata.Validate(); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "failed to validate denom metadata: %s", err)
	}

	k.bankKeeper.SetDenomMetaData(ctx, denomMetadata)
	return nil
}

// Mint mints new fungible token.
func (k Keeper) Mint(ctx sdk.Context, sender, recipient sdk.AccAddress, coin sdk.Coin) error {
	def, err := k.GetDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	if err = def.CheckFeatureAllowed(sender, types.Feature_minting); err != nil {
		return err
	}

	return k.mintIfReceivable(ctx, def, coin.Amount, recipient)
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
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidCoins, "freeze amount should be positive")
	}

	if err := k.freezingChecks(ctx, sender, addr, coin); err != nil {
		return err
	}

	frozenStore := k.frozenAccountBalanceStore(ctx, addr)
	frozenBalance := frozenStore.Balance(coin.Denom)
	newFrozenBalance := frozenBalance.Add(coin)
	frozenStore.SetBalance(newFrozenBalance)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventFrozenAmountChanged{
		Account:        addr.String(),
		Denom:          coin.Denom,
		PreviousAmount: frozenBalance.Amount,
		CurrentAmount:  newFrozenBalance.Amount,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit EventFrozenAmountChanged event: %s", err)
	}

	return nil
}

// Unfreeze unfreezes specified tokens from the specified account.
func (k Keeper) Unfreeze(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error {
	if !coin.IsPositive() {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidCoins, "freeze amount should be positive")
	}

	if err := k.freezingChecks(ctx, sender, addr, coin); err != nil {
		return err
	}

	frozenStore := k.frozenAccountBalanceStore(ctx, addr)
	frozenBalance := frozenStore.Balance(coin.Denom)
	if !frozenBalance.IsGTE(coin) {
		return sdkerrors.Wrapf(cosmoserrors.ErrInsufficientFunds,
			"unfreeze request %s is greater than the available frozen balance %s",
			coin.String(),
			frozenBalance.String(),
		)
	}

	newFrozenBalance := frozenBalance.Sub(coin)
	frozenStore.SetBalance(newFrozenBalance)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventFrozenAmountChanged{
		Account:        addr.String(),
		Denom:          coin.Denom,
		PreviousAmount: frozenBalance.Amount,
		CurrentAmount:  newFrozenBalance.Amount,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit EventFrozenAmountChanged event: %s", err)
	}

	return nil
}

// SetFrozen sets frozen amount on the specified account.
func (k Keeper) SetFrozen(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error {
	if coin.IsNegative() {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidCoins, "frozen amount must not be negative")
	}

	if err := k.freezingChecks(ctx, sender, addr, coin); err != nil {
		return err
	}

	frozenStore := k.frozenAccountBalanceStore(ctx, addr)
	frozenBalance := frozenStore.Balance(coin.Denom)
	frozenStore.SetBalance(coin)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventFrozenAmountChanged{
		Account:        addr.String(),
		Denom:          coin.Denom,
		PreviousAmount: frozenBalance.Amount,
		CurrentAmount:  coin.Amount,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit EventFrozenAmountChanged event: %s", err)
	}

	return nil
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
func (k Keeper) GetAccountsFrozenBalances(
	ctx sdk.Context,
	pagination *query.PageRequest,
) ([]types.Balance, *query.PageResponse, error) {
	return collectBalances(k.cdc, k.frozenBalancesStore(ctx), pagination)
}

// IterateAccountsFrozenBalances iterates over all frozen balances of all accounts and applies the provided callback.
// If true is returned from the callback, iteration is stopped.
func (k Keeper) IterateAccountsFrozenBalances(ctx sdk.Context, cb func(sdk.AccAddress, sdk.Coin) bool) error {
	return k.frozenAccountsBalanceStore(ctx).IterateAllBalances(cb)
}

// GetFrozenBalances returns the frozen balance of an account.
func (k Keeper) GetFrozenBalances(
	ctx sdk.Context,
	addr sdk.AccAddress,
	pagination *query.PageRequest,
) (sdk.Coins, *query.PageResponse, error) {
	return k.frozenAccountBalanceStore(ctx, addr).Balances(pagination)
}

// GetFrozenBalance returns the frozen balance of a denom and account.
func (k Keeper) GetFrozenBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if k.isGloballyFrozen(ctx, denom) {
		return k.bankKeeper.GetBalance(ctx, addr, denom)
	}
	return k.frozenAccountBalanceStore(ctx, addr).Balance(denom)
}

// SetFrozenBalances sets the frozen balances of a specified account.
// Pay attention that the sdk.NewCoins() sanitizes/removes the empty coins,
// hence if you need set zero amount use the slice []sdk.Coins.
func (k Keeper) SetFrozenBalances(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	frozenStore := k.frozenAccountBalanceStore(ctx, addr)
	for _, coin := range coins {
		frozenStore.SetBalance(coin)
	}
}

// SetGlobalFreeze enables/disables global freeze on a fungible token depending on frozen arg.
func (k Keeper) SetGlobalFreeze(ctx sdk.Context, denom string, frozen bool) {
	if frozen {
		ctx.KVStore(k.storeKey).Set(types.CreateGlobalFreezeKey(denom), types.StoreTrue)
		return
	}
	ctx.KVStore(k.storeKey).Delete(types.CreateGlobalFreezeKey(denom))
}

// Clawback confiscates specified token from the specified account.
func (k Keeper) Clawback(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error {
	if !coin.IsPositive() {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidCoins, "clawback amount should be positive")
	}

	if err := k.validateClawbackAllowed(ctx, sender, addr, coin); err != nil {
		return err
	}

	if err := k.bankKeeper.SendCoins(ctx, addr, sender, sdk.NewCoins(coin)); err != nil {
		return sdkerrors.Wrapf(err, "can't send coins from account %s to issuer %s", addr.String(), sender.String())
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventAmountClawedBack{
		Account: addr.String(),
		Denom:   coin.Denom,
		Amount:  coin.Amount,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit EventAmountClawedBack event: %s", err)
	}

	return nil
}

// SetWhitelistedBalance sets whitelisted limit for the account.
func (k Keeper) SetWhitelistedBalance(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error {
	if coin.IsNil() || coin.IsNegative() {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidCoins, "whitelisted limit amount should be greater than or equal to 0")
	}

	def, err := k.GetDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	if def.IsAdmin(addr) {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "admin's balance can't be whitelisted")
	}

	if err = def.CheckFeatureAllowed(sender, types.Feature_whitelisting); err != nil {
		return err
	}

	whitelistedStore := k.whitelistedAccountBalanceStore(ctx, addr)
	previousWhitelistedBalance := whitelistedStore.Balance(coin.Denom)
	whitelistedStore.SetBalance(coin)

	if err = ctx.EventManager().EmitTypedEvent(&types.EventWhitelistedAmountChanged{
		Account:        addr.String(),
		Denom:          coin.Denom,
		PreviousAmount: previousWhitelistedBalance.Amount,
		CurrentAmount:  coin.Amount,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit EventWhitelistedAmountChanged event: %s", err)
	}

	return nil
}

// GetAccountsWhitelistedBalances returns the whitelisted balance of all the account.
func (k Keeper) GetAccountsWhitelistedBalances(
	ctx sdk.Context,
	pagination *query.PageRequest,
) ([]types.Balance, *query.PageResponse, error) {
	return collectBalances(
		k.cdc, prefix.NewStore(ctx.KVStore(k.storeKey), types.WhitelistedBalancesKeyPrefix), pagination)
}

// IterateAccountsWhitelistedBalances iterates over all whitelisted balances of all accounts
// and applies the provided callback.
// If true is returned from the callback, iteration is halted.
func (k Keeper) IterateAccountsWhitelistedBalances(ctx sdk.Context, cb func(sdk.AccAddress, sdk.Coin) bool) error {
	return newBalanceStore(k.cdc, ctx.KVStore(k.storeKey), types.WhitelistedBalancesKeyPrefix).IterateAllBalances(cb)
}

// GetWhitelistedBalances returns the whitelisted balance of an account.
func (k Keeper) GetWhitelistedBalances(
	ctx sdk.Context,
	addr sdk.AccAddress,
	pagination *query.PageRequest,
) (sdk.Coins, *query.PageResponse, error) {
	return k.whitelistedAccountBalanceStore(ctx, addr).Balances(pagination)
}

// GetWhitelistedBalance returns the whitelisted balance of a denom and account.
func (k Keeper) GetWhitelistedBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return k.whitelistedAccountBalanceStore(ctx, addr).Balance(denom)
}

// SetWhitelistedBalances sets the whitelisted balances of a specified account.
// Pay attention that the sdk.NewCoins() sanitizes/removes the empty coins, hence if you
// need set zero amount use the slice []sdk.Coins.
func (k Keeper) SetWhitelistedBalances(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	whitelistedStore := k.whitelistedAccountBalanceStore(ctx, addr)
	for _, coin := range coins {
		whitelistedStore.SetBalance(coin)
	}
}

// DEXLock locks specified token from the specified account.
func (k Keeper) DEXLock(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error {
	// in the final implementation the `DEXLock` will accept lock reason struct from dex, to let assetft decide
	// whether the locking is allowed. The same struct will be passed to the extensions smart contract.
	if !coin.IsPositive() {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidCoins, "DEX locking amount must be positive")
	}

	if err := k.dexLockingChecks(ctx, addr, coin); err != nil {
		return err
	}

	dexLockedStore := k.dexLockedAccountBalanceStore(ctx, addr)
	dexLockedBalance := dexLockedStore.Balance(coin.Denom)
	newDEXLockedBalance := dexLockedBalance.Add(coin)
	dexLockedStore.SetBalance(newDEXLockedBalance)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventDEXLockedAmountChanged{
		Account:        addr.String(),
		Denom:          coin.Denom,
		PreviousAmount: dexLockedBalance.Amount,
		CurrentAmount:  newDEXLockedBalance.Amount,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit EventDEXLockedAmountChanged event: %s", err)
	}

	return nil
}

// DEXUnlock unlocks specified tokens from the specified account.
func (k Keeper) DEXUnlock(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error {
	if !coin.IsPositive() {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidCoins, "DEX unlock amount should be positive")
	}

	dexLockedStore := k.dexLockedAccountBalanceStore(ctx, addr)
	dexLockedBalance := dexLockedStore.Balance(coin.Denom)
	if !dexLockedBalance.IsGTE(coin) {
		return sdkerrors.Wrapf(cosmoserrors.ErrInsufficientFunds,
			"DEX unlock request %s is greater than the available locked balance %s",
			coin.String(),
			dexLockedBalance.String(),
		)
	}

	newDEXLockedBalance := dexLockedBalance.Sub(coin)
	dexLockedStore.SetBalance(newDEXLockedBalance)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventDEXLockedAmountChanged{
		Account:        addr.String(),
		Denom:          coin.Denom,
		PreviousAmount: dexLockedBalance.Amount,
		CurrentAmount:  newDEXLockedBalance.Amount,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit EventDEXLockedAmountChanged event: %s", err)
	}

	return nil
}

// DEXUnlockAndSend unlock the coin on the `fromAddr` balance and send to the `toAddr`.
func (k Keeper) DEXUnlockAndSend(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, coin sdk.Coin) error {
	if err := k.DEXUnlock(ctx, fromAddr, coin); err != nil {
		return err
	}
	if err := k.bankKeeper.SendCoins(ctx, fromAddr, toAddr, sdk.NewCoins(coin)); err != nil {
		return sdkerrors.Wrap(err, "failed to send DEX unlocked coins")
	}

	return nil
}

// GetDEXLockedBalance returns the DEX locked balance.
func (k Keeper) GetDEXLockedBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return k.dexLockedAccountBalanceStore(ctx, addr).Balance(denom)
}

// GetDEXLockedBalances returns the DEX locked balances of an account.
func (k Keeper) GetDEXLockedBalances(
	ctx sdk.Context,
	addr sdk.AccAddress,
	pagination *query.PageRequest,
) (sdk.Coins, *query.PageResponse, error) {
	return k.dexLockedAccountBalanceStore(ctx, addr).Balances(pagination)
}

// GetAccountsDEXLockedBalances returns the DEX locked balance on all the account.
func (k Keeper) GetAccountsDEXLockedBalances(
	ctx sdk.Context,
	pagination *query.PageRequest,
) ([]types.Balance, *query.PageResponse, error) {
	return collectBalances(k.cdc, k.dexLockedBalancesStore(ctx), pagination)
}

// SetDEXLockedBalances sets the DEX locked balances of a specified account.
// Pay attention that the sdk.NewCoins() sanitizes/removes the empty coins, hence if you
// need set zero amount use the slice []sdk.Coins.
func (k Keeper) SetDEXLockedBalances(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	dexLockedStore := k.dexLockedAccountBalanceStore(ctx, addr)
	for _, coin := range coins {
		dexLockedStore.SetBalance(coin)
	}
}

// GetSpendableBalance returns balance allowed to be spendable.
func (k Keeper) GetSpendableBalance(
	ctx sdk.Context,
	addr sdk.AccAddress,
	denom string,
) sdk.Coin {
	balance := k.bankKeeper.GetBalance(ctx, addr, denom)
	if balance.Amount.IsZero() {
		return balance
	}

	notLockedAmt := balance.Amount.
		Sub(k.GetDEXLockedBalance(ctx, addr, denom).Amount).
		Sub(k.bankKeeper.LockedCoins(ctx, addr).AmountOf(denom))
	if notLockedAmt.IsNegative() {
		return sdk.NewCoin(denom, sdkmath.ZeroInt())
	}

	notFrozenAmt := balance.Amount.Sub(k.GetFrozenBalance(ctx, addr, denom).Amount)
	if notFrozenAmt.IsNegative() {
		return sdk.NewCoin(denom, sdkmath.ZeroInt())
	}

	spendableAmount := sdkmath.MinInt(notLockedAmt, notFrozenAmt)
	return sdk.NewCoin(denom, spendableAmount)
}

// TransferAdmin changes admin of a fungible token.
func (k Keeper) TransferAdmin(ctx sdk.Context, sender, addr sdk.AccAddress, denom string) error {
	def, err := k.GetDefinition(ctx, denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", denom)
	}

	if !def.IsAdmin(sender) {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "only admin can transfer administration of an account")
	}

	previousAdmin := def.Admin

	subunit, issuer, err := types.DeconstructDenom(denom)
	if err != nil {
		return err
	}

	def.Admin = addr.String()
	k.SetDefinition(ctx, issuer, subunit, def)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventAdminTransferred{
		Denom:         denom,
		PreviousAdmin: previousAdmin,
		CurrentAdmin:  def.Admin,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit EventAdminTransferred event: %s", err)
	}

	return nil
}

// ClearAdmin removes admin of a fungible token.
func (k Keeper) ClearAdmin(ctx sdk.Context, sender sdk.AccAddress, denom string) error {
	def, err := k.GetDefinition(ctx, denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", denom)
	}

	if !def.IsAdmin(sender) {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "only admin can remove administration of an account")
	}

	previousAdmin := def.Admin

	subunit, issuer, err := types.DeconstructDenom(denom)
	if err != nil {
		return err
	}

	// if extension feature is disabled, after clearing admin, there is no one to send commission to, so the commission
	// rate sets to zero else only the admin is cleared and the extension receives the commission rate
	def.Admin = ""
	if !def.IsFeatureEnabled(types.Feature_extension) {
		def.SendCommissionRate = sdkmath.LegacyZeroDec()
	}

	k.SetDefinition(ctx, issuer, subunit, def)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventAdminCleared{
		Denom:         denom,
		PreviousAdmin: previousAdmin,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit EventAdminCleared event: %s", err)
	}

	return nil
}

// UpdateDEXSettings updates the DEX settings of a specified denom.
func (k Keeper) UpdateDEXSettings(
	ctx sdk.Context,
	sender sdk.AccAddress,
	denom string,
	settings types.DEXSettings,
) error {
	if err := types.ValidateDEXSettings(settings); err != nil {
		return err
	}

	if k.authority != sender.String() {
		def, err := k.GetDefinition(ctx, denom)
		if err != nil {
			return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", denom)
		}
		if !def.IsAdmin(sender) {
			return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "only admin and gov can update DEX settings")
		}
	}

	prevSettings, err := k.getDEXSettingsOrNil(ctx, denom)
	if err != nil {
		return err
	}
	k.SetDEXSettings(ctx, denom, settings)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventDEXSettingsChanged{
		PreviousSettings: prevSettings,
		NewSettings:      settings,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit EventDEXSettingsChanged event: %s", err)
	}

	return nil
}

// SetDEXSettings sets the DEX settings of a specified denom.
func (k Keeper) SetDEXSettings(ctx sdk.Context, denom string, settings types.DEXSettings) {
	ctx.KVStore(k.storeKey).Set(types.CreateDEXSettingsKey(denom), k.cdc.MustMarshal(&settings))
}

// GetDEXSettings gets the DEX settings of a specified denom.
func (k Keeper) GetDEXSettings(ctx sdk.Context, denom string) (types.DEXSettings, error) {
	bz := ctx.KVStore(k.storeKey).Get(types.CreateDEXSettingsKey(denom))
	if bz == nil {
		return types.DEXSettings{}, sdkerrors.Wrapf(types.ErrDEXSettingsNotFound, "denom: %s", denom)
	}
	var settings types.DEXSettings
	k.cdc.MustUnmarshal(bz, &settings)

	return settings, nil
}

// GetDEXSettingsWithDenoms returns all DEX settings with the corresponding denoms.
func (k Keeper) GetDEXSettingsWithDenoms(
	ctx sdk.Context,
	pagination *query.PageRequest,
) ([]types.DEXSettingsWithDenom, *query.PageResponse, error) {
	dexSettings := make([]types.DEXSettingsWithDenom, 0)
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DEXSettingsKeyPrefix)
	pageRes, err := query.Paginate(store, pagination, func(key, value []byte) error {
		denom, err := types.DecodeDenomFromKey(key)
		if err != nil {
			return err
		}
		var settings types.DEXSettings
		k.cdc.MustUnmarshal(value, &settings)

		dexSettings = append(dexSettings, types.DEXSettingsWithDenom{
			Denom:       denom,
			DEXSettings: settings,
		})

		return nil
	})

	return dexSettings, pageRes, err
}

func (k Keeper) mintIfReceivable(
	ctx sdk.Context,
	def types.Definition,
	amount sdkmath.Int,
	recipient sdk.AccAddress,
) error {
	if !amount.IsPositive() {
		return nil
	}

	if wasm.IsSmartContract(ctx, recipient, k.wasmKeeper) {
		ctx = cwasmtypes.WithSmartContractRecipient(ctx, recipient.String())
	}

	if err := k.validateCoinReceivable(ctx, recipient, def, amount); err != nil {
		return sdkerrors.Wrapf(err, "coins are not receivable")
	}

	coinsToMint := sdk.NewCoins(sdk.NewCoin(def.Denom, amount))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coinsToMint); err != nil {
		return sdkerrors.Wrapf(err, "can't mint %s for the module %s", coinsToMint.String(), types.ModuleName)
	}

	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coinsToMint); err != nil {
		return sdkerrors.Wrapf(
			err,
			"can't send minted coins from module %s to account %s",
			types.ModuleName,
			recipient.String(),
		)
	}

	return nil
}

func (k Keeper) burnIfSpendable(
	ctx sdk.Context,
	account sdk.AccAddress,
	def types.Definition,
	amount sdkmath.Int,
) error {
	if err := k.validateCoinSpendable(ctx, account, def, amount); err != nil {
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

func (k Keeper) validateCoinSpendable(
	ctx sdk.Context,
	addr sdk.AccAddress,
	def types.Definition,
	amount sdkmath.Int,
) error {
	// This check is effective when IBC transfer is acknowledged by the peer chain. It happens in two situations:
	// - when transfer succeeded
	// - when transfer has been rejected by the other chain.
	// `validateCoinSpendable` is called only in the second case, that's why we don't need to differentiate them here.
	// So, whenever it happens here, it means transfer has been rejected. It means that funds are going to be refunded
	// back to the sender by the IBC transfer module.
	// It should succeed even if the issuer decided, for whatever reason, to freeze the escrow address.
	// It is done before checking for global freeze because refunding should not be blocked by this.
	// Otherwise, funds would be lost forever, being blocked on the escrow account.
	if wibctransfertypes.IsPurposeAck(ctx) {
		return nil
	}

	// Same thing applies if IBC fails due to timeout.
	if wibctransfertypes.IsPurposeTimeout(ctx) {
		return nil
	}

	if def.IsFeatureEnabled(types.Feature_freezing) &&
		k.isGloballyFrozen(ctx, def.Denom) &&
		!def.HasAdminPrivileges(addr) {
		return sdkerrors.Wrapf(types.ErrGloballyFrozen, "%s is globally frozen", def.Denom)
	}

	// Checking for IBC-received transfer is done here (after call to k.isGloballyFrozen), because those transfers
	// should be affected by the global freeze checked above. We decided that if token is frozen globally it means
	// none of the balances for that token can be affected by the IBC incoming transfer during freezing period.
	// Otherwise, the transfer must work despite the fact that escrow address might have been frozen by the issuer.
	if wibctransfertypes.IsPurposeIn(ctx) {
		return nil
	}

	if def.IsFeatureEnabled(types.Feature_block_smart_contracts) &&
		!def.HasAdminPrivileges(addr) &&
		cwasmtypes.IsTriggeredBySmartContract(ctx) {
		return sdkerrors.Wrapf(
			cosmoserrors.ErrUnauthorized,
			"transfers made by smart contracts are disabled for %s",
			def.Denom,
		)
	}

	balance := k.bankKeeper.GetBalance(ctx, addr, def.Denom)
	if err := k.validateCoinIsNotLockedByDEXAndBank(ctx, addr, balance, sdk.NewCoin(def.Denom, amount)); err != nil {
		return err
	}

	if def.IsFeatureEnabled(types.Feature_freezing) && !def.HasAdminPrivileges(addr) {
		frozenAmt := k.GetFrozenBalance(ctx, addr, def.Denom).Amount
		notFrozenAmt := balance.Amount.Sub(frozenAmt)
		if notFrozenAmt.LT(amount) {
			return sdkerrors.Wrapf(cosmoserrors.ErrInsufficientFunds, "%s%s is not available, available %s%s",
				amount.String(), def.Denom, notFrozenAmt.String(), def.Denom)
		}
	}

	return nil
}

func (k Keeper) validateCoinReceivable(
	ctx sdk.Context,
	addr sdk.AccAddress,
	def types.Definition,
	amount sdkmath.Int,
) error {
	// This check is effective when funds for IBC transfers are received by the escrow address.
	// If IBC is enabled we always accept escrow address as a receiver of the funds because it must work
	// despite the fact that address is not whitelisted.
	// On the other hand, if IBC is disabled for the token, we reject the transfer to the escrow address.
	// We don't block on IsPurposeIn condition when IBC transfer is received because if token cannot be sent,
	// it cannot be received back by definition.
	if wibctransfertypes.IsPurposeOut(ctx) {
		if !def.IsFeatureEnabled(types.Feature_ibc) {
			return sdkerrors.Wrapf(cosmoserrors.ErrUnauthorized, "ibc transfers are disabled for %s", def.Denom)
		}
		return nil
	}

	// This check is effective when IBC transfer is acknowledged by the peer chain. It happens in two situations:
	// - when transfer succeeded
	// - when transfer has been rejected by the other chain.
	// `validateCoinReceivable` is called only in the second case, that's why we don't need to differentiate them here.
	// So, whenever it happens here, it means transfer has been rejected. It means that funds are going to be refunded
	// back to the sender by the IBC transfer module.
	// That means we should allow to do this even if the sender is no longer whitelisted. It might happen that between
	// sending IBC transfer and receiving ACK rejecting it, issuer decided to remove whitelisting for the sender.
	// Despite that, sender should receive his funds back because otherwise they are lost forever, being blocked
	// on the escrow address.
	if wibctransfertypes.IsPurposeAck(ctx) {
		return nil
	}

	// Same thing applies if IBC fails due to timeout.
	if wibctransfertypes.IsPurposeTimeout(ctx) {
		return nil
	}

	if def.IsFeatureEnabled(types.Feature_whitelisting) && !def.HasAdminPrivileges(addr) {
		balance := k.bankKeeper.GetBalance(ctx, addr, def.Denom)
		whitelistedBalance := k.GetWhitelistedBalance(ctx, addr, def.Denom)

		finalBalance := balance.Amount.Add(amount)
		if finalBalance.GT(whitelistedBalance.Amount) {
			return sdkerrors.Wrapf(
				types.ErrWhitelistedLimitExceeded,
				"balance whitelisted for %s is not enough to receive %s, current whitelisted balance: %s",
				addr, sdk.NewCoin(def.Denom, amount), whitelistedBalance)
		}
	}

	if def.IsFeatureEnabled(types.Feature_block_smart_contracts) &&
		!def.HasAdminPrivileges(addr) &&
		cwasmtypes.IsReceivingSmartContract(ctx, addr.String()) {
		return sdkerrors.Wrapf(cosmoserrors.ErrUnauthorized, "transfers to smart contracts are disabled for %s", def.Denom)
	}

	return nil
}

func (k Keeper) isSymbolDuplicated(ctx sdk.Context, symbol string, issuer sdk.AccAddress) bool {
	compositeKey := types.CreateSymbolKey(issuer, symbol)
	rawBytes := ctx.KVStore(k.storeKey).Get(compositeKey)
	return rawBytes != nil
}

func (k Keeper) getDefinitions(
	ctx sdk.Context,
	pagination *query.PageRequest,
) ([]types.Definition, *query.PageResponse, error) {
	return k.getDefinitionsFromStore(prefix.NewStore(ctx.KVStore(k.storeKey), types.TokenKeyPrefix), pagination)
}

func (k Keeper) getIssuerDefinitions(
	ctx sdk.Context,
	issuer sdk.AccAddress,
	pagination *query.PageRequest,
) ([]types.Definition, *query.PageResponse, error) {
	return k.getDefinitionsFromStore(
		prefix.NewStore(ctx.KVStore(k.storeKey), types.CreateIssuerTokensPrefix(issuer)),
		pagination,
	)
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
		if unit.Denom == metadata.Display {
			precision = int(unit.Exponent)
			break
		}
	}

	if precision < 0 {
		return types.Token{}, sdkerrors.Wrap(types.ErrInvalidInput, "precision not found")
	}

	dexSettings, err := k.getDEXSettingsOrNil(ctx, definition.Denom)
	if err != nil {
		return types.Token{}, err
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
		Version:            definition.Version,
		URI:                definition.URI,
		URIHash:            definition.URIHash,
		Admin:              definition.Admin,
		ExtensionCWAddress: definition.ExtensionCWAddress,
		DEXSettings:        dexSettings,
	}, nil
}

func (k Keeper) getDefinitionsFromStore(
	store prefix.Store,
	pagination *query.PageRequest,
) ([]types.Definition, *query.PageResponse, error) {
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
		return nil, nil, sdkerrors.Wrapf(types.ErrInvalidInput, "failed to paginate: %s", err)
	}

	definitions := make([]types.Definition, 0, len(definitionsPointers))
	for _, definition := range definitionsPointers {
		definitions = append(definitions, *definition)
	}

	return definitions, pageRes, nil
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

func (k Keeper) freezingChecks(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error {
	def, err := k.GetDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	if def.HasAdminPrivileges(addr) {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "admin's balance can't be frozen")
	}

	return def.CheckFeatureAllowed(sender, types.Feature_freezing)
}

func (k Keeper) isGloballyFrozen(ctx sdk.Context, denom string) bool {
	return bytes.Equal(ctx.KVStore(k.storeKey).Get(types.CreateGlobalFreezeKey(denom)), types.StoreTrue)
}

func (k Keeper) validateClawbackAllowed(ctx sdk.Context, sender, addr sdk.AccAddress, coin sdk.Coin) error {
	def, err := k.GetDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	if _, isModuleAccount := k.accountKeeper.GetAccount(ctx, addr).(*authtypes.ModuleAccount); isModuleAccount {
		return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "claw back from module accounts is prohibited")
	}

	balance := k.bankKeeper.GetBalance(ctx, addr, coin.Denom)
	if err := k.validateCoinIsNotLockedByDEXAndBank(ctx, addr, balance, coin); err != nil {
		return err
	}

	return def.CheckFeatureAllowed(sender, types.Feature_clawback)
}

// whitelistedAccountBalanceStore gets the store for the whitelisted balances of an account.
func (k Keeper) whitelistedAccountBalanceStore(ctx sdk.Context, addr sdk.AccAddress) balanceStore {
	return newBalanceStore(k.cdc, ctx.KVStore(k.storeKey), types.CreateWhitelistedBalancesKey(addr))
}

// dexLockedBalancesStore get the store for the DEX locked balances of all accounts.
func (k Keeper) dexLockedBalancesStore(ctx sdk.Context) prefix.Store {
	return prefix.NewStore(ctx.KVStore(k.storeKey), types.DEXLockedBalancesKeyPrefix)
}

// dexLockedAccountBalanceStore gets the store for the DEX locked balances of an account.
func (k Keeper) dexLockedAccountBalanceStore(ctx sdk.Context, addr sdk.AccAddress) balanceStore {
	return newBalanceStore(k.cdc, ctx.KVStore(k.storeKey), types.CreateDEXLockedBalancesKey(addr))
}

func (k Keeper) dexLockingChecks(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error {
	def, err := k.GetDefinition(ctx, coin.Denom)
	if err != nil {
		// check if the token is assetft
		if !sdkerrors.IsOf(err, types.ErrInvalidDenom, types.ErrTokenNotFound) {
			return err
		}
	} else {
		if def.ExtensionCWAddress != "" {
			return sdkerrors.Wrapf(
				types.ErrInvalidInput,
				"failed to DEX lock %s, not supported for the tokens with extensions",
				coin.String(),
			)
		}
		if def.IsFeatureEnabled(types.Feature_block_dex) {
			return sdkerrors.Wrapf(
				cosmoserrors.ErrUnauthorized,
				"locking coins for DEX disabled for %s",
				def.Denom,
			)
		}
	}

	balance := k.bankKeeper.GetBalance(ctx, addr, coin.Denom)
	if err := k.validateCoinIsNotLockedByDEXAndBank(ctx, addr, balance, coin); err != nil {
		return err
	}

	frozenAmt := k.GetFrozenBalance(ctx, addr, coin.Denom).Amount
	notFrozenTotalAmt := balance.Amount.Sub(frozenAmt)
	if notFrozenTotalAmt.LT(coin.Amount) {
		return sdkerrors.Wrapf(
			cosmoserrors.ErrInsufficientFunds,
			"failed to DEX lock %s available %s%s",
			coin.String(),
			notFrozenTotalAmt,
			coin.Denom,
		)
	}

	return nil
}

func (k Keeper) validateCoinIsNotLockedByDEXAndBank(
	ctx sdk.Context,
	addr sdk.AccAddress,
	balance, coin sdk.Coin,
) error {
	dexLockedAmt := k.GetDEXLockedBalance(ctx, addr, coin.Denom).Amount
	availableAmt := balance.Amount.Sub(dexLockedAmt)
	if availableAmt.LT(coin.Amount) {
		return sdkerrors.Wrapf(cosmoserrors.ErrInsufficientFunds, "%s is not available, available %s%s",
			coin.String(), availableAmt.String(), coin.Denom)
	}

	bankLockedAmt := k.bankKeeper.LockedCoins(ctx, addr).AmountOf(coin.Denom)
	// validate that we don't use the coins locked by bank
	availableAmt = availableAmt.Sub(bankLockedAmt)
	if availableAmt.LT(coin.Amount) {
		return sdkerrors.Wrapf(cosmoserrors.ErrInsufficientFunds, "%s is not available, available %s%s",
			coin.String(), availableAmt.String(), coin.Denom)
	}

	return nil
}

func (k Keeper) getDEXSettingsOrNil(ctx sdk.Context, denom string) (*types.DEXSettings, error) {
	dexSettings, err := k.GetDEXSettings(ctx, denom)
	if err != nil {
		if errors.Is(err, types.ErrDEXSettingsNotFound) {
			return nil, nil //nolint:nilnil //returns nil if data not found
		}
		return nil, err
	}

	return &dexSettings, nil
}

// logger returns the Keeper logger.
func (k Keeper) logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
