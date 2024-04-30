package types

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

// Type of messages for amino.
const (
	TypeMsgPlaceOrder = "place-order"
)

var (
	_ sdk.Msg            = &MsgPlaceOrder{}
	_ legacytx.LegacyMsg = &MsgPlaceOrder{}
)

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgPlaceOrder{}, fmt.Sprintf("%s/MsgPlaceOrder", ModuleName), nil)
}

// ValidateBasic validates the message.
func (m MsgPlaceOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid sender %s", m.Sender)
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
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgPlaceOrder) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgPlaceOrder) Type() string {
	return TypeMsgPlaceOrder
}

var (
	amino          = codec.NewLegacyAmino()
	moduleAminoCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}
