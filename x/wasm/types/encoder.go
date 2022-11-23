package types

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

// Handler handles encoding of custom message
type Handler func(sender sdk.AccAddress, messages map[string]json.RawMessage) ([]sdk.Msg, error)

// Querier handles custom queries
type Querier func(ctx sdk.Context, queries map[string]json.RawMessage) ([]byte, bool, error)

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

// NewCustomQuerier handles custom queries
func NewCustomQuerier(queriers ...Querier) *wasmkeeper.QueryPlugins {
	return &wasmkeeper.QueryPlugins{
		Custom: func(ctx sdk.Context, request json.RawMessage) ([]byte, error) {
			queries := map[string]json.RawMessage{}
			if err := json.Unmarshal(request, &queries); err != nil {
				return nil, errors.WithStack(err)
			}
			for _, q := range queriers {
				res, ok, err := q(ctx, queries)
				if err != nil {
					return nil, err
				}
				if ok {
					return res, nil
				}
			}
			return nil, errors.New("query not supported")
		},
	}
}
