package ibc

import sdk "github.com/cosmos/cosmos-sdk/types"

// SimpleStateMethod is a type used to represent the methods available inside the simple state contract.
type HooksMethod string

const (
	HooksGetCount HooksMethod = "get_count"

	HooksGetTotalFunds HooksMethod = "get_total_funds"
)

type HooksCounterState struct {
	Count int `json:"count"`
}

type HooksTotalFundsState struct {
	TotalFunds sdk.Coins `json:"total_funds"`
}

type HooksBodyRequest struct {
	Addr string `json:"addr"`
}
