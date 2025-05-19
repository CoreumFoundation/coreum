package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

func (k Keeper) getAccountNumber(ctx sdk.Context, addr sdk.AccAddress) (uint64, error) {
	acc := k.accountKeeper.GetAccount(ctx, addr)
	if acc == nil {
		return 0, sdkerrors.Wrapf(types.ErrInvalidInput, "account not found: %s", addr.String())
	}

	return acc.GetAccountNumber(), nil
}
