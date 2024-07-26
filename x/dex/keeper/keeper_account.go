package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func (k Keeper) getAccountNumber(ctx sdk.Context, addr string) (uint64, error) {
	accAddr, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return 0, sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", addr)
	}
	acc := k.accountKeeper.GetAccount(ctx, accAddr)
	if acc == nil {
		return 0, sdkerrors.Wrapf(types.ErrInvalidInput, "account not found: %v", addr)
	}

	return acc.GetAccountNumber(), nil
}

func (k Keeper) getAccountAddress(ctx sdk.Context, accountNumber uint64) (sdk.AccAddress, error) {
	addr := k.accountKeeper.GetAccountAddressByID(ctx, accountNumber)

	acc, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", addr)
	}

	return acc, nil
}
