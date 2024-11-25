package keeper

import (
	"math/big"

	sdkstore "cosmossdk.io/core/store"
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	"cosmossdk.io/store/prefix"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/pkg/errors"
	"github.com/samber/lo"

	cbig "github.com/CoreumFoundation/coreum/v5/pkg/math/big"
	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

// Keeper is the dex module keeper.
type Keeper struct {
	cdc                codec.BinaryCodec
	storeService       sdkstore.KVStoreService
	accountKeeper      types.AccountKeeper
	accountQueryServer types.AccountQueryServer
	assetFTKeeper      types.AssetFTKeeper
	delayKeeper        types.DelayKeeper
	authority          string
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeService sdkstore.KVStoreService,
	accountKeeper types.AccountKeeper,
	accountQueryServer types.AccountQueryServer,
	assetFTKeeper types.AssetFTKeeper,
	delayKeeper types.DelayKeeper,
	authority string,
) Keeper {
	return Keeper{
		cdc:                cdc,
		storeService:       storeService,
		accountKeeper:      accountKeeper,
		accountQueryServer: accountQueryServer,
		assetFTKeeper:      assetFTKeeper,
		authority:          authority,
		delayKeeper:        delayKeeper,
	}
}

// PlaceOrder places an order on the corresponding order book, and matches the order.
func (k Keeper) PlaceOrder(ctx sdk.Context, order types.Order) error {
	k.logger(ctx).Debug("Placing order.", "order", order)

	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}
	if err := k.validateOrder(ctx, params, order); err != nil {
		return err
	}

	creator, err := sdk.AccAddressFromBech32(order.Creator)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", order.Creator)
	}

	accNumber, err := k.getAccountNumber(ctx, creator)
	if err != nil {
		return err
	}

	// validate duplicated order ID
	_, err = k.getOrderSequenceByID(ctx, accNumber, order.ID)
	if err != nil {
		if !sdkerrors.IsOf(err, types.ErrRecordNotFound) {
			return err
		}
	} else {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "order with the id %q is already created", order.ID)
	}

	orderBookID, oppositeOrderBookID, err := k.getOrGenOrderBookIDs(ctx, order.BaseDenom, order.QuoteDenom)
	if err != nil {
		return err
	}

	return k.matchOrder(ctx, params, accNumber, orderBookID, oppositeOrderBookID, order)
}

// CancelOrder cancels order and unlock locked balance.
func (k Keeper) CancelOrder(ctx sdk.Context, acc sdk.AccAddress, orderID string) error {
	return k.cancelOrder(ctx, acc, orderID)
}

// CancelOrderBySequence cancels order and unlock locked balance by order sequence.
func (k Keeper) CancelOrderBySequence(ctx sdk.Context, acc sdk.AccAddress, orderSequence uint64) error {
	return k.cancelOrderBySequence(ctx, acc, orderSequence)
}

// CancelOrdersByDenom cancels all orders of specified denom.
func (k Keeper) CancelOrdersByDenom(ctx sdk.Context, admin, acc sdk.AccAddress, denom string) error {
	if err := k.assetFTKeeper.ValidateDEXCancelOrdersByDenomIsAllowed(ctx, admin, denom); err != nil {
		return err
	}

	accNumber, err := k.getAccountNumber(ctx, acc)
	if err != nil {
		return err
	}
	accountDenomKeyPrefix, err := types.CreateAccountDenomKeyPrefix(accNumber, denom)
	if err != nil {
		return err
	}

	store := k.storeService.OpenKVStore(ctx)

	iterator := prefix.NewStore(runtime.KVStoreAdapter(store), accountDenomKeyPrefix).Iterator(nil, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		orderSequence, err := types.DecodeAccountDenomKeyOrderSequence(iterator.Key())
		if err != nil {
			return err
		}
		if err := k.cancelOrderBySequence(ctx, acc, orderSequence); err != nil {
			return err
		}
	}

	return nil
}

// GetOrderBookIDByDenoms returns order book ID by it's denoms.
func (k Keeper) GetOrderBookIDByDenoms(ctx sdk.Context, baseDenom, quoteDenom string) (uint32, error) {
	return k.getOrderBookIDByDenoms(ctx, baseDenom, quoteDenom)
}

// GetOrderByAddressAndID returns order by holder address and it's ID.
func (k Keeper) GetOrderByAddressAndID(ctx sdk.Context, acc sdk.AccAddress, orderID string) (types.Order, error) {
	order, _, err := k.getOrderWithRecordByAddressAndID(ctx, acc, orderID)
	if err != nil {
		return types.Order{}, err
	}

	return order, nil
}

// GetOrders returns creator orders.
func (k Keeper) GetOrders(
	ctx sdk.Context,
	creator sdk.AccAddress,
	pagination *query.PageRequest,
) ([]types.Order, *query.PageResponse, error) {
	return k.getPaginatedOrders(ctx, creator, pagination)
}

// GetOrderBooks returns order books.
func (k Keeper) GetOrderBooks(
	ctx sdk.Context,
	pagination *query.PageRequest,
) ([]types.OrderBookData, *query.PageResponse, error) {
	orderBookWithIDs, pageRes, err := k.getPaginatedOrderBooksWithID(ctx, pagination)
	if err != nil {
		return nil, nil, err
	}
	return lo.Map(orderBookWithIDs, func(orderBookWithID types.OrderBookDataWithID, _ int) types.OrderBookData {
		return orderBookWithID.Data
	}), pageRes, nil
}

