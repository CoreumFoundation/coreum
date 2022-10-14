package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgCreate{}
)

func (msg MsgCreate) ValidateBasic() error {
	// FIXME (wojtek): implement this

	return nil
}

func (msg MsgCreate) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Sender),
	}
}
