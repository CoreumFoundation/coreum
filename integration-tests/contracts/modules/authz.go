package modules

import (
	"encoding/json"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
)

// AuthZExecuteTransferRequest generates json with transfer execution request.
func AuthZExecuteTransferRequest(address string, amount sdk.Coin) json.RawMessage {
	return must.Bytes(json.Marshal(map[string]map[string]interface{}{
		"transfer": {
			"address": address,
			"amount":  amount.Amount.String(),
			"denom":   amount.Denom,
		},
	}))
}

// AuthZExecuteStargateRequest generates json with stargate execution request.
func AuthZExecuteStargateRequest(msg sdk.Msg) json.RawMessage {
	msgAny, err := codectypes.NewAnyWithValue(msg)
	if err != nil {
		panic(err)
	}
	return must.Bytes(json.Marshal(map[string]map[string]interface{}{
		"stargate": {
			"type_url": msgAny.TypeUrl,
			"value":    msgAny.Value,
		},
	}))
}
