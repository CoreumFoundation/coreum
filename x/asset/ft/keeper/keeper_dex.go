package keeper

import (
	"encoding/json"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
	cwasmtypes "github.com/CoreumFoundation/coreum/v6/x/wasm/types"
)

// ExtensionPlaceOrderMethod is the function name of the extension smart contract, which will be invoked
// when place and DEX order.
const ExtensionPlaceOrderMethod = "extension_place_order"

// sudoExtensionPlaceOrderMsg contains the fields passed to extension method call.
type sudoExtensionPlaceOrderMsg struct {
	Order    types.DEXOrder `json:"order"`
	Spent    sdk.Coin       `json:"spent"`
	Received sdk.Coin       `json:"received"`
}

// DEXExecuteActions executes a series of DEX actions which include checking order amounts,
// adjusting locked balances, and updating expected to receive balances. It performs necessary
// validations and updates the state accordingly based on the provided actions.
func (k Keeper) DEXExecuteActions(ctx sdk.Context, actions types.DEXActions) error {
	if err := k.DEXCheckOrderAmounts(
		ctx,
		actions.Order,
		actions.CreatorExpectedToSpend,
		actions.CreatorExpectedToReceive,
	); err != nil {
		return err
	}

	for _, lock := range actions.IncreaseLocked {
		if err := k.DEXIncreaseLocked(ctx, lock.Address, lock.Coin); err != nil {
			return err
		}
	}

	for _, unlock := range actions.DecreaseLocked {
		if err := k.DEXDecreaseLocked(ctx, unlock.Address, unlock.Coin); err != nil {
			return err
		}
	}

	for _, increase := range actions.IncreaseExpectedToReceive {
		if err := k.DEXIncreaseExpectedToReceive(ctx, increase.Address, increase.Coin); err != nil {
			return err
		}
	}

	for _, decrease := range actions.DecreaseExpectedToReceive {
		if err := k.DEXDecreaseExpectedToReceive(ctx, decrease.Address, decrease.Coin); err != nil {
			return err
		}
	}

	for _, send := range actions.Send {
		k.logger(ctx).Debug(
			"DEX sending coin",
			"from", send.FromAddress.String(),
			"to", send.ToAddress.String(),
			"coin", send.Coin.String(),
		)
		if err := k.bankKeeper.SendCoins(ctx, send.FromAddress, send.ToAddress, sdk.NewCoins(send.Coin)); err != nil {
			return sdkerrors.Wrap(err, "failed to DEX send coins")
		}
	}

	// The actual call to the smart contract should always be the last line, because when the wasm calls the contract,
	// it can return some messages that will be executed right away before finalizing order. It can be called
	// recursively either unlimited times, or inject orders that might match before finalizing original order which
	// causes errors in calculations and other attacks.
	return k.DEXInvokeAssetExtension(
		ctx,
		actions.Order,
		actions.CreatorExpectedToSpend,
		actions.CreatorExpectedToReceive,
	)
}

// DEXDecreaseLimits decreases the DEX limits.
func (k Keeper) DEXDecreaseLimits(
	ctx sdk.Context,
	addr sdk.AccAddress,
	lockedCoins sdk.Coins, expectedToReceiveCoin sdk.Coin,
) error {
	for _, coin := range lockedCoins {
		if err := k.DEXDecreaseLocked(ctx, addr, coin); err != nil {
			return err
		}
	}

	return k.DEXDecreaseExpectedToReceive(ctx, addr, expectedToReceiveCoin)
}

// DEXCheckOrderAmounts validates that the order's creator is allowed to place and order with the provided amounts.
func (k Keeper) DEXCheckOrderAmounts(
	ctx sdk.Context,
	order types.DEXOrder,
	expectedToSpend, expectedToReceive sdk.Coin,
) error {
	if err := k.dexCheckExpectedToSpend(ctx, order, expectedToSpend, expectedToReceive); err != nil {
		return err
	}

	return k.dexCheckExpectedToReceive(ctx, order, expectedToSpend, expectedToReceive)
}

