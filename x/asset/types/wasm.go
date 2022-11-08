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
			var createFungibleTokenMsg MsgIssueFungibleToken
			if err := json.Unmarshal(msg, &createFungibleTokenMsg); err != nil {
				return nil, errors.WithStack(err)
			}
			createFungibleTokenMsg.Issuer = sender.String()
			res = append(res, &createFungibleTokenMsg)
		}
	}
	return res, nil
}