// GetOrderBookOrders returns order book records sorted by price asc. For the buy side it's expected to use the reverse
// pagination, and sort the orders by the order sequence asc additionally on the client side.
func (k Keeper) GetOrderBookOrders(
	ctx sdk.Context,
	baseDenom, quoteDenom string,
	side types.Side,
	pagination *query.PageRequest,
) ([]types.Order, *query.PageResponse, error) {
	if err := side.Validate(); err != nil {
		return nil, nil, err
	}

	return k.getPaginatedOrderBookOrders(ctx, baseDenom, quoteDenom, side, pagination)
}

// GetParams gets the parameters of the module.
func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	bz, err := k.storeService.OpenKVStore(ctx).Get(types.ParamsKey)
	if err != nil {
		return types.Params{}, err
	}
	var params types.Params
	k.cdc.MustUnmarshal(bz, &params)
	return params, nil
}

// UpdateParams is a governance operation that sets parameters of the module.
func (k Keeper) UpdateParams(ctx sdk.Context, authority string, params types.Params) error {
	if k.authority != authority {
		return sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.authority, authority)
	}
	if err := params.ValidateBasic(); err != nil {
		return err
	}

	return k.SetParams(ctx, params)
}

// SetParams sets the parameters of the module.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Set(types.ParamsKey, bz)
}

// GetAccountsOrders returns paginated orders.
func (k Keeper) GetAccountsOrders(
	ctx sdk.Context,
	pagination *query.PageRequest,
) ([]types.Order, *query.PageResponse, error) {
	moduleStore := k.storeService.OpenKVStore(ctx)
	store := prefix.NewStore(runtime.KVStoreAdapter(moduleStore), types.OrderIDToSequenceKeyPrefix)
	orderBookIDToOrderBookData := make(map[uint32]types.OrderBookData)
	accNumberToAddCache := make(map[uint64]sdk.AccAddress)
	orders, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
		pagination,
		// builder
		func(key []byte, record *gogotypes.UInt64Value) (*types.Order, error) {
			accNumber, _, err := types.DecodeOrderIDToSequenceKey(key)
			if err != nil {
				return nil, err
			}

			var acc sdk.AccAddress
			acc, err = k.getAccountAddressWithCache(ctx, accNumber, accNumberToAddCache)
			if err != nil {
				return nil, err
			}

			orderSequence := record.Value
			orderData, err := k.getOrderData(ctx, orderSequence)
			if err != nil {
				return nil, err
			}

			orderBookID := orderData.OrderBookID
			orderBookData, ok := orderBookIDToOrderBookData[orderBookID]
			if !ok {
				orderBookData, err = k.getOrderBookData(ctx, orderBookID)
				if err != nil {
					return nil, err
				}
				orderBookIDToOrderBookData[orderBookID] = orderBookData
			}

			orderBookRecord, err := k.getOrderBookRecord(
				ctx,
				orderData.OrderBookID,
				orderData.Side,
				orderData.Price,
				orderSequence,
			)
			if err != nil {
				return nil, err
			}

			return &types.Order{
				Creator:           acc.String(),
				Type:              types.ORDER_TYPE_LIMIT,
				ID:                orderBookRecord.OrderID,
				Sequence:          orderSequence,
				BaseDenom:         orderBookData.BaseDenom,
				QuoteDenom:        orderBookData.QuoteDenom,
				Price:             &orderData.Price,
				Quantity:          orderData.Quantity,
				Side:              orderData.Side,
				GoodTil:           orderData.GoodTil,
				TimeInForce:       types.TIME_IN_FORCE_GTC,
				RemainingQuantity: orderBookRecord.RemainingQuantity,
				RemainingBalance:  orderBookRecord.RemainingBalance,
				Reserve:           orderData.Reserve,
			}, nil
		},
		// constructor
		func() *gogotypes.UInt64Value {
			return &gogotypes.UInt64Value{}
		},
	)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(types.ErrInvalidInput, "failed to paginate: %s", err)
	}
	return lo.Map(orders, func(order *types.Order, _ int) types.Order {
		return *order
	}), pageRes, nil
}

// GetOrderBooksWithID returns order books with IDs.
func (k Keeper) GetOrderBooksWithID(
	ctx sdk.Context,
	pagination *query.PageRequest,
) ([]types.OrderBookDataWithID, *query.PageResponse, error) {
	return k.getPaginatedOrderBooksWithID(ctx, pagination)
}

// SaveOrderBookIDWithData saves order book ID with corresponding data.
func (k Keeper) SaveOrderBookIDWithData(ctx sdk.Context, orderBookID uint32, data types.OrderBookData) error {
	return k.saveOrderBookIDWithData(ctx, orderBookID, data.BaseDenom, data.QuoteDenom)
}

// GetOrderSequence returns current order sequence.
func (k Keeper) GetOrderSequence(ctx sdk.Context) (uint64, error) {
	return k.getUint64Value(ctx, types.OrderSequenceKey)
}

// SetOrderSequence sets order sequence.
func (k Keeper) SetOrderSequence(ctx sdk.Context, sequence uint64) error {
	return k.setUint64Value(ctx, types.OrderSequenceKey, sequence)
}

// SetOrderBookSequence sets order book sequence.
func (k Keeper) SetOrderBookSequence(ctx sdk.Context, sequence uint32) error {
	return k.setUint32Value(ctx, types.OrderBookSequenceKey, sequence)
}