// DEXInvokeAssetExtension invokes the smart contract place order.
func (k Keeper) DEXInvokeAssetExtension(ctx sdk.Context, order types.DEXOrder, spent, received sdk.Coin) error {
	if err := k.dexInvokeAssetExtensionWithAmount(ctx, order, spent, received, spent.Denom); err != nil {
		return err
	}

	return k.dexInvokeAssetExtensionWithAmount(ctx, order, spent, received, received.Denom)
}

// SetDEXSettings sets the DEX settings of a specified denom.
func (k Keeper) SetDEXSettings(ctx sdk.Context, denom string, settings types.DEXSettings) error {
	return k.storeService.OpenKVStore(ctx).Set(types.CreateDEXSettingsKey(denom), k.cdc.MustMarshal(&settings))
}

// GetDEXSettings gets the DEX settings of a specified denom.
func (k Keeper) GetDEXSettings(ctx sdk.Context, denom string) (types.DEXSettings, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.CreateDEXSettingsKey(denom))
	if err != nil {
		return types.DEXSettings{}, err
	}
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
	moduleStore := k.storeService.OpenKVStore(ctx)
	dexSettings := make([]types.DEXSettingsWithDenom, 0)
	store := prefix.NewStore(runtime.KVStoreAdapter(moduleStore), types.DEXSettingsKeyPrefix)
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

// UpdateDEXUnifiedRefAmount updates the DEX unified ref amount .
func (k Keeper) UpdateDEXUnifiedRefAmount(
	ctx sdk.Context,
	sender sdk.AccAddress,
	denom string,
	unifiedRefAmount sdkmath.LegacyDec,
) error {
	return k.updateDEXSettings(ctx, sender, denom, types.DEXSettings{UnifiedRefAmount: &unifiedRefAmount})
}

// UpdateDEXWhitelistedDenoms updates the DEX whitelisted denoms of a specified denom.
func (k Keeper) UpdateDEXWhitelistedDenoms(
	ctx sdk.Context,
	sender sdk.AccAddress,
	denom string,
	whitelistedDenoms []string,
) error {
	if whitelistedDenoms == nil {
		// check to prevent mistakes using the `updateDEXSettings` method, set to empty slice if the input is nil
		whitelistedDenoms = make([]string, 0)
	}
	return k.updateDEXSettings(ctx, sender, denom, types.DEXSettings{WhitelistedDenoms: whitelistedDenoms})
}

// DEXIncreaseExpectedToReceive increases the expected to receive amount for the specified account.
func (k Keeper) DEXIncreaseExpectedToReceive(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error {
	k.logger(ctx).Debug("DEX increasing expected to receive coin", "address", addr.String(), "coin", coin.String())
	if !coin.IsPositive() {
		return sdkerrors.Wrap(
			cosmoserrors.ErrInvalidCoins, "amount to increase DEX expected to receive must be positive",
		)
	}

	shouldRecord, err := k.shouldRecordDEXExpectedToReceiveBalance(ctx, coin.Denom)
	if err != nil {
		return err
	}
	if !shouldRecord {
		return nil
	}

	dexExpectedToReceiveStore := k.dexExpectedToReceiveAccountBalanceStore(ctx, addr)
	prevExpectedToReceiveBalance, newExpectedToReceiveBalance := dexExpectedToReceiveStore.AddBalance(coin)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventDEXExpectedToReceiveAmountChanged{
		Account:        addr.String(),
		Denom:          coin.Denom,
		PreviousAmount: prevExpectedToReceiveBalance.Amount,
		CurrentAmount:  newExpectedToReceiveBalance.Amount,
	}); err != nil {
		return sdkerrors.Wrapf(
			types.ErrInvalidState, "failed to emit EventDEXExpectedToReceiveAmountChanged event: %s", err,
		)
	}

	return nil
}

