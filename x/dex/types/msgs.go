package types

import (
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
	_ extendedMsg = &MsgCancelOrdersByDenom{}
)

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	legacy.RegisterAminoMsg(cdc, &MsgPlaceOrder{}, ModuleName+"/MsgPlaceOrder")
	legacy.RegisterAminoMsg(cdc, &MsgCancelOrder{}, ModuleName+"/MsgCancelOrder")
	legacy.RegisterAminoMsg(cdc, &MsgUpdateParams{}, ModuleName+"/MsgUpdateParams")
	legacy.RegisterAminoMsg(cdc, &MsgCancelOrdersByDenom{}, ModuleName+"/MsgCancelOrdersByDenom")
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

// ValidateBasic validates the message.
func (m MsgCancelOrdersByDenom) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid address: %s", m.Sender)
	}

	if _, err := sdk.AccAddressFromBech32(m.Account); err != nil {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid address: %s", m.Account)
	}

	if err := sdk.ValidateDenom(m.Denom); err != nil {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid denom: %s", m.Denom)
	}

	return nil
}
