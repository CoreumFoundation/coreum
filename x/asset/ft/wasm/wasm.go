package wasm

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/asset/ft/keeper"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
	wasmtypes "github.com/CoreumFoundation/coreum/x/wasm/types"
)

// MsgHandler handles conversion of messages received from smart contracts.
func MsgHandler(sender sdk.AccAddress, messages map[string]json.RawMessage) ([]sdk.Msg, error) {
	var res []sdk.Msg
	for msgType, msg := range messages {
		if msgType == "AssetFTMsgIssue" {
			var msgIssue types.MsgIssue
			if err := json.Unmarshal(msg, &msgIssue); err != nil {
				return nil, errors.WithStack(err)
			}
			msgIssue.Issuer = sender.String() // sender is the address of smart contract
			res = append(res, &msgIssue)
		}
	}
	return res, nil
}

// QueryHandler handles queries from smart contracts.
func QueryHandler(keeper keeper.Keeper) wasmtypes.Querier {
	return func(ctx sdk.Context, queries map[string]json.RawMessage) ([]byte, bool, error) {
		for qType, q := range queries {
			if qType == "AssetFTGetToken" {
				qToken := struct {
					Denom string `json:"denom"`
				}{}
				if err := json.Unmarshal(q, &qToken); err != nil {
					return nil, false, errors.WithStack(err)
				}

				ft, err := keeper.GetToken(ctx, qToken.Denom)
				if err != nil {
					return nil, false, errors.WithStack(err)
				}

				raw, err := json.Marshal(ft)
				if err != nil {
					return nil, false, errors.WithStack(err)
				}
				return raw, true, nil
			}
		}
		return nil, false, nil
	}
}