// SaveOrderWithOrderBookRecord saves order with order book record.
func (k Keeper) SaveOrderWithOrderBookRecord(
	ctx sdk.Context,
	order types.Order,
	record types.OrderBookRecord,
) error {
	return k.saveOrderWithOrderBookRecord(ctx, order, record)
}

// GetAccountDenomOrdersCount returns account's denom orders count.
func (k Keeper) GetAccountDenomOrdersCount(
	ctx sdk.Context,
	acc sdk.AccAddress,
	denom string,
) (uint64, error) {
	accNumber, err := k.getAccountNumber(ctx, acc)
	if err != nil {
		return 0, err
	}

	return k.getAccountDenomOrdersCounter(ctx, accNumber, denom)
}

// GetAccountsDenomsOrdersCounts returns accounts denoms orders count.
func (k Keeper) GetAccountsDenomsOrdersCounts(
	ctx sdk.Context,
	pagination *query.PageRequest,
) ([]types.AccountDenomOrdersCount, *query.PageResponse, error) {
	moduleStore := k.storeService.OpenKVStore(ctx)
	store := prefix.NewStore(runtime.KVStoreAdapter(moduleStore), types.AccountDenomOrdersCountKeyPrefix)
	counts, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
		pagination,
		// builder
		func(key []byte, record *gogotypes.UInt64Value) (*types.AccountDenomOrdersCount, error) {
			accNumber, denom, err := types.DecodeAccountDenomOrdersCountKey(key)
			if err != nil {
				return nil, err
			}

			return &types.AccountDenomOrdersCount{
				AccountNumber: accNumber,
				Denom:         denom,
				OrdersCount:   record.Value,
			}, nil
		},
		// constructor
		func() *gogotypes.UInt64Value {
			return &gogotypes.UInt64Value{}
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return lo.Map(counts, func(c *types.AccountDenomOrdersCount, _ int) types.AccountDenomOrdersCount {
		return *c
	}), pageRes, nil
}

// SetAccountDenomOrdersCount sets accounts denoms orders count.
func (k Keeper) SetAccountDenomOrdersCount(
	ctx sdk.Context,
	accountDenomOrdersCount types.AccountDenomOrdersCount,
) error {
	return k.setAccountDenomOrdersCount(ctx, accountDenomOrdersCount)
}

func (k Keeper) validatePriceTick(
	ctx sdk.Context,
	params types.Params,
	baseDenom, quoteDenom string,
	price types.Price,
) error {
	baseDenomRefAmount, buyRefAmountFound, err := k.getAssetFTUnifiedRefAmount(ctx, baseDenom)
	if err != nil {
		return err
	}

	quoteDenomRefAmount, sellRefAmountFound, err := k.getAssetFTUnifiedRefAmount(ctx, quoteDenom)
	if err != nil {
		return err
	}

	if !buyRefAmountFound {
		baseDenomRefAmount = params.DefaultUnifiedRefAmount
	}
	if !sellRefAmountFound {
		quoteDenomRefAmount = params.DefaultUnifiedRefAmount
	}

	priceTickRat := ComputePriceTick(baseDenomRefAmount, quoteDenomRefAmount, params.PriceTickExponent)
	_, remainder := cbig.RatQuoWithIntRemainder(price.Rat(), priceTickRat)
	if !cbig.IntEqZero(remainder) {
		return sdkerrors.Wrapf(
			types.ErrInvalidInput,
			"invalid price %s, the price must be multiple of %s",
			price.Rat().String(), priceTickRat.String(),
		)
	}

	return nil
}

func (k Keeper) validateOrder(ctx sdk.Context, params types.Params, order types.Order) error {
	if err := order.Validate(); err != nil {
		return err
	}

	// price
	if order.Type == types.ORDER_TYPE_LIMIT {
		if err := k.validatePriceTick(ctx, params, order.BaseDenom, order.QuoteDenom, *order.Price); err != nil {
			return err
		}
	}

	// good til
	if order.GoodTil != nil {
		if err := k.validateGoodTil(ctx, order); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) validateGoodTil(ctx sdk.Context, order types.Order) error {
	if order.GoodTil.GoodTilBlockHeight > 0 {
		currentHeight := ctx.BlockHeight()
		if order.GoodTil.GoodTilBlockHeight <= uint64(currentHeight) {
			return sdkerrors.Wrapf(
				types.ErrInvalidInput,
				"good til block height %d must be greater than current block height %d",
				order.GoodTil.GoodTilBlockHeight, currentHeight,
			)
		}
	}
	if order.GoodTil.GoodTilBlockTime != nil {
		currentTime := ctx.BlockTime()
		if !order.GoodTil.GoodTilBlockTime.After(currentTime) {
			return sdkerrors.Wrapf(
				types.ErrInvalidInput,
				"good til block height %s must be greater than current block height %s",
				order.GoodTil.GoodTilBlockTime, currentTime,
			)
		}
	}

	return nil
}

func (k Keeper) getAssetFTUnifiedRefAmount(ctx sdk.Context, denom string) (sdkmath.LegacyDec, bool, error) {
	settings, err := k.assetFTKeeper.GetDEXSettings(ctx, denom)
	if err != nil {
		if !errors.Is(err, assetfttypes.ErrDEXSettingsNotFound) {
			return sdkmath.LegacyDec{}, false, err
		}
		return sdkmath.LegacyDec{}, false, nil
	}
	if settings.UnifiedRefAmount == nil {
		return sdkmath.LegacyDec{}, false, nil
	}

	return *settings.UnifiedRefAmount, true, nil
}

func (k Keeper) getOrderBookIDByDenoms(ctx sdk.Context, baseDenom, quoteDenom string) (uint32, error) {
	orderBookIDKey, err := types.CreateOrderBookKey(baseDenom, quoteDenom)
	if err != nil {
		return 0, err
	}

	orderBookID, err := k.getOrderBookIDByKey(ctx, orderBookIDKey)
	if err != nil {
		return 0, sdkerrors.Wrapf(err, "faild to get order book ID, baseDenom: %s, quoteDenom: %s", baseDenom, quoteDenom)
	}

	return orderBookID, nil
}

func (k Keeper) getOrGenOrderBookIDs(ctx sdk.Context, baseDenom, quoteDenom string) (uint32, uint32, error) {
	// the function optimizes the read, by writing ordered denoms
	var denom0, denom1 string
	if baseDenom < quoteDenom {
		denom0 = baseDenom
		denom1 = quoteDenom
	} else {
		denom0 = quoteDenom
		denom1 = baseDenom
	}

	key0, err := types.CreateOrderBookKey(denom0, denom1)
	if err != nil {
		return 0, 0, err
	}
	orderBookID0, err := k.getOrderBookIDByKey(ctx, key0)
	if err != nil {
		if sdkerrors.IsOf(err, types.ErrRecordNotFound) {
			orderBookID0, err = k.genOrderBookIDsFromDenoms(ctx, denom0, denom1)
			if err != nil {
				return 0, 0, err
			}
		} else {
			return 0, 0, err
		}
	}

	if denom0 == baseDenom {
		return orderBookID0, orderBookID0 + 1, nil
	}

	return orderBookID0 + 1, orderBookID0, nil
}

func (k Keeper) getOrderBookIDByKey(ctx sdk.Context, key []byte) (uint32, error) {
	var val gogotypes.UInt32Value
	if err := k.getDataFromStore(ctx, key, &val); err != nil {
		return 0, sdkerrors.Wrapf(err, "faild to get order book ID by key %v", key)
	}

	return val.GetValue(), nil
}

func (k Keeper) genOrderBookIDsFromDenoms(ctx sdk.Context, denom0, denom1 string) (uint32, error) {
	orderBookID0, err := k.genNextOrderBookID(ctx)
	if err != nil {
		return 0, err
	}
	if err := k.saveOrderBookIDWithData(ctx, orderBookID0, denom0, denom1); err != nil {
		return 0, err
	}

	orderBookID1, err := k.genNextOrderBookID(ctx)
	if err != nil {
		return 0, err
	}
	if err := k.saveOrderBookIDWithData(ctx, orderBookID1, denom1, denom0); err != nil {
		return 0, err
	}

	return orderBookID0, nil
}

func (k Keeper) genNextOrderBookID(ctx sdk.Context) (uint32, error) {
	id, err := k.genNextUint32Sequence(ctx, types.OrderBookSequenceKey)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (k Keeper) saveOrderBookIDWithData(
	ctx sdk.Context,
	orderBookID uint32,
	denom0, denom1 string,
) error {
	key, err := types.CreateOrderBookKey(denom0, denom1)
	if err != nil {
		return err
	}
	if err := k.setDataToStore(ctx, key, &gogotypes.UInt32Value{Value: orderBookID}); err != nil {
		return err
	}

	return k.saveOrderBookData(ctx, orderBookID, types.OrderBookData{
		BaseDenom:  denom0,
		QuoteDenom: denom1,
	})
}

func (k Keeper) createOrder(
	ctx sdk.Context,
	params types.Params,
	order types.Order,
	record types.OrderBookRecord,
) error {
	k.logger(ctx).Debug(
		"Creating new order.",
		"order", order.String(),
		"record", record.String(),
	)

	if err := k.incrementAccountDenomsOrdersCounter(
		ctx,
		record.AccountNumber,
		params.MaxOrdersPerDenom,
		order.Denoms(),
	); err != nil {
		return err
	}

	// update the reserve to be saved in the order data
	if params.OrderReserve.IsPositive() {
		order.Reserve = params.OrderReserve
	}

	// the remaining quantity and balance will be taker from record
	if err := k.saveOrderWithOrderBookRecord(ctx, order, record); err != nil {
		return err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderCreated{
		Creator:           order.Creator,
		ID:                order.ID,
		Sequence:          order.Sequence,
		RemainingQuantity: record.RemainingQuantity,
		RemainingBalance:  record.RemainingBalance,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "failed to emit event EventOrderCreated: %s", err)
	}

	return nil
}

func (k Keeper) saveOrderWithOrderBookRecord(
	ctx sdk.Context,
	order types.Order,
	record types.OrderBookRecord,
) error {
	// additional check to prevent in unexpected state
	if order.Type != types.ORDER_TYPE_LIMIT {
		return sdkerrors.Wrapf(
			types.ErrInvalidInput,
			"it's prohibited to save not limit order types, type: %s",
			order.Type.String(),
		)
	}
	if order.TimeInForce != types.TIME_IN_FORCE_GTC {
		return sdkerrors.Wrapf(
			types.ErrInvalidInput,
			"it's prohibited to save not GTC order types, type: %s",
			order.TimeInForce.String(),
		)
	}

	if err := k.saveOrderBookRecord(ctx, record); err != nil {
		return err
	}

	creator, err := sdk.AccAddressFromBech32(order.Creator)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", order.Creator)
	}

	if order.GoodTil != nil {
		if err := k.delayGoodTilCancellation(
			ctx,
			*order.GoodTil,
			record.OrderSequence,
			creator,
		); err != nil {
			return err
		}
	}

	if err := k.saveOrderData(ctx, record.OrderSequence, types.OrderData{
		OrderID:     order.ID,
		OrderBookID: record.OrderBookID,
		Price:       *order.Price,
		Quantity:    order.Quantity,
		Side:        order.Side,
		GoodTil:     order.GoodTil,
		Reserve:     order.Reserve,
	}); err != nil {
		return err
	}

	if err := k.saveOrderIDToSequence(ctx, record.AccountNumber, record.OrderID, record.OrderSequence); err != nil {
		return err
	}

	return k.saveAccountDenomOrderSequence(ctx, record.AccountNumber, order.Denoms(), record.OrderSequence)
}

func (k Keeper) removeOrderByRecord(
	ctx sdk.Context,
	creator sdk.AccAddress,
	record types.OrderBookRecord,
) error {
	k.logger(ctx).Debug(
		"Removing order by record.",
		"record", record,
	)

	if err := k.removeOrderBookRecord(
		ctx, record.OrderBookID, record.Side, record.Price, record.OrderSequence,
	); err != nil {
		return err
	}

	orderData, err := k.getOrderData(ctx, record.OrderSequence)
	if err != nil {
		return err
	}
	if orderData.GoodTil != nil {
		if err := k.removeGoodTilDelay(ctx, *orderData.GoodTil, record.OrderSequence); err != nil {
			return err
		}
	}

	if err = k.removeOrderData(ctx, record.OrderSequence); err != nil {
		return err
	}

	if err := k.removeOrderIDToSequence(ctx, record.AccountNumber, record.OrderID); err != nil {
		return err
	}

	orderBookData, err := k.getOrderBookData(ctx, record.OrderBookID)
	if err != nil {
		return err
	}

	denoms := []string{orderBookData.BaseDenom, orderBookData.QuoteDenom}
	if err := k.decrementAccountDenomOrdersCounter(ctx, record.AccountNumber, denoms); err != nil {
		return err
	}
	if err := k.removeAccountDenomOrderSequence(ctx, record.AccountNumber, denoms, record.OrderSequence); err != nil {
		return err
	}

	if err := ctx.EventManager().EmitTypedEvent(&types.EventOrderClosed{
		Creator:           creator.String(),
		ID:                record.OrderID,
		Sequence:          record.OrderSequence,
		RemainingQuantity: record.RemainingQuantity,
		RemainingBalance:  record.RemainingBalance,
	}); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "failed to emit event EventOrderCreated: %s", err)
	}

	return nil
}

