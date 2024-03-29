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

var (
	minPrice = sdk.MustNewDecFromStr("0.000000000000000001")
	maxPrice = sdk.MustNewDecFromStr("999999999999999999.999999999999999999")
)

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateLimitOrder{}, fmt.Sprintf("%s/MsgCreateLimitOrder", ModuleName), nil)
}

// ValidateBasic validates the message.
func (m MsgCreateLimitOrder) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Sender); err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid sender %s", m.Sender)
	}

	if err := m.Amount.Validate(); err != nil {
		return sdkerrors.Wrapf(ErrInvalidCoin, "invalid amount: %s", err.Error())
	}
	if m.Amount.IsZero() {
		return sdkerrors.Wrap(ErrInvalidCoin, "amount must be positive")
	}
	if err := m.SellPrice.Validate(); err != nil {
		return sdkerrors.Wrapf(ErrInvalidPrice, "invalid price: %s", err.Error())
	}
	if m.SellPrice.Amount.LT(minPrice) {
		return sdkerrors.Wrapf(ErrInvalidPrice, "price is lower than: %s", minPrice)
	}
	if m.SellPrice.Amount.GT(maxPrice) {
		return sdkerrors.Wrapf(ErrInvalidPrice, "price is higher than: %s", maxPrice)
	}
	if m.Amount.Denom == m.SellPrice.Denom {
		return sdkerrors.Wrap(ErrInvalidInput, "offered and requested denoms must be different")
	}

	return nil
}

// GetSigners returns the message signers.
func (m MsgCreateLimitOrder) GetSigners() []sdk.AccAddress {
	return []sdk.AccAddress{
		sdk.MustAccAddressFromBech32(m.Sender),
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
