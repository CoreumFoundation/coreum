package modules

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
)

type bankMethod string

const (
	withdraw bankMethod = "withdraw"
)

type bankWithdrawRequest struct {
	Amount    string `json:"amount,omitempty"`
	Denom     string `json:"denom,omitempty"`
	Recipient string `json:"recipient,omitempty"`
}

// WithdrawPayload generates json containing withdraw payload.
func WithdrawPayload(amount sdk.Coin, recipient sdk.AccAddress) json.RawMessage {
	return must.Bytes(json.Marshal(bankWithdrawRequest{
		Amount:    amount.Amount.String(),
		Denom:     amount.Denom,
		Recipient: recipient.String(),
	}))
}

// ExecuteWithdrawRequest generates json with withdraw execution request.
func ExecuteWithdrawRequest(amount sdk.Coin, recipient sdk.AccAddress) json.RawMessage {
	return must.Bytes(json.Marshal(map[bankMethod]json.RawMessage{
		withdraw: WithdrawPayload(amount, recipient),
	}))
}
