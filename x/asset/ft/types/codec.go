package types

import (
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/gogo/protobuf/proto"
)

// RegisterInterfaces registers the asset module tx interfaces.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgIssue{},
		&MsgMint{},
		&MsgBurn{},
		&MsgFreeze{},
		&MsgUnfreeze{},
		&MsgGloballyFreeze{},
		&MsgGloballyUnfreeze{},
		&MsgSetWhitelistedLimit{},
		&MsgTokenUpgradeV1{},
	)
	registry.RegisterImplementations((*proto.Message)(nil),
		&DelayedTokenUpgradeV1{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
