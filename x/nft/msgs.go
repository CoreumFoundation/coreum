package nft

import (
	"fmt"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
)

const (
	// TypeMsgSend nft message types.
	TypeMsgSend = "send"
)

var (
	_ sdk.Msg            = &MsgSend{}
	_ legacytx.LegacyMsg = &MsgSend{}
)

// RegisterLegacyAminoCodec registers the amino types and interfaces.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSend{}, fmt.Sprintf("%s/MsgSend", ModuleName), nil)
}

// ValidateBasic implements the Msg.ValidateBasic method.
func (m MsgSend) ValidateBasic() error {
	if err := ValidateClassID(m.ClassId); err != nil {
		return sdkerrors.Wrapf(ErrInvalidID, "invalid class id (%s)", m.ClassId)
	}

	if err := ValidateNFTID(m.Id); err != nil {
		return sdkerrors.Wrapf(ErrInvalidID, "invalid nft id (%s)", m.Id)
	}

	_, err := sdk.AccAddressFromBech32(m.Sender)
	if err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid sender address (%s)", m.Sender)
	}

	_, err = sdk.AccAddressFromBech32(m.Receiver)
	if err != nil {
		return sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid receiver address (%s)", m.Receiver)
	}
	return nil
}

// GetSigners implements Msg.
func (m MsgSend) GetSigners() []sdk.AccAddress {
	signer, _ := sdk.AccAddressFromBech32(m.Sender)
	return []sdk.AccAddress{signer}
}

// GetSignBytes returns sign bytes for LegacyMsg.
func (m MsgSend) GetSignBytes() []byte {
	return sdk.MustSortJSON(moduleAminoCdc.MustMarshalJSON(&m))
}

// Route returns message route for LegacyMsg.
func (m MsgSend) Route() string {
	return RouterKey
}

// Type returns message type for LegacyMsg.
func (m MsgSend) Type() string {
	return TypeMsgSend
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