func (k Keeper) saveOrderBookData(ctx sdk.Context, orderBookID uint32, data types.OrderBookData) error {
	return k.setDataToStore(ctx, types.CreateOrderBookDataKey(orderBookID), &data)
}

func (k Keeper) cancelOrderBySequence(ctx sdk.Context, acc sdk.AccAddress, orderSequence uint64) error {
	orderData, err := k.getOrderData(ctx, orderSequence)
	if err != nil {
		return err
	}
	return k.cancelOrder(ctx, acc, orderData.OrderID)
}

func (k Keeper) cancelOrder(ctx sdk.Context, acc sdk.AccAddress, orderID string) error {
	order, record, err := k.getOrderWithRecordByAddressAndID(ctx, acc, orderID)
	if err != nil {
		return err
	}

	if err := k.removeOrderByRecord(ctx, acc, record); err != nil {
		return err
	}

	lockedCoins := sdk.NewCoins(sdk.NewCoin(order.GetSpendDenom(), order.RemainingBalance))
	expectedToReceiveCoin, err := types.ComputeLimitOrderExpectedToReceiveBalance(
		order.Side, order.BaseDenom, order.QuoteDenom, record.RemainingQuantity, *order.Price,
	)
	if err != nil {
		return err
	}

	// unlock the reserve if present
	if order.Reserve.IsPositive() {
		lockedCoins = lockedCoins.Add(order.Reserve)
	}

	return k.assetFTKeeper.DEXDecreaseLimits(
		ctx, acc, lockedCoins, expectedToReceiveCoin,
	)
}

