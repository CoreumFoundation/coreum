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
	if _, err := sdk.AccAddressFromBech32(m.Owner); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid issuer %s", m.Owner)
	}

	if err := m.OfferedAmount.Validate(); err != nil {
		return sdkerrors.Wrapf(ErrInvalidCoin, "invalid offered amount: %s", err.Error())
	}
	if m.OfferedAmount.IsZero() {
		return sdkerrors.Wrap(ErrInvalidCoin, "offered amount must be positive")
	}
	if err := m.SellPrice.Validate(); err != nil {
		return sdkerrors.Wrapf(ErrInvalidPrice, "invalid price: %s", err.Error())
	}
	if m.SellPrice.IsZero() {
		return sdkerrors.Wrap(ErrInvalidPrice, "sell price must be positive")
	}
	if m.OfferedAmount.Denom == m.SellPrice.Denom {
		return sdkerrors.Wrap(ErrInvalidInput, "offered and requested denoms must be different")
	}

	return nil
}

// GetSigners returns the message signers.
func (m MsgCreateLimitOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Owner),
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
