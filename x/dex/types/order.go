package types

import (
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Order defines order methods required by the keeper.
type Order interface {
	codec.ProtoMarshaler
	Account() string
	DenomOffered() string
	DenomRequested() string
	AmountOffered() sdkmath.Int
	Price() sdk.Dec
	String() string

	ReduceOfferedAmount(reduceAmount sdkmath.Int)
}

// Account returns the account who placed the order.
func (o *OrderLimit) Account() string {
	return o.Sender
}

// DenomOffered returns the offered denom.
func (o *OrderLimit) DenomOffered() string {
	return o.Amount.Denom
}

// DenomRequested returns the requested denom.
func (o *OrderLimit) DenomRequested() string {
	return o.SellPrice.Denom
}

// AmountOffered returns the offered amount.
func (o *OrderLimit) AmountOffered() sdkmath.Int {
	return o.Amount.Amount
}

// Price returns the sell price.
func (o *OrderLimit) Price() sdk.Dec {
	return o.SellPrice.Amount
}

// ReduceOfferedAmount reduces offered amount.
func (o *OrderLimit) ReduceOfferedAmount(reduceAmount sdkmath.Int) {
	o.Amount.Amount = o.Amount.Amount.Sub(reduceAmount)
}