func (k Keeper) getOrderBookData(ctx sdk.Context, orderBookID uint32) (types.OrderBookData, error) {
	var val types.OrderBookData
	if err := k.getDataFromStore(ctx, types.CreateOrderBookDataKey(orderBookID), &val); err != nil {
		return types.OrderBookData{},
			sdkerrors.Wrapf(err, "failed to get order book data, orderBookID: %d", orderBookID)
	}
	return val, nil
}

func (k Keeper) genNextOrderSequence(ctx sdk.Context) (uint64, error) {
	return k.genNextUint64Sequence(ctx, types.OrderSequenceKey)
}

func (k Keeper) saveOrderBookRecord(
	ctx sdk.Context,
	record types.OrderBookRecord,
) error {
	k.logger(ctx).Debug("Saving order book record.", "record", record.String())

	key, err := types.CreateOrderBookRecordKey(record.OrderBookID, record.Side, record.Price, record.OrderSequence)
	if err != nil {
		return err
	}

	return k.setDataToStore(ctx, key, &types.OrderBookRecordData{
		OrderID:           record.OrderID,
		AccountNumber:     record.AccountNumber,
		RemainingQuantity: record.RemainingQuantity,
		RemainingBalance:  record.RemainingBalance,
	})
}

func (k Keeper) getOrderWithRecordByAddressAndID(
	ctx sdk.Context,
	acc sdk.AccAddress,
	orderID string,
) (types.Order, types.OrderBookRecord, error) {
	accNumber, err := k.getAccountNumber(ctx, acc)
	if err != nil {
		return types.Order{}, types.OrderBookRecord{}, err
	}

	orderSequence, err := k.getOrderSequenceByID(ctx, accNumber, orderID)
	if err != nil {
		return types.Order{}, types.OrderBookRecord{}, err
	}

	orderData, err := k.getOrderData(ctx, orderSequence)
	if err != nil {
		return types.Order{}, types.OrderBookRecord{}, err
	}

	orderBookRecord, err := k.getOrderBookRecord(
		ctx,
		orderData.OrderBookID,
		orderData.Side,
		orderData.Price,
		orderSequence,
	)
	if err != nil {
		return types.Order{}, types.OrderBookRecord{}, err
	}

	orderBookData, err := k.getOrderBookData(ctx, orderData.OrderBookID)
	if err != nil {
		return types.Order{}, types.OrderBookRecord{}, err
	}

	return types.Order{
			Creator:           acc.String(),
			Type:              types.ORDER_TYPE_LIMIT,
			ID:                orderID,
			Sequence:          orderSequence,
			BaseDenom:         orderBookData.BaseDenom,
			QuoteDenom:        orderBookData.QuoteDenom,
			Price:             &orderBookRecord.Price,
			Quantity:          orderData.Quantity,
			Side:              orderBookRecord.Side,
			GoodTil:           orderData.GoodTil,
			TimeInForce:       types.TIME_IN_FORCE_GTC,
			RemainingQuantity: orderBookRecord.RemainingQuantity,
			RemainingBalance:  orderBookRecord.RemainingBalance,
			Reserve:           orderData.Reserve,
		},
		orderBookRecord,
		nil
}

