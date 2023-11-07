package modules

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
)

// BankSendWithdrawPayload generates json containing withdraw payload.
func BankSendWithdrawPayload(amount sdk.Coin, recipient sdk.AccAddress) json.RawMessage {
	return must.Bytes(json.Marshal(map[string]interface{}{
		"amount":    amount.Amount.String(),
		"denom":     amount.Denom,
		"recipient": recipient.String(),
	}))
}

// BankSendExecuteWithdrawRequest generates json with withdraw execution request.
func BankSendExecuteWithdrawRequest(amount sdk.Coin, recipient sdk.AccAddress) json.RawMessage {
	return must.Bytes(json.Marshal(map[string]json.RawMessage{
		"withdraw": BankSendWithdrawPayload(amount, recipient),
	}))
}
