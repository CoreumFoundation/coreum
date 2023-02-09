package types

import (
	"context"
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	assetfttype "github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// CustomMsg describes the common object the defines wasm custom message.
type CustomMsg struct {
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload"`
}

// CustomQuery describes the common object the defines wasm custom query.
type CustomQuery struct {
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload"`
}

// NewMsgHandler returns handler that handles messages received from smart contracts.
// The in the input sender is the address of smart contract.
func NewMsgHandler() Handler {
	msgDecoders := make(map[string]func(sender sdk.AccAddress, rawMsg json.RawMessage) (sdk.Msg, error))
	// asset ft handlers
	msgDecoders[protoName(&assetfttype.MsgIssue{})] = func(sender sdk.AccAddress, payload json.RawMessage) (sdk.Msg, error) {
		msg := assetfttype.MsgIssue{}
		if err := decodeMessage(payload, &msg, func(m *assetfttype.MsgIssue) {
			m.Issuer = sender.String()
		}); err != nil {
			return nil, err
		}
		return &msg, nil
	}
	msgDecoders[protoName(&assetfttype.MsgMint{})] = func(sender sdk.AccAddress, payload json.RawMessage) (sdk.Msg, error) {
		msg := assetfttype.MsgMint{}
		if err := decodeMessage(payload, &msg, func(m *assetfttype.MsgMint) {
			m.Sender = sender.String()
		}); err != nil {
			return nil, err
		}
		return &msg, nil
	}
	msgDecoders[protoName(&assetfttype.MsgBurn{})] = func(sender sdk.AccAddress, payload json.RawMessage) (sdk.Msg, error) {
		msg := assetfttype.MsgBurn{}
		if err := decodeMessage(payload, &msg, func(m *assetfttype.MsgBurn) {
			m.Sender = sender.String()
		}); err != nil {
			return nil, err
		}
		return &msg, nil
	}
	msgDecoders[protoName(&assetfttype.MsgFreeze{})] = func(sender sdk.AccAddress, payload json.RawMessage) (sdk.Msg, error) {
		msg := assetfttype.MsgFreeze{}
		if err := decodeMessage(payload, &msg, func(m *assetfttype.MsgFreeze) {
			m.Sender = sender.String()
		}); err != nil {
			return nil, err
		}
		return &msg, nil
	}
	msgDecoders[protoName(&assetfttype.MsgUnfreeze{})] = func(sender sdk.AccAddress, payload json.RawMessage) (sdk.Msg, error) {
		msg := assetfttype.MsgUnfreeze{}
		if err := decodeMessage(payload, &msg, func(m *assetfttype.MsgUnfreeze) {
			m.Sender = sender.String()
		}); err != nil {
			return nil, err
		}
		return &msg, nil
	}
	msgDecoders[protoName(&assetfttype.MsgGloballyFreeze{})] = func(sender sdk.AccAddress, payload json.RawMessage) (sdk.Msg, error) {
		msg := assetfttype.MsgGloballyFreeze{}
		if err := decodeMessage(payload, &msg, func(m *assetfttype.MsgGloballyFreeze) {
			m.Sender = sender.String()
		}); err != nil {
			return nil, err
		}
		return &msg, nil
	}
	msgDecoders[protoName(&assetfttype.MsgGloballyUnfreeze{})] = func(sender sdk.AccAddress, payload json.RawMessage) (sdk.Msg, error) {
		msg := assetfttype.MsgGloballyUnfreeze{}
		if err := decodeMessage(payload, &msg, func(m *assetfttype.MsgGloballyUnfreeze) {
			m.Sender = sender.String()
		}); err != nil {
			return nil, err
		}
		return &msg, nil
	}
	msgDecoders[protoName(&assetfttype.MsgSetWhitelistedLimit{})] = func(sender sdk.AccAddress, payload json.RawMessage) (sdk.Msg, error) {
		msg := assetfttype.MsgSetWhitelistedLimit{}
		if err := decodeMessage(payload, &msg, func(m *assetfttype.MsgSetWhitelistedLimit) {
			m.Sender = sender.String()
		}); err != nil {
			return nil, err
		}
		return &msg, nil
	}

	return func(sender sdk.AccAddress, messages map[string]json.RawMessage) ([]sdk.Msg, error) {
		var res []sdk.Msg

		for _, rawMsg := range messages {
			var customMsg CustomMsg
			if err := json.Unmarshal(rawMsg, &customMsg); err != nil {
				return nil, errors.WithStack(err)
			}

			decoder, ok := msgDecoders[customMsg.Name]
			if !ok {
				return nil, errors.Errorf("handled unknown message type for custom message handler, %s", string(rawMsg))
			}

			msg, err := decoder(sender, customMsg.Payload)
			if err != nil {
				return nil, err
			}

			res = append(res, msg)
		}
		return res, nil
	}
}

// NewQueryHandler returns the handler which handles queries from smart contracts.
func NewQueryHandler(assetFTQueryServer assetfttype.QueryServer) Querier {
	queriers := make(map[string]func(ctx sdk.Context, rawMsg json.RawMessage) ([]byte, bool, error))
	// asset FT queries
	queriers[protoName(&assetfttype.QueryTokenRequest{})] = func(ctx sdk.Context, rawMsg json.RawMessage) ([]byte, bool, error) {
		return processQuery(ctx, rawMsg, &assetfttype.QueryTokenRequest{}, assetFTQueryServer.Token)
	}
	queriers[protoName(&assetfttype.QueryFrozenBalanceRequest{})] = func(ctx sdk.Context, rawMsg json.RawMessage) ([]byte, bool, error) {
		return processQuery(ctx, rawMsg, &assetfttype.QueryFrozenBalanceRequest{}, assetFTQueryServer.FrozenBalance)
	}
	queriers[protoName(&assetfttype.QueryWhitelistedBalanceRequest{})] = func(ctx sdk.Context, rawMsg json.RawMessage) ([]byte, bool, error) {
		return processQuery(ctx, rawMsg, &assetfttype.QueryWhitelistedBalanceRequest{}, assetFTQueryServer.WhitelistedBalance)
	}

	return func(ctx sdk.Context, queries map[string]json.RawMessage) ([]byte, bool, error) {
		for _, rawQuery := range queries {
			var customQuery CustomQuery
			if err := json.Unmarshal(rawQuery, &customQuery); err != nil {
				return nil, false, errors.WithStack(err)
			}

			querier, ok := queriers[customQuery.Name]
			if !ok {
				return nil, false, errors.Errorf("handled unknown message type for custom qeury handler, %s", string(rawQuery))
			}

			return querier(ctx, customQuery.Payload)
		}
		return nil, false, nil
	}
}

func protoName(message proto.Message) string {
	return proto.MessageName(message)
}

func decodeMessage[T sdk.Msg](
	rawMsg json.RawMessage,
	msg T,
	postProcessor func(T),
) error {
	rawTypedMsg, err := decodeEnumStruct(rawMsg)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(rawTypedMsg, msg); err != nil {
		return errors.WithStack(err)
	}
	postProcessor(msg)

	if err := msg.ValidateBasic(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

// The wasm represents custom messages and queries as enums which are represented as object with one attribute in json.
// In our design we set the messages and queries using full message path so the enum name is redundant,
// that's why we take the first value of the map.
func decodeEnumStruct(rawMsg json.RawMessage) (json.RawMessage, error) {
	enumMapping := make(map[string]json.RawMessage)
	if err := json.Unmarshal(rawMsg, &enumMapping); err != nil {
		return nil, errors.WithStack(err)
	}

	if len(enumMapping) != 1 {
		return nil, errors.Errorf("handled unexpected custom wasm message struct: %v", enumMapping)
	}

	// set first map value
	var rawTypedMsg json.RawMessage
	for _, rawTypedMsg = range enumMapping {
		break
	}

	return rawTypedMsg, nil
}

func processQuery[T, K any](
	ctx sdk.Context,
	rawQuery json.RawMessage,
	reqStruct T,
	reqExecutor func(ctx context.Context, req T) (K, error),
) (json.RawMessage, bool, error) {
	rawTypedQuery, err := decodeEnumStruct(rawQuery)
	if err != nil {
		return nil, false, err
	}

	if err := json.Unmarshal(rawTypedQuery, &reqStruct); err != nil {
		return nil, false, errors.WithStack(err)
	}

	res, err := reqExecutor(sdk.WrapSDKContext(ctx), reqStruct)
	if err != nil {
		return nil, false, errors.WithStack(err)
	}

	raw, err := json.Marshal(res)
	if err != nil {
		return nil, false, errors.WithStack(err)
	}
	return raw, true, nil
}
