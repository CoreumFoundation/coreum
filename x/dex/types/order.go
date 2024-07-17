package types

import (
	"math/big"
	"math/bits"
	"regexp"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	cbig "github.com/CoreumFoundation/coreum/v4/pkg/math/big"
)

const (
	// maxWordLen defines the maximum word length supported by Int and Uint types.
	maxSDKIntWordLen = sdkmath.MaxBitLen / bits.UintSize
)

var (
	orderIDRegexStr = `^[a-zA-Z0-9/+:._-]{1,40}$`
	orderIDRegex    *regexp.Regexp
)

func init() {
	orderIDRegex = regexp.MustCompile(orderIDRegexStr)
}

// Opposite returns opposite side.
func (s Side) Opposite() (Side, error) {
	switch s {
	case Side_sell:
		return Side_buy, nil
	case Side_buy:
		return Side_sell, nil
	default:
		return 0, sdkerrors.Wrapf(ErrInvalidInput, "invalid side: %s", s)
	}
}

// NewOrderFormMsgPlaceOrder creates and validates Order from MsgPlaceOrder.
func NewOrderFormMsgPlaceOrder(msg MsgPlaceOrder) (Order, error) {
	o := Order{
		Account:    msg.Sender,
		ID:         msg.ID,
		BaseDenom:  msg.BaseDenom,
		QuoteDenom: msg.QuoteDenom,
		Price:      msg.Price,
		Quantity:   msg.Quantity,
		Side:       msg.Side,
	}
	if err := o.Validate(); err != nil {
		return Order{}, err
	}

	return o, nil
}

// Validate validates order object.
func (o Order) Validate() error {
	if _, err := sdk.AccAddressFromBech32(o.Account); err != nil {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid address: %s", o.Account)
	}

	if !orderIDRegex.MatchString(o.ID) {
		return sdkerrors.Wrapf(ErrInvalidInput, "order ID must match regex format '%s'", orderIDRegex)
	}

	if o.BaseDenom == "" {
		return sdkerrors.Wrap(ErrInvalidInput, "base denom can't be empty")
	}

	if o.QuoteDenom == "" {
		return sdkerrors.Wrap(ErrInvalidInput, "quote denom can't be empty")
	}

	if !o.Quantity.IsPositive() {
		return sdkerrors.Wrap(ErrInvalidInput, "quantity must be positive")
	}

	switch o.Side {
	case Side_sell, Side_buy:
		// ignore
	default:
		return sdkerrors.Wrapf(ErrInvalidInput, "only %s and %s sides are allowed", o.Side.String(), o.Side.String())
	}

	if !o.RemainingQuantity.IsNil() {
		return sdkerrors.Wrap(ErrInvalidInput, "initial remaining quantity must be nil")
	}

	if !o.RemainingBalance.IsNil() {
		return sdkerrors.Wrap(ErrInvalidInput, "initial remaining balance must be nil")
	}

	if _, err := o.ComputeLockedBalance(); err != nil {
		return err
	}

	return nil
}

// ComputeLockedBalance computes the balance locked for the order.
func (o Order) ComputeLockedBalance() (sdk.Coin, error) {
	if o.Side == Side_buy {
		balance, remainder := cbig.IntMulRatWithRemainder(o.Quantity.BigInt(), o.Price.Rat())
		if !cbig.IntEqZero(remainder) {
			return sdk.Coin{}, sdkerrors.Wrapf(
				ErrInvalidInput,
				"quantity multiplied by price must be an integer, for %s side",
				Side_buy.String(),
			)
		}
		if isBigIntOverflowsSDKInt(balance) {
			return sdk.Coin{}, sdkerrors.Wrapf(
				ErrInvalidInput,
				"invalid order quantity and price, order balance is out of supported sdkmath.Int range",
			)
		}
		return sdk.NewCoin(o.QuoteDenom, sdkmath.NewIntFromBigInt(balance)), nil
	}

	return sdk.NewCoin(o.BaseDenom, o.Quantity), nil
}

// GetLockedBalanceDenom returns locked balance denom.
func (o Order) GetLockedBalanceDenom() string {
	if o.Side == Side_buy {
		return o.QuoteDenom
	}

	return o.BaseDenom
}

// isBigIntOverflowsSDKInt checks if the big int overflows the sdkmath.Int.
// copy form sdkmath.Int.
func isBigIntOverflowsSDKInt(i *big.Int) bool {
	if len(i.Bits()) > maxSDKIntWordLen {
		return i.BitLen() > sdkmath.MaxBitLen
	}
	return false
}
