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
	case SIDE_SELL:
		return SIDE_BUY, nil
	case SIDE_BUY:
		return SIDE_SELL, nil
	default:
		return 0, sdkerrors.Wrapf(ErrInvalidInput, "invalid side: %s", s)
	}
}

// Validate validates order side.
func (s Side) Validate() error {
	switch s {
	case SIDE_SELL, SIDE_BUY:
		return nil
	default:
		return sdkerrors.Wrapf(ErrInvalidInput, "only %s and %s sides are allowed", s.String(), s.String())
	}
}

// NewOrderFormMsgPlaceOrder creates and validates Order from MsgPlaceOrder.
func NewOrderFormMsgPlaceOrder(msg MsgPlaceOrder) (Order, error) {
	o := Order{
		Creator:    msg.Sender,
		Type:       msg.Type,
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
	if _, err := sdk.AccAddressFromBech32(o.Creator); err != nil {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid address: %s", o.Creator)
	}

	if err := validateOrderID(o.ID); err != nil {
		return err
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

	if err := o.Side.Validate(); err != nil {
		return err
	}

	switch o.Type {
	case ORDER_TYPE_LIMIT:
		if o.Price == nil {
			return sdkerrors.Wrap(
				ErrInvalidInput, "price must be not nil for the limit order",
			)
		}
		if _, err := o.ComputeLimitOrderLockedBalance(); err != nil {
			return err
		}
	case ORDER_TYPE_MARKET:
		if o.Price != nil {
			return sdkerrors.Wrap(
				ErrInvalidInput, "price must be nil for the limit order",
			)
		}
	default:
		return sdkerrors.Wrapf(
			ErrInvalidInput, "unsupported order type : %s", o.Type.String(),
		)
	}

	if !o.RemainingQuantity.IsNil() {
		return sdkerrors.Wrap(ErrInvalidInput, "initial remaining quantity must be nil")
	}

	if !o.RemainingBalance.IsNil() {
		return sdkerrors.Wrap(ErrInvalidInput, "initial remaining balance must be nil")
	}

	return nil
}

// ComputeLimitOrderLockedBalance computes the order locked balance.
func (o Order) ComputeLimitOrderLockedBalance() (sdk.Coin, error) {
	if o.Side == SIDE_BUY {
		balance, remainder := cbig.IntMulRatWithRemainder(o.Quantity.BigInt(), o.Price.Rat())
		if !cbig.IntEqZero(remainder) {
			balance = cbig.IntAdd(balance, big.NewInt(1))
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

// GetBalanceDenom returns order balance denom.
func (o Order) GetBalanceDenom() string {
	if o.Side == SIDE_BUY {
		return o.QuoteDenom
	}

	return o.BaseDenom
}

// GetOppositeFromBalanceDenom returns the order denom which is not balance denom.
func (o Order) GetOppositeFromBalanceDenom() string {
	if o.BaseDenom == o.GetBalanceDenom() {
		return o.QuoteDenom
	}
	return o.BaseDenom
}

func validateOrderID(id string) error {
	if !orderIDRegex.MatchString(id) {
		return sdkerrors.Wrapf(ErrInvalidInput, "order ID must match regex format '%s'", orderIDRegex)
	}
	return nil
}

// isBigIntOverflowsSDKInt checks if the big int overflows the sdkmath.Int.
// copy form sdkmath.Int.
func isBigIntOverflowsSDKInt(i *big.Int) bool {
	if len(i.Bits()) > maxSDKIntWordLen {
		return i.BitLen() > sdkmath.MaxBitLen
	}
	return false
}