// DEXDecreaseExpectedToReceive decreases the expected to receive amount for the specified account.
func (k Keeper) DEXDecreaseExpectedToReceive(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error {
	k.logger(ctx).Debug("DEX decreasing expected to receive coin", "address", addr.String(), "coin", coin.String())
	if !coin.IsPositive() {
		return sdkerrors.Wrap(
			cosmoserrors.ErrInvalidCoins, "amount to decrease DEX expected to receive must be positive",
		)
	}

	shouldRecord, err := k.shouldRecordDEXExpectedToReceiveBalance(ctx, coin.Denom)
	if err != nil {
		return err
	}
	if !shouldRecord {
		return nil
	}

	dexExpectedToReceiveStore := k.dexExpectedToReceiveAccountBalanceStore(ctx, addr)
	prevExpectedToReceiveBalance, newExpectedToReceiveBalance, err := dexExpectedToReceiveStore.SubBalance(coin)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to cancel DEX whitelisted")
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventDEXExpectedToReceiveAmountChanged{
		Account:        addr.String(),
		Denom:          coin.Denom,
		PreviousAmount: prevExpectedToReceiveBalance.Amount,
		CurrentAmount:  newExpectedToReceiveBalance.Amount,
	}); err != nil {
		return sdkerrors.Wrapf(
			types.ErrInvalidState, "failed to emit EventDEXExpectedToReceiveAmountChanged event: %s", err,
		)
	}

	return nil
}

// GetDEXExpectedToReceivedBalance returns the DEX expected to receive balance.
func (k Keeper) GetDEXExpectedToReceivedBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	return k.dexExpectedToReceiveAccountBalanceStore(ctx, addr).Balance(denom)
}

// GetDEXExpectedToReceiveBalances returns the DEX expected to receive balances of an account.
func (k Keeper) GetDEXExpectedToReceiveBalances(
	ctx sdk.Context,
	addr sdk.AccAddress,
	pagination *query.PageRequest,
) (sdk.Coins, *query.PageResponse, error) {
	return k.dexExpectedToReceiveAccountBalanceStore(ctx, addr).Balances(pagination)
}

// GetAccountsDEXExpectedToReceiveBalances returns the DEX expected to receive balance on all the account.
func (k Keeper) GetAccountsDEXExpectedToReceiveBalances(
	ctx sdk.Context,
	pagination *query.PageRequest,
) ([]types.Balance, *query.PageResponse, error) {
	return collectBalances(k.cdc, k.dexExpectedToReceiveBalancesStore(ctx), pagination)
}

// SetDEXExpectedToReceiveBalances sets the DEX expected to receive balances of a specified account.
// Pay attention that the sdk.NewCoins() sanitizes/removes the empty coins, hence if you
// need set zero amount use the slice []sdk.Coins.
func (k Keeper) SetDEXExpectedToReceiveBalances(ctx sdk.Context, addr sdk.AccAddress, coins sdk.Coins) {
	dexExpectedToReceiveStore := k.dexExpectedToReceiveAccountBalanceStore(ctx, addr)
	for _, coin := range coins {
		dexExpectedToReceiveStore.SetBalance(coin)
	}
}

