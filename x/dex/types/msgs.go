package types

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/legacy"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type extendedMsg interface {
	sdk.Msg
	sdk.HasValidateBasic
}

var (
	_ extendedMsg = &MsgUpdateParams{}
	_ extendedMsg = &MsgPlaceOrder{}
	_ extendedMsg = &MsgCancelOrder{}
)

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgPlaceOrder{}, fmt.Sprintf("%s/MsgPlaceOrder", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgCancelOrder{}, fmt.Sprintf("%s/MsgCancelOrder", ModuleName))
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, fmt.Sprintf("%s/MsgUpdateParams", ModuleName))
}

// ValidateBasic checks that message fields are valid.
func (m MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return cosmoserrors.ErrInvalidAddress.Wrapf("invalid authority address: %s", err)
	}

	return m.Params.ValidateBasic()
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
