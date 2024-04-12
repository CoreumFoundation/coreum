package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// RegisterInterfaces registers the asset module tx interfaces.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgIssue{},
		&MsgMint{},
		&MsgBurn{},
		&MsgFreeze{},
		&MsgUnfreeze{},
		&MsgSetFrozen{},
		&MsgGloballyFreeze{},
		&MsgGloballyUnfreeze{},
		&MsgClawback{},
		&MsgSetWhitelistedLimit{},
		&MsgUpgradeTokenV1{},
	)
	registry.RegisterImplementations((*codec.ProtoMarshaler)(nil),
		&DelayedTokenUpgradeV1{},
	)
	registry.RegisterImplementations(
		(*authz.Authorization)(nil),
		&MintAuthorization{},
		&BurnAuthorization{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
