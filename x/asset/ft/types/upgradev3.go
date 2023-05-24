package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"

	delaytypes "github.com/CoreumFoundation/coreum/x/delay/types"
)

var _ sdk.Msg = &MsgEnableIBCRequest{}

// ValidateBasic checks that message fields are valid.
func (msg MsgEnableIBCRequest) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	return nil
}

// GetSigners returns the required signers of this message type.
func (msg MsgEnableIBCRequest) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}

// UpgradeV3Keeper defines methods required from keeper managing v3 upgrade.
type UpgradeV3Keeper interface {
	EnableIBC(ctx sdk.Context, denom string) error
}

// NewEnableIBCHandler enables IBC for the token.
func NewEnableIBCHandler(keeper UpgradeV3Keeper) delaytypes.Handler {
	return func(ctx sdk.Context, msg proto.Message) error {
		return keeper.EnableIBC(ctx, msg.(*MsgEnableIBCExecutor).Denom)
	}
}