// DEXIncreaseLocked locks specified token for the specified account.
func (k Keeper) DEXIncreaseLocked(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error {
	k.logger(ctx).Debug("DEX increasing locked coin", "addr", addr.String(), "coin", coin.String())
	if !coin.IsPositive() {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidCoins, "amount to lock DEX tokens must be positive")
	}

	if err := k.validateCoinIsNotLockedByDEXAndBank(ctx, addr, coin); err != nil {
		return sdkerrors.Wrapf(types.ErrDEXInsufficientSpendableBalance, "%s", err)
	}

	dexLockedStore := k.dexLockedAccountBalanceStore(ctx, addr)
	prevLockedBalance, newLockedBalance := dexLockedStore.AddBalance(coin)

	if err := ctx.EventManager().EmitTypedEvent(&types.EventDEXLockedAmountChanged{
		Account:        addr.String(),
		Denom:          coin.Denom,
		PreviousAmount: prevLockedBalance.Amount,
		CurrentAmount:  newLockedBalance.Amount,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit EventDEXLockedAmountChanged event: %s", err)
	}

	return nil
}

// DEXDecreaseLocked unlocks specified tokens from the specified account.
func (k Keeper) DEXDecreaseLocked(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) error {
	k.logger(ctx).Debug("DEX decrease locked coin", "address", addr.String(), "coin", coin.String())
	if !coin.IsPositive() {
		return sdkerrors.Wrap(cosmoserrors.ErrInvalidCoins, "amount to unlock DEX tokens must be positive")
	}

	dexLockedStore := k.dexLockedAccountBalanceStore(ctx, addr)
	prevLockedBalance, newLockedBalance, err := dexLockedStore.SubBalance(coin)
	if err != nil {
		return sdkerrors.Wrap(err, "failed to unlock DEX")
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventDEXLockedAmountChanged{
		Account:        addr.String(),
		Denom:          coin.Denom,
		PreviousAmount: prevLockedBalance.Amount,
		CurrentAmount:  newLockedBalance.Amount,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit EventDEXLockedAmountChanged event: %s", err)
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

// ValidateDEXCancelOrdersByDenomIsAllowed validates whether the cancellation of  orders by denom is allowed.
func (k Keeper) ValidateDEXCancelOrdersByDenomIsAllowed(ctx sdk.Context, addr sdk.AccAddress, denom string) error {
	def, err := k.GetDefinition(ctx, denom)
	if err != nil {
		return err
	}

	if !def.HasAdminPrivileges(addr) {
		return sdkerrors.Wrapf(cosmoserrors.ErrUnauthorized, "only admin is able to cancel orders by denom %s", denom)
	}
	if !def.IsFeatureEnabled(types.Feature_dex_order_cancellation) {
		return sdkerrors.Wrapf(
			cosmoserrors.ErrUnauthorized,
			"order cancellation is not allowed by denom %s, feature %s is disabled",
			denom, types.Feature_dex_order_cancellation,
		)
	}

	return nil
}

func (k Keeper) dexCheckExpectedToSpend(
	ctx sdk.Context,
	order types.DEXOrder,
	expectedToSpend, expectedToReceive sdk.Coin,
) error {
	// validate that the order creator has enough balance, for both extension and non-extension coin
	if err := k.validateCoinIsNotLockedByDEXAndBank(ctx, order.Creator, expectedToSpend); err != nil {
		return sdkerrors.Wrapf(types.ErrDEXInsufficientSpendableBalance, "%s", err)
	}

	spendDef, err := k.getDefinitionOrNil(ctx, expectedToSpend.Denom)
	if err != nil {
		return err
	}

	if spendDef == nil {
		return nil
	}

	if err = k.dexChecksForDenom(ctx, order.Creator, spendDef, expectedToReceive.Denom); err != nil {
		return err
	}

	if err = k.validateCoinSpendable(ctx, order.Creator, *spendDef, expectedToSpend.Amount); err != nil {
		return sdkerrors.Wrapf(types.ErrDEXInsufficientSpendableBalance, "err: %s", err)
	}

	return nil
}

func (k Keeper) dexCheckExpectedToReceive(
	ctx sdk.Context,
	order types.DEXOrder,
	expectedToSpend, expectedToReceive sdk.Coin,
) error {
	receiveDef, err := k.getDefinitionOrNil(ctx, expectedToReceive.Denom)
	if err != nil {
		return err
	}
	if receiveDef == nil {
		return nil
	}

	if err = k.dexChecksForDenom(ctx, order.Creator, receiveDef, expectedToSpend.Denom); err != nil {
		return err
	}

	return k.validateCoinReceivable(ctx, order.Creator, *receiveDef, expectedToReceive.Amount)
}

func (k Keeper) dexInvokeAssetExtensionWithAmount(
	ctx sdk.Context,
	order types.DEXOrder,
	spent, received sdk.Coin,
	assetExtensionDenom string,
) error {
	receiveDef, err := k.getDefinitionOrNil(ctx, assetExtensionDenom)
	if err != nil {
		return err
	}
	if receiveDef == nil {
		return nil
	}

	if receiveDef.IsFeatureEnabled(types.Feature_extension) {
		extensionContract, err := sdk.AccAddressFromBech32(receiveDef.ExtensionCWAddress)
		if err != nil {
			return err
		}
		return k.invokeAssetExtensionPlaceOrderMethod(
			ctx, extensionContract, order, spent, received,
		)
	}

	return nil
}

func (k Keeper) invokeAssetExtensionPlaceOrderMethod(
	ctx sdk.Context,
	extensionContract sdk.AccAddress,
	order types.DEXOrder,
	spent, received sdk.Coin,
) error {
	contractMsg := map[string]interface{}{
		ExtensionPlaceOrderMethod: sudoExtensionPlaceOrderMsg{
			Order:    order,
			Spent:    spent,
			Received: received,
		},
	}
	contractMsgBytes, err := json.Marshal(contractMsg)
	if err != nil {
		return sdkerrors.Wrapf(err, "failed to marshal contract msg")
	}

	_, err = k.wasmPermissionedKeeper.Sudo(
		ctx,
		extensionContract,
		contractMsgBytes,
	)
	if err != nil {
		return types.ErrExtensionCallFailed.Wrapf("wasm error: %s", err)
	}

	return nil
}

func (k Keeper) dexChecksForDenom(
	ctx sdk.Context,
	acc sdk.AccAddress,
	def *types.Definition, oppositeDenom string,
) error {
	if def == nil {
		return nil
	}

	if err := k.dexChecksForDefinition(ctx, acc, *def); err != nil {
		return err
	}

	// settings specific validation
	settings, err := k.getDEXSettingsOrNil(ctx, def.Denom)
	if err != nil {
		return err
	}

	if settings != nil {
		// validate whitelisted denoms
		if len(settings.WhitelistedDenoms) == 0 {
			return nil
		}
		if !lo.Contains(settings.WhitelistedDenoms, oppositeDenom) {
			return sdkerrors.Wrapf(
				cosmoserrors.ErrUnauthorized,
				"locking coins for DEX is prohibited, denom %s not whitelisted for %s",
				oppositeDenom, def.Denom,
			)
		}
	}

	return nil
}

func (k Keeper) dexChecksForDefinition(ctx sdk.Context, acc sdk.AccAddress, def types.Definition) error {
	if def.IsFeatureEnabled(types.Feature_dex_block) {
		return sdkerrors.Wrapf(
			cosmoserrors.ErrUnauthorized,
			"usage of %s is not supported for DEX, the token has %s feature enabled",
			def.Denom, types.Feature_dex_block.String(),
		)
	}

	// don't allow the smart contract to use the denom with Feature_block_smart_contracts if not admin
	if def.IsFeatureEnabled(types.Feature_block_smart_contracts) &&
		!def.HasAdminPrivileges(acc) &&
		cwasmtypes.IsTriggeredBySmartContract(ctx) {
		return sdkerrors.Wrapf(
			cosmoserrors.ErrUnauthorized,
			"usage of %s is not supported for DEX in smart contract, the token has %s feature enabled",
			def.Denom, types.Feature_block_smart_contracts.String(),
		)
	}

	if def.IsFeatureEnabled(types.Feature_freezing) {
		isGloballyFrozen, err := k.isGloballyFrozen(ctx, def.Denom)
		if err != nil {
			return err
		}
		if isGloballyFrozen &&
			// sill allow the admin to do the trade, to follow same logic as we have in the sending
			!def.HasAdminPrivileges(acc) {
			return sdkerrors.Wrapf(
				cosmoserrors.ErrUnauthorized,
				"usage of %s for DEX is blocked because the token is globally frozen",
				def.Denom,
			)
		}
	}

	return nil
}

func (k Keeper) updateDEXSettings(
	ctx sdk.Context,
	sender sdk.AccAddress,
	denom string,
	settings types.DEXSettings,
) error {
	prevSettings, err := k.getDEXSettingsOrNil(ctx, denom)
	if err != nil {
		return err
	}
	if prevSettings == nil {
		prevSettings = &types.DEXSettings{}
	}

	newSettings := *prevSettings
	// update not nil settings
	if settings.WhitelistedDenoms != nil {
		newSettings.WhitelistedDenoms = settings.WhitelistedDenoms
	}
	if settings.UnifiedRefAmount != nil {
		newSettings.UnifiedRefAmount = settings.UnifiedRefAmount
	}

	if err := types.ValidateDEXSettings(settings); err != nil {
		return err
	}

	def, err := k.getDefinitionOrNil(ctx, denom)
	if err != nil {
		return err
	}
	// the gov can update any DEX setting even if the features are disabled
	if k.authority != sender.String() { //nolint:nestif // the ifs are for the error checks mostly
		if def != nil {
			if !def.IsAdmin(sender) {
				return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "only admin and gov can update DEX settings")
			}
			if err := types.ValidateDEXSettingsAccess(newSettings, *def); err != nil {
				return err
			}
		} else {
			return sdkerrors.Wrap(cosmoserrors.ErrUnauthorized, "only admin or gov can update DEX settings")
		}
	}

	if err = k.SetDEXSettings(ctx, denom, newSettings); err != nil {
		return err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventDEXSettingsChanged{
		PreviousSettings: prevSettings,
		NewSettings:      newSettings,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to emit EventDEXSettingsChanged event: %s", err)
	}

	return nil
}

func (k Keeper) validateCoinIsNotLockedByDEXAndBank(
	ctx sdk.Context,
	addr sdk.AccAddress,
	coin sdk.Coin,
) error {
	balance := k.bankKeeper.GetBalance(ctx, addr, coin.Denom)
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

// dexExpectedToReceiveBalancesStore get the store for the DEX expected to receive balances of all accounts.
func (k Keeper) dexExpectedToReceiveBalancesStore(ctx sdk.Context) prefix.Store {
	store := k.storeService.OpenKVStore(ctx)
	return prefix.NewStore(runtime.KVStoreAdapter(store), types.DEXExpectedToReceiveBalancesKeyPrefix)
}

// dexExpectedToReceiveAccountBalanceStore gets the store for the DEX expected to receive balances of an account.
func (k Keeper) dexExpectedToReceiveAccountBalanceStore(ctx sdk.Context, addr sdk.AccAddress) balanceStore {
	store := k.storeService.OpenKVStore(ctx)
	return newBalanceStore(k.cdc, runtime.KVStoreAdapter(store), types.CreateDEXExpectedToReceiveBalancesKey(addr))
}

func (k Keeper) shouldRecordDEXExpectedToReceiveBalance(ctx sdk.Context, denom string) (bool, error) {
	def, err := k.getDefinitionOrNil(ctx, denom)
	if err != nil {
		return false, err
	}
	// increase for FT with the whitelisting enabled only
	if def != nil && def.IsFeatureEnabled(types.Feature_whitelisting) {
		return true, nil
	}

	return false, nil
}

// dexLockedBalancesStore get the store for the DEX locked balances of all accounts.
func (k Keeper) dexLockedBalancesStore(ctx sdk.Context) prefix.Store {
	store := k.storeService.OpenKVStore(ctx)
	return prefix.NewStore(runtime.KVStoreAdapter(store), types.DEXLockedBalancesKeyPrefix)
}

// dexLockedAccountBalanceStore gets the store for the DEX locked balances of an account.
func (k Keeper) dexLockedAccountBalanceStore(ctx sdk.Context, addr sdk.AccAddress) balanceStore {
	store := k.storeService.OpenKVStore(ctx)
	return newBalanceStore(k.cdc, runtime.KVStoreAdapter(store), types.CreateDEXLockedBalancesKey(addr))
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
