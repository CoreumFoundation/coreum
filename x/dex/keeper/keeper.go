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

	_, found, err := k.getOrderSeqByID(ctx, accNumber, order.ID)
	if err != nil {
		return err
	}
	if found {
		return sdkerrors.Wrapf(types.ErrInvalidState, "order with the id %q is already created", order.ID)
	}

	orderBookID, _, err := k.getOrGenOrderBookIDs(ctx, order.BaseDenom, order.QuoteDenom)
	if err != nil {
		return err
	}

	return k.matchOrder(ctx, accNumber, orderBookID, order)
}

// GetOrderBookIDByDenoms returns order book ID by it's denoms.
func (k Keeper) GetOrderBookIDByDenoms(ctx sdk.Context, baseDenom, quoteDenom string) (uint32, bool, error) {
	orderBookIDKey, err := types.CreateOrderBookKey(baseDenom, quoteDenom)
	if err != nil {
		return 0, false, err
	}

	return k.getOrderBookIDByKey(ctx, orderBookIDKey)
}

// GetOrderByAddressAndID returns order by holder address and it's ID.
func (k Keeper) GetOrderByAddressAndID(ctx sdk.Context, acc sdk.AccAddress, orderID string) (types.Order, bool, error) {
	accountNumber, err := k.getAccountNumber(ctx, acc.String())
	if err != nil {
		return types.Order{}, false, err
	}

	orderSeq, found, err := k.getOrderSeqByID(ctx, accountNumber, orderID)
	if err != nil {
		return types.Order{}, false, err
	}
	if !found {
		return types.Order{}, false, nil
	}

	orderData, found, err := k.getOrderData(ctx, orderSeq)
	if err != nil {
		return types.Order{}, false, err
	}
	if !found {
		return types.Order{}, false, sdkerrors.Wrapf(
			types.ErrInvalidState, "failed to get order data by order seq, not found: %d", orderSeq,
		)
	}

	orderBookData, found, err := k.getOrderBookData(ctx, orderData.OrderBookID)
	if err != nil {
		return types.Order{}, false, err
	}
	if !found {
		return types.Order{}, false, sdkerrors.Wrapf(
			types.ErrInvalidState, "failed to get order book data by ID, not found: %d", orderData.OrderBookID,
		)
	}

	orderBookRecord, found, err := k.getOrderBookRecord(
		ctx,
		orderData.OrderBookID,
		orderData.Side,
		orderData.Price,
		orderSeq,
	)
	if err != nil {
		return types.Order{}, false, err
	}
	if !found {
		return types.Order{}, false, sdkerrors.Wrapf(
			types.ErrInvalidState, "failed to get order record, orderData: %v, orderSeq: %d", orderData, orderSeq,
		)
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
	}, true, nil
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
	orderBookID0, found, err := k.getOrderBookIDByKey(ctx, key0)
	if err != nil {
		return 0, 0, err
	}
	if !found {
		orderBookID0, err = k.genOrderBookIDsFromDenoms(ctx, key0, denom0, denom1)
		if err != nil {
			return 0, 0, err
		}
	}

	if denom0 == baseDenom {
		return orderBookID0, orderBookID0 + 1, nil
	}

	return orderBookID0 + 1, orderBookID0, nil
}

func (k Keeper) getOrderBookIDByKey(ctx sdk.Context, key []byte) (uint32, bool, error) {
	var val gogotypes.UInt32Value
	found, err := k.getDataFromStore(ctx, key, &val)
	return val.GetValue(), found, err
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

func (k Keeper) getOrderBookData(ctx sdk.Context, orderBookID uint32) (types.OrderBookData, bool, error) {
	var val types.OrderBookData
	found, err := k.getDataFromStore(ctx, types.CreateOrderBookDataKey(orderBookID), &val)
	return val, found, err
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
) (types.OrderBookRecordData, bool, error) {
	key, err := types.CreateOrderBookRecordKey(orderBookID, side, price, orderSeq)
	if err != nil {
		return types.OrderBookRecordData{}, false, err
	}

	var val types.OrderBookRecordData
	found, err := k.getDataFromStore(ctx, key, &val)
	return val, found, err
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

func (k Keeper) getOrderData(ctx sdk.Context, orderSeq uint64) (types.OrderData, bool, error) {
	var val types.OrderData
	found, err := k.getDataFromStore(ctx, types.CreateOrderKey(orderSeq), &val)
	return val, found, err
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

func (k Keeper) getOrderSeqByID(ctx sdk.Context, accountNumber uint64, orderID string) (uint64, bool, error) {
	key, err := types.CreateOrderIDToSeqKey(accountNumber, orderID)
	if err != nil {
		return 0, false, err
	}
	var val gogotypes.UInt64Value
	found, err := k.getDataFromStore(ctx, key, &val)
	return val.GetValue(), found, err
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
) (bool, error) {
	bz := ctx.KVStore(k.storeKey).Get(key)
	if bz == nil {
		return false, nil
	}

	if err := k.cdc.Unmarshal(bz, val); err != nil {
		return false, sdkerrors.Wrapf(types.ErrInvalidState, "failed to unmarshal %T, err: %s", err, val)
	}

	return true, nil
}

// logger returns the Keeper logger.
func (k Keeper) logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}
