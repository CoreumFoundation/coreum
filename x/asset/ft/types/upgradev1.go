package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/gogo/protobuf/proto"

	delaytypes "github.com/CoreumFoundation/coreum/x/delay/types"
)

var _ sdk.Msg = &MsgTokenUpgradeV1{}

// ValidateBasic checks that message fields are valid.
func (msg MsgTokenUpgradeV1) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Sender); err != nil {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, "invalid sender address")
	}

	return nil
}

// GetSigners returns the required signers of this message type.
func (msg MsgTokenUpgradeV1) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}

// UpgradeV1Keeper defines methods required to update tokens to V1.
type UpgradeV1Keeper interface {
	UpgradeTokenToV1(ctx sdk.Context, data *DelayedTokenUpgradeV1) error
}

// NewTokenUpgradeV1Handler handles token V1 upgrade.
func NewTokenUpgradeV1Handler(keeper UpgradeV1Keeper) delaytypes.Handler {
	return func(ctx sdk.Context, data proto.Message) error {
		return keeper.UpgradeTokenToV1(ctx, data.(*DelayedTokenUpgradeV1))
	}
}