func (k Keeper) getOrderBookRecord(
	ctx sdk.Context,
	orderBookID uint32,
	side types.Side,
	price types.Price,
	orderSequence uint64,
) (types.OrderBookRecord, error) {
	key, err := types.CreateOrderBookRecordKey(orderBookID, side, price, orderSequence)
	if err != nil {
		return types.OrderBookRecord{}, err
	}

	var val types.OrderBookRecordData
	if err := k.getDataFromStore(ctx, key, &val); err != nil {
		return types.OrderBookRecord{},
			sdkerrors.Wrapf(
				err,
				"faild to get order book record, orderBookID: %d, side: %s, price: %s, orderSequence: %d",
				orderBookID, side.String(), price.String(), orderSequence)
	}
	return types.OrderBookRecord{
		OrderBookID:       orderBookID,
		Side:              side,
		Price:             price,
		OrderSequence:     orderSequence,
		OrderID:           val.OrderID,
		AccountNumber:     val.AccountNumber,
		RemainingQuantity: val.RemainingQuantity,
		RemainingBalance:  val.RemainingBalance,
	}, nil
}

func (k Keeper) getPaginatedOrders(
	ctx sdk.Context,
	acc sdk.AccAddress,
	pagination *query.PageRequest,
) ([]types.Order, *query.PageResponse, error) {
	accNumber, err := k.getAccountNumber(ctx, acc)
	if err != nil {
		return nil, nil, err
	}
	moduleStore := k.storeService.OpenKVStore(ctx)
	store := prefix.NewStore(runtime.KVStoreAdapter(moduleStore), types.CreateOrderIDToSequenceKeyPrefix(accNumber))
	orderBookIDToOrderBookData := make(map[uint32]types.OrderBookData)
	orders, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
		pagination,
		// builder
		func(_ []byte, record *gogotypes.UInt64Value) (*types.Order, error) {
			orderSequence := record.Value
			orderData, err := k.getOrderData(ctx, orderSequence)
			if err != nil {
				return nil, err
			}

			orderBookID := orderData.OrderBookID
			orderBookData, ok := orderBookIDToOrderBookData[orderBookID]
			if !ok {
				orderBookData, err = k.getOrderBookData(ctx, orderBookID)
				if err != nil {
					return nil, err
				}
				orderBookIDToOrderBookData[orderBookID] = orderBookData
			}

			orderBookRecord, err := k.getOrderBookRecord(
				ctx,
				orderData.OrderBookID,
				orderData.Side,
				orderData.Price,
				orderSequence,
			)
			if err != nil {
				return nil, err
			}

			return &types.Order{
				Creator:           acc.String(),
				Type:              types.ORDER_TYPE_LIMIT,
				ID:                orderBookRecord.OrderID,
				Sequence:          orderSequence,
				BaseDenom:         orderBookData.BaseDenom,
				QuoteDenom:        orderBookData.QuoteDenom,
				Price:             &orderData.Price,
				Quantity:          orderData.Quantity,
				Side:              orderData.Side,
				GoodTil:           orderData.GoodTil,
				TimeInForce:       types.TIME_IN_FORCE_GTC,
				RemainingQuantity: orderBookRecord.RemainingQuantity,
				RemainingBalance:  orderBookRecord.RemainingBalance,
				Reserve:           orderData.Reserve,
			}, nil
		},
		// constructor
		func() *gogotypes.UInt64Value {
			return &gogotypes.UInt64Value{}
		},
	)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(types.ErrInvalidInput, "failed to paginate: %s", err)
	}
	return lo.Map(orders, func(order *types.Order, _ int) types.Order {
		return *order
	}), pageRes, nil
}

