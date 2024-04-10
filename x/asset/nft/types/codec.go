package types

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/cosmos/gogoproto/proto"
)

// RegisterInterfaces registers the asset module tx interfaces.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterInterface(proto.MessageName((*DataBytes)(nil)), (*proto.Message)(nil), (*DataBytes)(nil))
	registry.RegisterInterface(proto.MessageName((*DataDynamic)(nil)), (*proto.Message)(nil), (*DataDynamic)(nil))
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgIssueClass{},
		&MsgMint{},
		&MsgBurn{},
		&MsgFreeze{},
		&MsgUnfreeze{},
		&MsgAddToWhitelist{},
		&MsgRemoveFromWhitelist{},
		&MsgAddToClassWhitelist{},
		&MsgRemoveFromClassWhitelist{},
		&MsgClassFreeze{},
		&MsgClassUnfreeze{},
	)
	registry.RegisterImplementations(
		(*authz.Authorization)(nil),
		&SendAuthorization{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
