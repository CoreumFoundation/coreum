package keeper

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gogotypes "github.com/cosmos/gogoproto/types"

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

	accNumber, err := k.getAccountNumber(ctx, order.Account)
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
	accountNumber, err := k.getAccountNumber(ctx, acc.String())
	if err != nil {
		return types.Order{}, err
	}

	orderSeq, err := k.getOrderSeqByID(ctx, accountNumber, orderID)
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
		Account:           acc.String(),
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
	key, err := types.CreateOrderIDToSeqKey(accountNumber, orderID)
	if err != nil {
		return err
	}

	return k.setDataToStore(ctx, key, &gogotypes.UInt64Value{Value: orderSeq})
}

func (k Keeper) removeOrderIDToSeq(ctx sdk.Context, accountNumber uint64, orderID string) error {
	key, err := types.CreateOrderIDToSeqKey(accountNumber, orderID)
	if err != nil {
		return err
	}
	ctx.KVStore(k.storeKey).Delete(key)
	return nil
}

func (k Keeper) getOrderSeqByID(ctx sdk.Context, accountNumber uint64, orderID string) (uint64, error) {
	key, err := types.CreateOrderIDToSeqKey(accountNumber, orderID)
	if err != nil {
		return 0, err
	}
	var val gogotypes.UInt64Value
	if err := k.getDataFromStore(ctx, key, &val); err != nil {
		return 0, sdkerrors.Wrapf(err, "failed to get order seq, accountNumber: %d, orderID: %s", accountNumber, orderID)
	}
	return val.GetValue(), err
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
