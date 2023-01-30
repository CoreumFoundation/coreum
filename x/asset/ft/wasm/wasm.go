package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
	wasmtypes "github.com/CoreumFoundation/coreum/x/wasm/types"
)

const (
	modulePrefix = "AssetFT"
	msgPrefix    = modulePrefix
	queryPrefix  = modulePrefix + "Query"
)

var (
	// msg names.
	msgIssueName               = buildMsgName(&types.MsgIssue{})
	msgMintName                = buildMsgName(&types.MsgMint{})
	msgBurnName                = buildMsgName(&types.MsgBurn{})
	msgFreezeName              = buildMsgName(&types.MsgFreeze{})
	msgUnfreezeName            = buildMsgName(&types.MsgUnfreeze{})
	msgGloballyFreezeName      = buildMsgName(&types.MsgGloballyFreeze{})
	msgGloballyUnfreezeName    = buildMsgName(&types.MsgGloballyUnfreeze{})
	msgSetWhitelistedLimitName = buildMsgName(&types.MsgSetWhitelistedLimit{})
	// query names.
	queryTokenName              = buildQueryName("Token")
	queryFrozenBalanceName      = buildQueryName("FrozenBalance")
	queryWhitelistedBalanceName = buildQueryName("WhitelistedBalance")
)

// MsgHandler handles conversion of messages received from smart contracts.
// The in the input sender is the address of smart contract.
func MsgHandler(sender sdk.AccAddress, messages map[string]json.RawMessage) ([]sdk.Msg, error) {
	var res []sdk.Msg
	for msgType, rawMsg := range messages {
		switch msgType {
		case msgIssueName:
			msg := types.MsgIssue{}
			if err := decodeMessage(rawMsg, &msg, func(m *types.MsgIssue) {
				m.Issuer = sender.String()
			}); err != nil {
				return nil, err
			}
			res = append(res, &msg)
		case msgMintName:
			msg := types.MsgMint{}
			if err := decodeMessage(rawMsg, &msg, func(m *types.MsgMint) {
				m.Sender = sender.String()
			}); err != nil {
				return nil, err
			}
			res = append(res, &msg)
		case msgBurnName:
			msg := types.MsgBurn{}
			if err := decodeMessage(rawMsg, &msg, func(m *types.MsgBurn) {
				m.Sender = sender.String()
			}); err != nil {
				return nil, err
			}
			res = append(res, &msg)
		case msgFreezeName:
			msg := types.MsgFreeze{}
			if err := decodeMessage(rawMsg, &msg, func(m *types.MsgFreeze) {
				m.Sender = sender.String()
			}); err != nil {
				return nil, err
			}
			res = append(res, &msg)
		case msgUnfreezeName:
			msg := types.MsgUnfreeze{}
			if err := decodeMessage(rawMsg, &msg, func(m *types.MsgUnfreeze) {
				m.Sender = sender.String()
			}); err != nil {
				return nil, err
			}
			res = append(res, &msg)
		case msgGloballyFreezeName:
			msg := types.MsgGloballyFreeze{}
			if err := decodeMessage(rawMsg, &msg, func(m *types.MsgGloballyFreeze) {
				m.Sender = sender.String()
			}); err != nil {
				return nil, err
			}
			res = append(res, &msg)
		case msgGloballyUnfreezeName:
			msg := types.MsgGloballyUnfreeze{}
			if err := decodeMessage(rawMsg, &msg, func(m *types.MsgGloballyUnfreeze) {
				m.Sender = sender.String()
			}); err != nil {
				return nil, err
			}
			res = append(res, &msg)
		case msgSetWhitelistedLimitName:
			msg := types.MsgSetWhitelistedLimit{}
			if err := decodeMessage(rawMsg, &msg, func(m *types.MsgSetWhitelistedLimit) {
				m.Sender = sender.String()
			}); err != nil {
				return nil, err
			}
			res = append(res, &msg)
		}
	}
	return res, nil
}

// QueryHandler handles queries from smart contracts.
func QueryHandler(queryServer types.QueryServer) wasmtypes.Querier {
	return func(ctx sdk.Context, queries map[string]json.RawMessage) ([]byte, bool, error) {
		for queryType, rawQuery := range queries {
			switch queryType {
			case queryTokenName:
				return processQuery(ctx, rawQuery, &types.QueryTokenRequest{}, queryServer.Token)
			case queryFrozenBalanceName:
				return processQuery(ctx, rawQuery, &types.QueryFrozenBalanceRequest{}, queryServer.FrozenBalance)
			case queryWhitelistedBalanceName:
				return processQuery(ctx, rawQuery, &types.QueryWhitelistedBalanceRequest{}, queryServer.WhitelistedBalance)
			}
		}
		return nil, false, nil
	}
}

func buildMsgName(message proto.Message) string {
	parts := strings.Split(proto.MessageName(message), ".")
	last := parts[len(parts)-1]

	return fmt.Sprintf("%s%s", msgPrefix, last)
}

func buildQueryName(name string) string {
	return fmt.Sprintf("%s%s", queryPrefix, name)
}

func decodeMessage[T sdk.Msg](
	rawMsg json.RawMessage,
	msg T,
	postProcessor func(T),
) error {
	if err := json.Unmarshal(rawMsg, msg); err != nil {
		return errors.WithStack(err)
	}
	postProcessor(msg)
	if err := msg.ValidateBasic(); err != nil {
		return errors.WithStack(err)
	}

	return nil
}

func processQuery[T, K any](
	ctx sdk.Context,
	rawQuery json.RawMessage,
	reqStruct T,
	reqExecutor func(ctx context.Context, req T) (K, error),
) (json.RawMessage, bool, error) {
	if err := json.Unmarshal(rawQuery, &reqStruct); err != nil {
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
