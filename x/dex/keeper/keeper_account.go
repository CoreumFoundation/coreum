package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

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
	addr, err := k.accountQueryServer.AccountAddressByID(
		ctx,
		&authtypes.QueryAccountAddressByIDRequest{AccountId: accountNumber},
	)
	if err != nil {
		return nil, err
	}

	acc, err := sdk.AccAddressFromBech32(addr.AccountAddress)
	if err != nil {
		return nil, sdkerrors.Wrapf(types.ErrInvalidInput, "invalid address: %s", addr)
	}

	return acc, nil
}
