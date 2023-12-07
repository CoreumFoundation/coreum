package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum/v4/x/wasm"
	cwasmtypes "github.com/CoreumFoundation/coreum/v4/x/wasm/types"
	"github.com/CoreumFoundation/coreum/v4/x/wbank/types"
)

// BaseKeeperWrapper is a wrapper of the cosmos-sdk bank module.
type BaseKeeperWrapper struct {
	bankkeeper.BaseKeeper
	ak         banktypes.AccountKeeper
	wasmKeeper cwasmtypes.WasmKeeper
	ftProvider types.FungibleTokenProvider
}

// NewKeeper returns a new BaseKeeperWrapper instance.
func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	ak banktypes.AccountKeeper,
	wasmKeeper cwasmtypes.WasmKeeper,
	blockedAddrs map[string]bool,
	ftProvider types.FungibleTokenProvider,
	authority string,
) BaseKeeperWrapper {
	return BaseKeeperWrapper{
		BaseKeeper: bankkeeper.NewBaseKeeper(cdc, storeKey, ak, blockedAddrs, authority),
		ak:         ak,
		wasmKeeper: wasmKeeper,
		ftProvider: ftProvider,
	}
}

// SendCoinsFromModuleToAccount transfers coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist. An error is returned if
// the recipient address is black-listed or if sending the tokens fails.
// !!! The code is the copy of the corresponding func of the bank module !!!
func (k BaseKeeperWrapper) SendCoinsFromModuleToAccount(
	ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
) error {
	senderAddr := k.ak.GetModuleAddress(senderModule)
	if senderAddr == nil {
		panic(sdkerrors.Wrapf(cosmoserrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	if k.BlockedAddr(recipientAddr) {
		return sdkerrors.Wrapf(cosmoserrors.ErrUnauthorized, "%s is not allowed to receive funds", recipientAddr)
	}

	return k.SendCoins(ctx, senderAddr, recipientAddr, amt)
}

// SendCoinsFromModuleToModule transfers coins from a ModuleAccount to another.
// It will panic if either module account does not exist.
// !!! The code is the copy of the corresponding func of the bank module !!!
func (k BaseKeeperWrapper) SendCoinsFromModuleToModule(
	ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins,
) error {
	senderAddr := k.ak.GetModuleAddress(senderModule)
	if senderAddr == nil {
		panic(sdkerrors.Wrapf(cosmoserrors.ErrUnknownAddress, "module account %s does not exist", senderModule))
	}

	recipientAcc := k.ak.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(sdkerrors.Wrapf(cosmoserrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
	}

	return k.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// SendCoinsFromAccountToModule transfers coins from an AccAddress to a ModuleAccount.
// It will panic if the module account does not exist.
// !!! The code is the copy of the corresponding func of the bank module !!!
func (k BaseKeeperWrapper) SendCoinsFromAccountToModule(
	ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
) error {
	recipientAcc := k.ak.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(sdkerrors.Wrapf(cosmoserrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
	}

	return k.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// SendCoins is a BaseKeeper SendCoins wrapped method.
func (k BaseKeeperWrapper) SendCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	if k.isSmartContract(ctx, fromAddr) {
		ctx = cwasmtypes.WithSmartContractSender(ctx, fromAddr.String())
	}
	if k.isSmartContract(ctx, toAddr) {
		ctx = cwasmtypes.WithSmartContractRecipient(ctx, toAddr.String())
	}

	if err := k.ftProvider.BeforeSendCoins(ctx, fromAddr, toAddr, amt); err != nil {
		return err
	}

	return k.BaseKeeper.SendCoins(ctx, fromAddr, toAddr, amt)
}

// InputOutputCoins is a BaseKeeper InputOutputCoins wrapped method.
func (k BaseKeeperWrapper) InputOutputCoins(
	ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output,
) error {
	for _, input := range inputs {
		addr, err := sdk.AccAddressFromBech32(input.Address)
		if err != nil {
			return err
		}
		if k.isSmartContract(ctx, addr) {
			ctx = cwasmtypes.WithSmartContractSender(ctx, input.Address)
		}
	}
	for _, output := range outputs {
		addr, err := sdk.AccAddressFromBech32(output.Address)
		if err != nil {
			return err
		}
		if k.isSmartContract(ctx, addr) {
			ctx = cwasmtypes.WithSmartContractRecipient(ctx, output.Address)
		}
	}

	if err := k.ftProvider.BeforeInputOutputCoins(ctx, inputs, outputs); err != nil {
		return err
	}

	return k.BaseKeeper.InputOutputCoins(ctx, inputs, outputs)
}

// ********** Query server **********

// SpendableBalances implements a gRPC query handler for retrieving an account's spendable balances including asset ft
// frozen coins.
func (k BaseKeeperWrapper) SpendableBalances(
	ctx context.Context, req *banktypes.QuerySpendableBalancesRequest,
) (*banktypes.QuerySpendableBalancesResponse, error) {
	res, err := k.BaseKeeper.SpendableBalances(ctx, req)
	if err != nil {
		return nil, err
	}
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid address %s", req.Address)
	}
	for i := range res.Balances {
		res.Balances[i] = k.getSpendableCoin(sdk.UnwrapSDKContext(ctx), addr, res.Balances[i])
	}

	return res, nil
}

// SpendableBalanceByDenom implements a gRPC query handler for retrieving an account's spendable balance for a specific
// denom, including asset ft frozen coins.
func (k BaseKeeperWrapper) SpendableBalanceByDenom(
	ctx context.Context, req *banktypes.QuerySpendableBalanceByDenomRequest,
) (*banktypes.QuerySpendableBalanceByDenomResponse, error) {
	res, err := k.BaseKeeper.SpendableBalanceByDenom(ctx, req)
	if err != nil {
		return nil, err
	}
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid address %s", req.Address)
	}
	if res.Balance == nil {
		return res, nil
	}

	spendableCoin := k.getSpendableCoin(sdk.UnwrapSDKContext(ctx), addr, *res.Balance)
	res.Balance = &spendableCoin

	return res, nil
}

func (k BaseKeeperWrapper) getSpendableCoin(ctx sdk.Context, addr sdk.AccAddress, coin sdk.Coin) sdk.Coin {
	denom := coin.Denom
	frozenCoin := k.ftProvider.GetFrozenBalance(ctx, addr, denom)
	spendableAmount := coin.Amount.Sub(frozenCoin.Amount)
	if spendableAmount.IsNegative() {
		return sdk.NewCoin(denom, sdkmath.ZeroInt())
	}

	return sdk.NewCoin(denom, spendableAmount)
}

func (k BaseKeeperWrapper) isSmartContract(ctx sdk.Context, addr sdk.AccAddress) bool {
	return wasm.IsSmartContract(ctx, addr, k.wasmKeeper)
}
