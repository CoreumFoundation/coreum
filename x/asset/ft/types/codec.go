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
		&MsgTransferAdmin{},
		&MsgClearAdmin{},
		&MsgSetWhitelistedLimit{},
	)
	registry.RegisterImplementations((*proto.Message)(nil),
		&DelayedTokenUpgradeV1{},
	)
	registry.RegisterImplementations(
		(*authz.Authorization)(nil),
		&MintAuthorization{},
		&BurnAuthorization{},
	)
	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