func (k Keeper) getPaginatedOrderBooksWithID(
	ctx sdk.Context,
	pagination *query.PageRequest,
) ([]types.OrderBookDataWithID, *query.PageResponse, error) {
	moduleStore := k.storeService.OpenKVStore(ctx)
	store := prefix.NewStore(runtime.KVStoreAdapter(moduleStore), types.OrderBookDataKeyPrefix)
	orders, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
		pagination,
		// builder
		func(key []byte, record *types.OrderBookData) (*types.OrderBookDataWithID, error) {
			id, err := types.DecodeOrderBookDataKey(key)
			if err != nil {
				return nil, err
			}

			return &types.OrderBookDataWithID{
				ID:   id,
				Data: *record,
			}, nil
		},
		// constructor
		func() *types.OrderBookData {
			return &types.OrderBookData{}
		},
	)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(types.ErrInvalidInput, "failed to paginate: %s", err)
	}
	return lo.Map(orders, func(data *types.OrderBookDataWithID, _ int) types.OrderBookDataWithID {
		return *data
	}), pageRes, nil
}

func (k Keeper) getPaginatedOrderBookOrders(
	ctx sdk.Context,
	baseDenom, quoteDenom string,
	side types.Side,
	pagination *query.PageRequest,
) ([]types.Order, *query.PageResponse, error) {
	orderBookID, err := k.getOrderBookIDByDenoms(ctx, baseDenom, quoteDenom)
	if err != nil {
		return nil, nil, err
	}

	moduleStore := k.storeService.OpenKVStore(ctx)
	store := prefix.NewStore(runtime.KVStoreAdapter(moduleStore), types.CreateOrderBookSideKey(orderBookID, side))
	accNumberToAddCache := make(map[uint64]sdk.AccAddress)

	orders, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
		pagination,
		// builder
		func(key []byte, record *types.OrderBookRecordData) (*types.Order, error) {
			// decode key to values
			price, orderSequence, err := types.DecodeOrderBookSideRecordKey(key)
			if err != nil {
				return nil, err
			}

			var acc sdk.AccAddress
			acc, err = k.getAccountAddressWithCache(ctx, record.AccountNumber, accNumberToAddCache)
			if err != nil {
				return nil, err
			}

			orderData, err := k.getOrderData(ctx, orderSequence)
			if err != nil {
				return nil, err
			}

			return &types.Order{
				Creator:           acc.String(),
				Type:              types.ORDER_TYPE_LIMIT,
				ID:                record.OrderID,
				Sequence:          orderSequence,
				BaseDenom:         baseDenom,
				QuoteDenom:        quoteDenom,
				Price:             &price,
				Quantity:          orderData.Quantity,
				Side:              side,
				GoodTil:           orderData.GoodTil,
				TimeInForce:       types.TIME_IN_FORCE_GTC,
				RemainingQuantity: record.RemainingQuantity,
				RemainingBalance:  record.RemainingBalance,
				Reserve:           orderData.Reserve,
			}, nil
		},
		// constructor
		func() *types.OrderBookRecordData {
			return &types.OrderBookRecordData{}
		},
	)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(types.ErrInvalidInput, "failed to paginate: %s", err)
	}
	return lo.Map(orders, func(order *types.Order, _ int) types.Order {
		return *order
	}), pageRes, nil
}

func (k Keeper) removeOrderBookRecord(
	ctx sdk.Context,
	orderBookID uint32,
	side types.Side,
	price types.Price,
	orderSequence uint64,
) error {
	key, err := types.CreateOrderBookRecordKey(orderBookID, side, price, orderSequence)
	if err != nil {
		return err
	}
	return k.storeService.OpenKVStore(ctx).Delete(key)
}

func (k Keeper) saveOrderData(ctx sdk.Context, orderSequence uint64, data types.OrderData) error {
	return k.setDataToStore(ctx, types.CreateOrderKey(orderSequence), &data)
}

func (k Keeper) removeOrderData(ctx sdk.Context, orderSequence uint64) error {
	return k.storeService.OpenKVStore(ctx).Delete(types.CreateOrderKey(orderSequence))
}

func (k Keeper) getOrderData(ctx sdk.Context, orderSequence uint64) (types.OrderData, error) {
	var val types.OrderData
	if err := k.getDataFromStore(ctx, types.CreateOrderKey(orderSequence), &val); err != nil {
		return types.OrderData{}, sdkerrors.Wrapf(err, "failed to get order data, orderSequence: %d", orderSequence)
	}
	return val, nil
}

func (k Keeper) saveOrderIDToSequence(ctx sdk.Context, accNumber uint64, orderID string, orderSequence uint64) error {
	key := types.CreateOrderIDToSequenceKey(accNumber, orderID)
	return k.setDataToStore(ctx, key, &gogotypes.UInt64Value{Value: orderSequence})
}

