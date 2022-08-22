package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgFreezeCoin = "freeze_coin"

var _ sdk.Msg = &MsgFreezeCoin{}

func NewMsgFreezeCoin(creator string, address string, coin sdk.Coin) *MsgFreezeCoin {
	return &MsgFreezeCoin{
		Creator: creator,
		Address: address,
		Coin:    coin,
	}
}

func (msg *MsgFreezeCoin) Route() string {
	return RouterKey
}

func (msg *MsgFreezeCoin) Type() string {
	return TypeMsgFreezeCoin
}

func (msg *MsgFreezeCoin) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgFreezeCoin) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgFreezeCoin) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
