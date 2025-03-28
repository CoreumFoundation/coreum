package types

import (
	"encoding/json"

	cosmossdk_io_math "cosmossdk.io/math"
)

// OrderBookRecord is a single order book record, it combines both key and value from the store.
//
//nolint:tagliatelle
type OrderBookRecord struct {
	// order_book_id is order book ID.
	OrderBookID uint32 `json:"order_book_id,omitempty"`
	// side is order side.
	Side Side `json:"side,omitempty"`
	// price is order book record price.
	Price Price `json:"price"`
	// order_sequence is order sequence.
	OrderSequence uint64 `json:"order_sequence,omitempty"`
	// order ID provided by the creator.
	OrderID string `json:"order_id,omitempty"`
	// account_number is account number which corresponds the order creator.
	AccountNumber uint64 `json:"account_number,omitempty"`
	// remaining_base_quantity - is remaining quantity of base denom which user wants to sell or buy.
	RemainingBaseQuantity cosmossdk_io_math.Int `json:"remaining_base_quantity"`
	// remaining_spendable_balance - is balance up to which user wants to spend to execute the order.
	RemainingSpendableBalance cosmossdk_io_math.Int `json:"remaining_spendable_balance"`
}

func (o *OrderBookRecord) String() string {
	serialized, err := json.Marshal(o)
	if err != nil {
		return err.Error()
	}
	return string(serialized)
}
