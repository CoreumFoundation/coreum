package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgUnfreezeCoin = "unfreeze_coin"

var _ sdk.Msg = &MsgUnfreezeCoin{}

func NewMsgUnfreezeCoin(creator string, address string, coin sdk.Coin) *MsgUnfreezeCoin {
	return &MsgUnfreezeCoin{
		Creator: creator,
		Address: address,
		Coin:    coin,
	}
}

func (msg *MsgUnfreezeCoin) Route() string {
	return RouterKey
}

func (msg *MsgUnfreezeCoin) Type() string {
	return TypeMsgUnfreezeCoin
}

func (msg *MsgUnfreezeCoin) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUnfreezeCoin) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUnfreezeCoin) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
