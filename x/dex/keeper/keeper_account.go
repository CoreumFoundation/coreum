package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

func (k Keeper) getAccountNumber(ctx sdk.Context, addr sdk.AccAddress) (uint64, error) {
	acc := k.accountKeeper.GetAccount(ctx, addr)
	if acc == nil {
		return 0, sdkerrors.Wrapf(types.ErrInvalidInput, "account not found: %s", addr.String())
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

func (k Keeper) getAccountAddressWithCache(ctx sdk.Context, accountNumber uint64, cache map[uint64]sdk.AccAddress) (
	sdk.AccAddress,
	map[uint64]sdk.AccAddress,
	error,
) {
	addr, ok := cache[accountNumber]
	if !ok {
		var err error
		addr, err = k.getAccountAddress(ctx, accountNumber)
		if err != nil {
			return nil, nil, err
		}
		cache[accountNumber] = addr
	}

	return addr, cache, nil
}
