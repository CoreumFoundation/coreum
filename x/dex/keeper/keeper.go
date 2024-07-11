package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
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
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
) Keeper {
	return Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		accountKeeper: accountKeeper,
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

	orderSeq, err := k.genNextOrderSeq(ctx)
	if err != nil {
		return err
	}

	if err := k.saveOrderBookRecordData(ctx, orderBookID, order.Side, order.Price, orderSeq, types.OrderBookRecordData{
		OrderID:           order.ID,
		AccountNumber:     accNumber,
		RemainingQuantity: order.Quantity,
		// For now, we place an order with an expectation, that the user holds enough coins at the time of the match.
		RemainingBalance: sdkmath.Int{},
	}); err != nil {
		return err
	}

	if err := k.savaOrderData(ctx, orderSeq, types.OrderData{
		OrderBookID: orderBookID,
		Price:       order.Price,
		Quantity:    order.Quantity,
		Side:        order.Side,
	}); err != nil {
		return err
	}

	return k.savaOrderIDToSeq(ctx, accNumber, order.ID, orderSeq)
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
		return types.Order{}, false, nil
	}

	orderBookData, found, err := k.getOrderBookData(ctx, orderData.OrderBookID)
	if err != nil {
		return types.Order{}, false, err
	}
	if !found {
		return types.Order{}, false, nil
	}

	return types.Order{
		Account:    acc.String(),
		ID:         orderID,
		BaseDenom:  orderBookData.BaseDenom,
		QuoteDenom: orderBookData.QuoteDenom,
		Price:      orderData.Price,
		Quantity:   orderData.Quantity,
		Side:       orderData.Side,
	}, true, nil
}

// IterateOrderBook iterates an order book.
func (k Keeper) IterateOrderBook(
	ctx sdk.Context,
	orderBookID uint32,
	side types.Side,
	reverse bool,
	cb func(types.OrderBookRecord) (bool, error),
) error {
	var iterator storetypes.Iterator
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.CreateOrderBookSideKey(orderBookID, side))
	if reverse {
		iterator = store.ReverseIterator(nil, nil)
	} else {
		iterator = store.Iterator(nil, nil)
	}
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var storedRecord types.OrderBookRecordData
		if err := k.cdc.Unmarshal(iterator.Value(), &storedRecord); err != nil {
			return sdkerrors.Wrapf(types.ErrInvalidState, "failed to unmarshal OrderBookRecordData, err: %s", err)
		}

		price, orderSeq, err := types.DecodeOrderBookSideRecordKey(iterator.Key())
		if err != nil {
			return err
		}

		stop, err := cb(types.OrderBookRecord{
			// key attributes
			OrderBookID: orderBookID,
			Side:        side,
			Price:       price,
			OrderSeq:    orderSeq,
			// value attributes
			OrderID:           storedRecord.OrderID,
			AccountNumber:     storedRecord.AccountNumber,
			RemainingQuantity: storedRecord.RemainingQuantity,
			RemainingBalance:  storedRecord.RemainingBalance,
		})
		if err != nil {
			return err
		}
		if stop {
			break
		}
	}

	return nil
}

func (k Keeper) getOrGenOrderBookIDs(ctx sdk.Context, baseDenom, quoteDenom string) (uint32, uint32, error) {
	// the function optimizes the read, by writing asc ordered denoms
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

func (k Keeper) genOrderBookIDsFromDenoms(ctx sdk.Context, key0 []byte, denom0, denom1 string) (uint32, error) {
	orderBookID0, err := k.genNextOrderBookID(ctx, key0)
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

func (k Keeper) saveOrderBookData(ctx sdk.Context, orderBookID uint32, data types.OrderBookData) error {
	return k.setDataToStore(ctx, types.CreateOrderBookDataKey(orderBookID), &data)
}

func (k Keeper) getOrderBookData(ctx sdk.Context, orderBookID uint32) (types.OrderBookData, bool, error) {
	var val types.OrderBookData
	found, err := k.getDataFromStore(ctx, types.CreateOrderBookDataKey(orderBookID), &val)
	return val, found, err
}

func (k Keeper) getAccountNumber(ctx sdk.Context, addr string) (uint64, error) {
	accAddr, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return 0, sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", addr)
	}
	acc := k.accountKeeper.GetAccount(ctx, accAddr)
	if acc == nil {
		return 0, sdkerrors.Wrapf(types.ErrInvalidInput, "account not found: %v", addr)
	}

	return acc.GetAccountNumber(), nil
}

func (k Keeper) genNextOrderSeq(ctx sdk.Context) (uint64, error) {
	return k.genNextUint64Seq(ctx, types.OrderSequenceKey)
}

func (k Keeper) saveOrderBookRecordData(
	ctx sdk.Context,
	orderBookID uint32,
	side types.Side,
	price types.Price,
	orderSeq uint64,
	data types.OrderBookRecordData,
) error {
	key, err := types.CreateOrderBookRecordKey(orderBookID, side, price, orderSeq)
	if err != nil {
		return err
	}

	return k.setDataToStore(ctx, key, &data)
}

func (k Keeper) savaOrderData(ctx sdk.Context, orderSeq uint64, data types.OrderData) error {
	return k.setDataToStore(ctx, types.CreateOrderKey(orderSeq), &data)
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
