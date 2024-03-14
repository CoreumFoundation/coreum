package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

// Type of messages for amino.
const (
	TypeMsgCreateLimitOrder = "create-limit-order"
)

var (
	_ sdk.Msg            = &MsgCreateLimitOrder{}
	_ legacytx.LegacyMsg = &MsgCreateLimitOrder{}
)

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateLimitOrder{}, fmt.Sprintf("%s/MsgCreateLimitOrder", ModuleName), nil)
}

// ValidateBasic validates the message.
func (m MsgCreateLimitOrder) ValidateBasic() error {
	// TODO: Implement
	return nil
}

// GetSigners returns the message signers.
func (m MsgCreateLimitOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Issuer),
	}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgCreateLimitOrder) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgCreateLimitOrder) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgCreateLimitOrder) Type() string {
	return TypeMsgCreateLimitOrder
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
