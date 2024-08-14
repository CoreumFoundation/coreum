package types

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

// Type of messages for amino.
const (
	TypeMsgPlaceOrder  = "place-order"
	TypeMsgCancelOrder = "cancel-order"
)

var (
	_ sdk.Msg            = &MsgPlaceOrder{}
	_ legacytx.LegacyMsg = &MsgPlaceOrder{}
	_ sdk.Msg            = &MsgCancelOrder{}
	_ legacytx.LegacyMsg = &MsgCancelOrder{}
)

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgPlaceOrder{}, fmt.Sprintf("%s/MsgPlaceOrder", ModuleName), nil)
	cdc.RegisterConcrete(&MsgCancelOrder{}, fmt.Sprintf("%s/MsgCancelOrder", ModuleName), nil)
}

// ValidateBasic validates the message.
func (m MsgPlaceOrder) ValidateBasic() error {
	if _, err := NewOrderFormMsgPlaceOrder(m); err != nil {
		return err
	}

	return nil
}

// GetSigners returns the message signers.
func (m MsgPlaceOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Sender),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgPlaceOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(amino.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgPlaceOrder) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgPlaceOrder) Type() string {
	return TypeMsgPlaceOrder
}

// ValidateBasic validates the message.
func (m MsgCancelOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(ErrInvalidInput, "invalid address: %s", m.Sender)
	}

	return validateOrderID(m.ID)
}

// GetSigners returns the message signers.
func (m MsgCancelOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Sender),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgCancelOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(amino.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgCancelOrder) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgCancelOrder) Type() string {
	return TypeMsgCancelOrder
}

var amino = codec.NewLegacyAmino()

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
