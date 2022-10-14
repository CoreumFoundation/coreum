package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgCreate{}
	_ sdk.Msg = &MsgClaim{}
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

func (msg MsgClaim) ValidateBasic() error {
	// FIXME (wojtek): implement this

	return nil
}

func (msg MsgClaim) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(msg.Recipient),
	}
}
