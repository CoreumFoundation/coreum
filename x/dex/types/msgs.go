package types

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type extendedMsg interface {
	sdk.Msg
	sdk.HasValidateBasic
}

var (
	_ extendedMsg = &MsgPlaceOrder{}
	_ extendedMsg = &MsgCancelOrder{}
)

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgPlaceOrder{}, fmt.Sprintf("%s/MsgPlaceOrder", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgCancelOrder{}, fmt.Sprintf("%s/MsgCancelOrder", ModuleName))
}

// ValidateBasic validates the message.
func (m MsgPlaceOrder) ValidateBasic() error {
	if _, err := NewOrderFormMsgPlaceOrder(m); err != nil {
		return err
	}

	return nil
}

// ValidateBasic validates the message.
func (m MsgCancelOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid address: %s", m.Sender)
	}

	return validateOrderID(m.ID)
}
