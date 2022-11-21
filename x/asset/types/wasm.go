package types

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

// WASMHandler handles conversion of messages received from smart contracts
func WASMHandler(sender sdk.AccAddress, messages map[string]json.RawMessage) ([]sdk.Msg, error) {
	var res []sdk.Msg
	for msgType, msg := range messages {
		if msgType == "MsgIssueFungibleToken" {
			var msgIssueFungibleToken MsgIssueFungibleToken
			if err := json.Unmarshal(msg, &msgIssueFungibleToken); err != nil {
				return nil, errors.WithStack(err)
			}
			msgIssueFungibleToken.Issuer = sender.String()
			res = append(res, &msgIssueFungibleToken)
		}
	}
	return res, nil
}
