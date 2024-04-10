package ibc

import sdk "github.com/cosmos/cosmos-sdk/types"

// HooksMethod is a type used to represent the methods available inside ibc-hooks-counter contract.
type HooksMethod string

const (
	// HooksGetCount is a method used to get current counter for an address.
	HooksGetCount HooksMethod = "get_count"
	// HooksGetTotalFunds is a method used go get total funds transferred by address.
	HooksGetTotalFunds HooksMethod = "get_total_funds"
)

// HooksBodyRequest is a query request for get_count & get_total_funds.
type HooksBodyRequest struct {
	Addr string `json:"addr"`
}

// HooksCounterState is a struct used to initialize ibc-hooks-counter contract
// and also represents response returned from get_count.
type HooksCounterState struct {
	Count int `json:"count"`
}

// HooksTotalFundsState is a structure that represents response returned from get_total_funds.
//
//nolint:tagliatelle // wasm requirements
type HooksTotalFundsState struct {
	TotalFunds sdk.Coins `json:"total_funds"`
}
