package types

import (
	"regexp"

	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	orderIDRegexStr = `^[a-zA-Z0-9/+:._-]{1,40}$`
	orderIDRegex    *regexp.Regexp
)

func init() {
	orderIDRegex = regexp.MustCompile(orderIDRegexStr)
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

	return nil
}