func (k Keeper) removeOrderIDToSequence(ctx sdk.Context, accNumber uint64, orderID string) error {
	return k.storeService.OpenKVStore(ctx).Delete(types.CreateOrderIDToSequenceKey(accNumber, orderID))
}

func (k Keeper) getOrderSequenceByID(ctx sdk.Context, accNumber uint64, orderID string) (uint64, error) {
	var val gogotypes.UInt64Value
	if err := k.getDataFromStore(ctx, types.CreateOrderIDToSequenceKey(accNumber, orderID), &val); err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to get order sequence, accNumber: %d, orderID: %s", accNumber, orderID)
	}

	return val.GetValue(), nil
}

func (k Keeper) setAccountDenomOrdersCount(
	ctx sdk.Context,
	accountDenomOrdersCount types.AccountDenomOrdersCount,
) error {
	key, err := types.CreateAccountDenomOrdersCountKey(
		accountDenomOrdersCount.AccountNumber, accountDenomOrdersCount.Denom,
	)
	if err != nil {
		return err
	}

	return k.setUint64Value(ctx, key, accountDenomOrdersCount.OrdersCount)
}

func (k Keeper) incrementAccountDenomsOrdersCounter(
	ctx sdk.Context,
	accNumber uint64,
	maxOrdersPerDenom uint64,
	denoms []string,
) error {
	for _, denom := range denoms {
		key, err := types.CreateAccountDenomOrdersCountKey(accNumber, denom)
		if err != nil {
			return err
		}
		orderPerDenomCount, err := k.incrementUint64Counter(ctx, key)
		if err != nil {
			return err
		}
		if orderPerDenomCount > maxOrdersPerDenom {
			return sdkerrors.Wrapf(
				types.ErrInvalidInput,
				"it's prohibited to save more than %d orders per denom",
				maxOrdersPerDenom,
			)
		}
	}

	return nil
}

func (k Keeper) decrementAccountDenomOrdersCounter(
	ctx sdk.Context,
	accNumber uint64,
	denoms []string,
) error {
	for _, denom := range denoms {
		key, err := types.CreateAccountDenomOrdersCountKey(accNumber, denom)
		if err != nil {
			return err
		}
		_, err = k.decrementUint64Counter(ctx, key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) saveAccountDenomOrderSequence(
	ctx sdk.Context, accNumber uint64, denoms []string, orderSequence uint64,
) error {
	for _, denom := range denoms {
		key, err := types.CreateAccountDenomOrderSequenceKey(accNumber, denom, orderSequence)
		if err != nil {
			return err
		}
		// save empty slice

		if err = k.storeService.OpenKVStore(ctx).Set(key, make([]byte, 0)); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) removeAccountDenomOrderSequence(
	ctx sdk.Context, accNumber uint64, denoms []string, orderSequence uint64,
) error {
	for _, denom := range denoms {
		key, err := types.CreateAccountDenomOrderSequenceKey(accNumber, denom, orderSequence)
		if err != nil {
			return err
		}
		// remove all
		err = k.storeService.OpenKVStore(ctx).Delete(key)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) getAccountDenomOrdersCounter(ctx sdk.Context, accNumber uint64, denom string) (uint64, error) {
	key, err := types.CreateAccountDenomOrdersCountKey(accNumber, denom)
	if err != nil {
		return 0, err
	}

	var val gogotypes.UInt64Value
	err = k.getDataFromStore(ctx, key, &val)
	if err != nil {
		if !sdkerrors.IsOf(err, types.ErrRecordNotFound) {
			return 0, err
		}
		// record not found so the count is zero
		return 0, nil
	}

	return val.Value, nil
}

// logger returns the Keeper logger.
func (k Keeper) logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", "x/"+types.ModuleName)
}

// ComputePriceTick returns the price tick of a given ref amounts and price tick exponent.
func ComputePriceTick(baseDenomRefAmount, quoteRefAmount sdkmath.LegacyDec, priceTickExponent int32) *big.Rat {
	// 10^(floor(log10((quoteRefAmountRat / baseRefAmountRat))) + price_tick_exponent)
	exponent := ratFloorLog10(
		cbig.NewRatFromBigInts(quoteRefAmount.BigInt(), baseDenomRefAmount.BigInt()),
	) + int(priceTickExponent)
	if exponent < 0 {
		return cbig.NewRatFromBigInts(big.NewInt(1), cbig.IntTenToThePower(big.NewInt(int64(-exponent))))
	}

	return cbig.NewRatFromBigInt(cbig.IntTenToThePower(big.NewInt(int64(exponent))))
}

func ratFloorLog10(val *big.Rat) int {
	num := val.Num()
	denom := val.Denom()

	// if val >= 1 the floor(log10(val)) value is equal to length of int part
	if cbig.IntGTE(num, denom) {
		return len(cbig.IntQuo(num, denom).Text(10)) - 1
	}

	// define the max exponent as dif or num and denom length
	exponent := len(num.Text(10)) - len(denom.Text(10))
	// if (val * 10^-exp) < 1 we need to decrease the exponent to get the correct floor(log10(val))
	if cbig.RatLT(
		cbig.RatMul(val, cbig.NewRatFromBigInt(cbig.IntTenToThePower(big.NewInt(int64(-exponent))))),
		cbig.NewRatFromInt64(1),
	) {
		exponent--
	}

	return exponent
}
