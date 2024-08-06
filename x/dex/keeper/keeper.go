package keeper

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

// Keeper is the dex module keeper.
type Keeper struct {
	cdc           codec.BinaryCodec
	storeKey      storetypes.StoreKey
	accountKeeper types.AccountKeeper
	bankKeeper    types.BankKeeper
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	bankKeeper types.BankKeeper,
) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper,
		bankKeeper:    bankKeeper,
	}
}

// PlaceOrder places an order on the corresponding order book, and matches the order.
func (k Keeper) PlaceOrder(ctx sdk.Context, order types.Order) error {
	if err := order.Validate(); err != nil {
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
	_, err = k.getOrderSeqByID(ctx, accNumber, order.ID)
	if err != nil {
		if !sdkerrors.IsOf(err, types.ErrRecordNotFound) {
			return err
		}
	} else {
		return sdkerrors.Wrapf(types.ErrInvalidInput, "order with the id %q is already created", order.ID)
	}

	orderBookID, _, err := k.getOrGenOrderBookIDs(ctx, order.BaseDenom, order.QuoteDenom)
	if err != nil {
		return err
	}

	return k.matchOrder(ctx, accNumber, orderBookID, order)
}

// GetOrderBookIDByDenoms returns order book ID by it's denoms.
func (k Keeper) GetOrderBookIDByDenoms(ctx sdk.Context, baseDenom, quoteDenom string) (uint32, error) {
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

// GetOrderByAddressAndID returns order by holder address and it's ID.
func (k Keeper) GetOrderByAddressAndID(ctx sdk.Context, acc sdk.AccAddress, orderID string) (types.Order, error) {
	accNumber, err := k.getAccountNumber(ctx, acc)
	if err != nil {
		return types.Order{}, err
	}

	orderSeq, err := k.getOrderSeqByID(ctx, accNumber, orderID)
	if err != nil {
		return types.Order{}, err
	}

	orderData, err := k.getOrderData(ctx, orderSeq)
	if err != nil {
		return types.Order{}, err
	}

	orderBookData, err := k.getOrderBookData(ctx, orderData.OrderBookID)
	if err != nil {
		return types.Order{}, err
	}

	orderBookRecord, err := k.getOrderBookRecord(
		ctx,
		orderData.OrderBookID,
		orderData.Side,
		orderData.Price,
		orderSeq,
	)
	if err != nil {
		return types.Order{}, err
	}

	return types.Order{
		Creator:           acc.String(),
		ID:                orderID,
		BaseDenom:         orderBookData.BaseDenom,
		QuoteDenom:        orderBookData.QuoteDenom,
		Price:             orderData.Price,
		Quantity:          orderData.Quantity,
		Side:              orderData.Side,
		RemainingQuantity: orderBookRecord.RemainingQuantity,
		RemainingBalance:  orderBookRecord.RemainingBalance,
	}, nil
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
	return k.getPaginatedOrderBooks(ctx, pagination)
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
			orderBookID0, err = k.genOrderBookIDsFromDenoms(ctx, key0, denom0, denom1)
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

func (k Keeper) genOrderBookIDsFromDenoms(ctx sdk.Context, key []byte, denom0, denom1 string) (uint32, error) {
	orderBookID0, err := k.genNextOrderBookID(ctx, key)
	if err != nil {
		return 0, err
	}
	if err := k.saveOrderBookData(ctx, orderBookID0, types.OrderBookData{
		BaseDenom:  denom0,
		QuoteDenom: denom1,
	}); err != nil {
		return 0, err
	}

	key1, err := types.CreateOrderBookKey(denom1, denom0)
	if err != nil {
		return 0, err
	}
	orderBookID1, err := k.genNextOrderBookID(ctx, key1)
	if err != nil {
		return 0, err
	}
	if err := k.saveOrderBookData(ctx, orderBookID1, types.OrderBookData{
		BaseDenom:  denom1,
		QuoteDenom: denom0,
	}); err != nil {
		return 0, err
	}

	return orderBookID0, nil
}

func (k Keeper) genNextOrderBookID(ctx sdk.Context, key []byte) (uint32, error) {
	id, err := k.genNextUint32Seq(ctx, types.OrderBookSeqKey)
	if err != nil {
		return 0, err
	}
	if err := k.setDataToStore(ctx, key, &gogotypes.UInt32Value{Value: id}); err != nil {
		return 0, err
	}

	return id, nil
}

func (k Keeper) createOrder(
	ctx sdk.Context,
	order types.Order,
	record types.OrderBookRecord,
) error {
	k.logger(ctx).Debug(
		"Creating new order.",
		"order", order.String(),
		"record", record.String(),
	)

	orderSeq, err := k.genNextOrderSeq(ctx)
	if err != nil {
		return err
	}
	record.OrderSeq = orderSeq

	if err := k.saveOrderBookRecord(ctx, record); err != nil {
		return err
	}

	if err := k.savaOrderData(ctx, orderSeq, types.OrderData{
		OrderBookID: record.OrderBookID,
		Price:       order.Price,
		Quantity:    order.Quantity,
		Side:        order.Side,
	}); err != nil {
		return err
	}

	return k.savaOrderIDToSeq(ctx, record.AccountNumber, record.OrderID, orderSeq)
}

func (k Keeper) removeOrderByRecord(
	ctx sdk.Context,
	record types.OrderBookRecord,
) error {
	k.logger(ctx).Debug(
		"Removing order.",
		"record", record,
	)

	if err := k.removeOrderBookRecord(ctx, record.OrderBookID, record.Side, record.Price, record.OrderSeq); err != nil {
		return err
	}
	k.removeOrderData(ctx, record.OrderSeq)
	return k.removeOrderIDToSeq(ctx, record.AccountNumber, record.OrderID)
}

func (k Keeper) saveOrderBookData(ctx sdk.Context, orderBookID uint32, data types.OrderBookData) error {
	return k.setDataToStore(ctx, types.CreateOrderBookDataKey(orderBookID), &data)
}

func (k Keeper) getOrderBookData(ctx sdk.Context, orderBookID uint32) (types.OrderBookData, error) {
	var val types.OrderBookData
	if err := k.getDataFromStore(ctx, types.CreateOrderBookDataKey(orderBookID), &val); err != nil {
		return types.OrderBookData{},
			sdkerrors.Wrapf(err, "failed to get order book data, orderBookID: %d", orderBookID)
	}
	return val, nil
}

func (k Keeper) genNextOrderSeq(ctx sdk.Context) (uint64, error) {
	return k.genNextUint64Seq(ctx, types.OrderSequenceKey)
}

func (k Keeper) saveOrderBookRecord(
	ctx sdk.Context,
	record types.OrderBookRecord,
) error {
	k.logger(ctx).Debug("Saving order book record.", "record", record.String())

	key, err := types.CreateOrderBookRecordKey(record.OrderBookID, record.Side, record.Price, record.OrderSeq)
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

func (k Keeper) getOrderBookRecord(
	ctx sdk.Context,
	orderBookID uint32,
	side types.Side,
	price types.Price,
	orderSeq uint64,
) (types.OrderBookRecordData, error) {
	key, err := types.CreateOrderBookRecordKey(orderBookID, side, price, orderSeq)
	if err != nil {
		return types.OrderBookRecordData{}, err
	}

	var val types.OrderBookRecordData
	if err := k.getDataFromStore(ctx, key, &val); err != nil {
		return types.OrderBookRecordData{},
			sdkerrors.Wrapf(
				err,
				"faild to get order book record, orderBookID: %d, side: %s, price: %s, orderSeq: %d",
				orderBookID, side.String(), price.String(), orderSeq)
	}
	return val, nil
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

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CreateOrderIDToSeqKeyPrefix(accNumber))
	orderBookIDToOrderBookData := make(map[uint32]types.OrderBookData)
	orders, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
		pagination,
		// builder
		func(_ []byte, record *gogotypes.UInt64Value) (*types.Order, error) {
			orderSeq := record.Value
			orderData, err := k.getOrderData(ctx, orderSeq)
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
				orderSeq,
			)
			if err != nil {
				return nil, err
			}

			return &types.Order{
				Creator:           acc.String(),
				ID:                orderBookRecord.OrderID,
				BaseDenom:         orderBookData.BaseDenom,
				QuoteDenom:        orderBookData.QuoteDenom,
				Price:             orderData.Price,
				Quantity:          orderData.Quantity,
				Side:              orderData.Side,
				RemainingQuantity: orderBookRecord.RemainingQuantity,
				RemainingBalance:  orderBookRecord.RemainingBalance,
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

func (k Keeper) getPaginatedOrderBooks(
	ctx sdk.Context,
	pagination *query.PageRequest,
) ([]types.OrderBookData, *query.PageResponse, error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.OrderBookDataKeyPrefix)
	orders, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
		pagination,
		// builder
		func(_ []byte, record *types.OrderBookData) (*types.OrderBookData, error) {
			return record, nil
		},
		// constructor
		func() *types.OrderBookData {
			return &types.OrderBookData{}
		},
	)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(types.ErrInvalidInput, "failed to paginate: %s", err)
	}
	return lo.Map(orders, func(data *types.OrderBookData, _ int) types.OrderBookData {
		return *data
	}), pageRes, nil
}

func (k Keeper) getPaginatedOrderBookOrders(
	ctx sdk.Context,
	baseDenom, quoteDenom string,
	side types.Side,
	pagination *query.PageRequest,
) ([]types.Order, *query.PageResponse, error) {
	orderBookID, err := k.GetOrderBookIDByDenoms(ctx, baseDenom, quoteDenom)
	if err != nil {
		return nil, nil, err
	}

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CreateOrderBookSideKey(orderBookID, side))
	accNumberToAddress := make(map[uint64]sdk.AccAddress)
	orders, pageRes, err := query.GenericFilteredPaginate(
		k.cdc,
		store,
		pagination,
		// builder
		func(key []byte, record *types.OrderBookRecordData) (*types.Order, error) {
			// decode key to values
			price, orderSeq, err := types.DecodeOrderBookSideRecordKey(key)
			if err != nil {
				return nil, err
			}

			acc, ok := accNumberToAddress[record.AccountNumber]
			if !ok {
				acc, err = k.getAccountAddress(ctx, record.AccountNumber)
				if err != nil {
					return nil, err
				}
				accNumberToAddress[record.AccountNumber] = acc
			}

			orderData, err := k.getOrderData(ctx, orderSeq)
			if err != nil {
				return nil, err
			}

			return &types.Order{
				Creator:           acc.String(),
				ID:                record.OrderID,
				BaseDenom:         baseDenom,
				QuoteDenom:        quoteDenom,
				Price:             price,
				Quantity:          orderData.Quantity,
				Side:              side,
				RemainingQuantity: record.RemainingQuantity,
				RemainingBalance:  record.RemainingBalance,
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
	orderSeq uint64,
) error {
	key, err := types.CreateOrderBookRecordKey(orderBookID, side, price, orderSeq)
	if err != nil {
		return err
	}
	ctx.KVStore(k.storeKey).Delete(key)

	return nil
}

func (k Keeper) savaOrderData(ctx sdk.Context, orderSeq uint64, data types.OrderData) error {
	return k.setDataToStore(ctx, types.CreateOrderKey(orderSeq), &data)
}

func (k Keeper) removeOrderData(ctx sdk.Context, orderSeq uint64) {
	ctx.KVStore(k.storeKey).Delete(types.CreateOrderKey(orderSeq))
}

func (k Keeper) getOrderData(ctx sdk.Context, orderSeq uint64) (types.OrderData, error) {
	var val types.OrderData
	if err := k.getDataFromStore(ctx, types.CreateOrderKey(orderSeq), &val); err != nil {
		return types.OrderData{}, sdkerrors.Wrapf(err, "failed to get order data, orderSeq: %d", orderSeq)
	}
	return val, nil
}

func (k Keeper) savaOrderIDToSeq(ctx sdk.Context, accountNumber uint64, orderID string, orderSeq uint64) error {
	key := types.CreateOrderIDToSeqKey(accountNumber, orderID)
	return k.setDataToStore(ctx, key, &gogotypes.UInt64Value{Value: orderSeq})
}

func (k Keeper) removeOrderIDToSeq(ctx sdk.Context, accountNumber uint64, orderID string) error {
	ctx.KVStore(k.storeKey).Delete(types.CreateOrderIDToSeqKey(accountNumber, orderID))
	return nil
}

func (k Keeper) getOrderSeqByID(ctx sdk.Context, accountNumber uint64, orderID string) (uint64, error) {
	var val gogotypes.UInt64Value
	if err := k.getDataFromStore(ctx, types.CreateOrderIDToSeqKey(accountNumber, orderID), &val); err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to get order seq, accountNumber: %d, orderID: %s", accountNumber, orderID)
	}

	return val.GetValue(), nil
}

func (k Keeper) setDataToStore(
	ctx sdk.Context,
	key []byte,
	val codec.ProtoMarshaler,
) error {
	bz, err := k.cdc.Marshal(val)
	if err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to marshal %T, err: %s", err, val)
	}
	ctx.KVStore(k.storeKey).Set(key, bz)
	return nil
}

func (k Keeper) getDataFromStore(
	ctx sdk.Context,
	key []byte,
	val codec.ProtoMarshaler,
) error {
	bz := ctx.KVStore(k.storeKey).Get(key)
	if bz == nil {
		return sdkerrors.Wrapf(types.ErrRecordNotFound, "store type %T", val)
	}

	if err := k.cdc.Unmarshal(bz, val); err != nil {
		return sdkerrors.Wrapf(types.ErrInvalidState, "failed to unmarshal %T, err: %s", err, val)
	}

	return nil
}

// logger returns the Keeper logger.
func (k Keeper) logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
