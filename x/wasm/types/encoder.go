package types

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

// Handler handles encoding of custom message
type Handler func(sender sdk.AccAddress, messages map[string]json.RawMessage) ([]sdk.Msg, error)

// NewCustomEncoder encodes custom messages received from smart contracts
func NewCustomEncoder(handlers ...Handler) *wasmkeeper.MessageEncoders {
	return &wasmkeeper.MessageEncoders{
		Custom: func(sender sdk.AccAddress, msg json.RawMessage) ([]sdk.Msg, error) {
			messages := map[string]json.RawMessage{}
			if err := json.Unmarshal(msg, &messages); err != nil {
				return nil, errors.WithStack(err)
			}
			res := make([]sdk.Msg, 0, len(messages))

			for _, h := range handlers {
				msgs, err := h(sender, messages)
				if err != nil {
					return nil, err
				}
				res = append(res, msgs...)
			}
			return res, nil
		},
	}
}
