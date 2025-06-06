package keeper

import (
	"context"
	"time"

	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// InfiniteAccountKeeper replaces the original gas meter with the infinite one before calling an
// underlying method of real keeper. Gas consumed by the real keeper is non-deterministic. To use
// some ante decorators at the stage where deterministic gas must be delivered we use this wrapper
// to ignore gas consumed by keeper calls required there.
type InfiniteAccountKeeper struct {
	ak ante.AccountKeeper
}

// NewInfiniteAccountKeeper returns new InfiniteAccountKeeper.
func NewInfiniteAccountKeeper(ak ante.AccountKeeper) InfiniteAccountKeeper {
	return InfiniteAccountKeeper{
		ak: ak,
	}
}

// GetParams returns params.
//
//nolint:contextcheck // this is correct context passing
func (iak InfiniteAccountKeeper) GetParams(ctx context.Context) (params types.Params) {
	return iak.ak.GetParams(sdk.UnwrapSDKContext(ctx).WithGasMeter(storetypes.NewInfiniteGasMeter()))
}

// GetAccount returns account info by address.
//
//nolint:contextcheck // this is correct context passing
func (iak InfiniteAccountKeeper) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	ctx = sdk.UnwrapSDKContext(ctx).WithGasMeter(storetypes.NewInfiniteGasMeter())
	return iak.ak.GetAccount(ctx, addr)
}

// SetAccount sets account info.
//
//nolint:contextcheck // this is correct context passing
func (iak InfiniteAccountKeeper) SetAccount(ctx context.Context, acc sdk.AccountI) {
	ctx = sdk.UnwrapSDKContext(ctx).WithGasMeter(storetypes.NewInfiniteGasMeter())
	iak.ak.SetAccount(ctx, acc)
}

// GetModuleAddress returns address of a module.
func (iak InfiniteAccountKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	return iak.ak.GetModuleAddress(moduleName)
}

// AddressCodec returns the AddressCodec.
func (iak InfiniteAccountKeeper) AddressCodec() address.Codec {
	return iak.ak.AddressCodec()
}

// UnorderedTransactionsEnabled indicates whether unordered transactions are allowed in the InfiniteAccountKeeper.
func (iak InfiniteAccountKeeper) UnorderedTransactionsEnabled() bool {
	// TODO implement me
	return false
}

// RemoveExpiredUnorderedNonces removes nonces that have expired based on the current context.
func (iak InfiniteAccountKeeper) RemoveExpiredUnorderedNonces(ctx sdk.Context) error {
	// TODO implement me
	panic("implement me")
}

// TryAddUnorderedNonce attempts to add a new unordered nonce for the sender, ensuring validity based on the timestamp.
func (iak InfiniteAccountKeeper) TryAddUnorderedNonce(ctx sdk.Context, sender []byte, timestamp time.Time) error {
	// TODO implement me
	panic("implement me")
}
