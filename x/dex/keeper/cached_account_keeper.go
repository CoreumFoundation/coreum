package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

type cachedAccountKeeper struct {
	accountKeeper      types.AccountKeeper
	accountQueryServer types.AccountQueryServer

	cache map[uint64]sdk.AccAddress
}

func newCachedAccountKeeper(accountKeeper types.AccountKeeper, accountQueryServer types.AccountQueryServer) cachedAccountKeeper {
	return cachedAccountKeeper{
		accountKeeper:      accountKeeper,
		accountQueryServer: accountQueryServer,
		cache:              make(map[uint64]sdk.AccAddress),
	}
}

func (cak cachedAccountKeeper) getAccountAddress(ctx sdk.Context, accNumber uint64) (sdk.AccAddress, error) {
	addr, err := cak.accountQueryServer.AccountAddressByID(
		ctx,
		&authtypes.QueryAccountAddressByIDRequest{AccountId: accNumber},
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

func (cak cachedAccountKeeper) getAccountAddressWithCache(ctx sdk.Context, accNumber uint64) (
	sdk.AccAddress,
	error,
) {
	addr, ok := cak.cache[accNumber]
	if !ok {
		var err error
		addr, err = cak.getAccountAddress(ctx, accNumber)
		if err != nil {
			return nil, err
		}
		cak.cache[accNumber] = addr
	}

	return addr, nil
}
