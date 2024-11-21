package keeper

import (
	"context"

	"cosmossdk.io/core/store"
	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/v5/x/wasm"
	cwasmtypes "github.com/CoreumFoundation/coreum/v5/x/wasm/types"
	"github.com/CoreumFoundation/coreum/v5/x/wbank/types"
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
	storeService store.KVStoreService,
	ak banktypes.AccountKeeper,
	wasmKeeper cwasmtypes.WasmKeeper,
	blockedAddrs map[string]bool,
	ftProvider types.FungibleTokenProvider,
	authority string,
	logger log.Logger,
) BaseKeeperWrapper {
	return BaseKeeperWrapper{
		BaseKeeper: bankkeeper.NewBaseKeeper(cdc, storeService, ak, blockedAddrs, authority, logger),
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
	ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins,
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
	ctx context.Context, senderModule, recipientModule string, amt sdk.Coins,
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
	ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
) error {
	recipientAcc := k.ak.GetModuleAccount(ctx, recipientModule)
	if recipientAcc == nil {
		panic(sdkerrors.Wrapf(cosmoserrors.ErrUnknownAddress, "module account %s does not exist", recipientModule))
	}

	return k.SendCoins(ctx, senderAddr, recipientAcc.GetAddress(), amt)
}

// SendCoins is a BaseKeeper SendCoins wrapped method.
//
//nolint:contextcheck // this is correct context passing.
func (k BaseKeeperWrapper) SendCoins(goCtx context.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if k.isSmartContract(ctx, fromAddr) {
		ctx = cwasmtypes.WithSmartContractSender(ctx, fromAddr.String())
	}
	if k.isSmartContract(ctx, toAddr) {
		ctx = cwasmtypes.WithSmartContractRecipient(ctx, toAddr.String())
	}

	return k.ftProvider.BeforeSendCoins(ctx, fromAddr, toAddr, amt)
}

// DelegateCoinsFromAccountToModule is a BaseKeeper DelegateCoinsFromAccountToModule wrapped method.
//
//nolint:contextcheck // this is correct context passing.
func (k BaseKeeperWrapper) DelegateCoinsFromAccountToModule(
	goCtx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins,
) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.beforeDelegateCoins(ctx, senderAddr, amt); err != nil {
		return err
	}
	return k.BaseKeeper.DelegateCoinsFromAccountToModule(ctx, senderAddr, recipientModule, amt)
}

func (k BaseKeeperWrapper) beforeDelegateCoins(ctx sdk.Context, senderAddr sdk.AccAddress, amt sdk.Coins) error {
	for _, coin := range amt {
		balance := k.BaseKeeper.GetBalance(ctx, senderAddr, coin.Denom)

		spendableBalance, err := balance.SafeSub(k.ftProvider.GetDEXLockedBalance(ctx, senderAddr, coin.Denom))
		if err != nil {
			return err
		}

		if spendableBalance.IsLT(coin) {
			return sdkerrors.Wrapf(cosmoserrors.ErrInsufficientFunds,
				"account %s does not have enough %s tokens to delegate", senderAddr.String(), coin.Denom,
			)
		}
	}
	return nil
}

// InputOutputCoins is a BaseKeeper InputOutputCoins wrapped method.
//
//nolint:contextcheck // this is correct context passing.
func (k BaseKeeperWrapper) InputOutputCoins(
	goCtx context.Context, input banktypes.Input, outputs []banktypes.Output,
) error {
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(input.Address)
	if err != nil {
		return err
	}
	if k.isSmartContract(ctx, addr) {
		ctx = cwasmtypes.WithSmartContractSender(ctx, input.Address)
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

	return k.ftProvider.BeforeInputOutputCoins(ctx, input, outputs)
}

// ********** Query server **********

// SpendableBalances implements a gRPC query handler for retrieving an account's spendable balances including asset ft
// frozen coins.
func (k BaseKeeperWrapper) SpendableBalances(
	ctx context.Context, req *banktypes.QuerySpendableBalancesRequest,
) (*banktypes.QuerySpendableBalancesResponse, error) {
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid address %s", req.Address)
	}

	balancesRes, err := k.BaseKeeper.AllBalances(ctx, &banktypes.QueryAllBalancesRequest{
		Address:    req.Address,
		Pagination: req.Pagination,
	})
	if err != nil {
		return nil, err
	}

	balances := balancesRes.Balances
	for i := range balances {
		spendableCoin, err := k.ftProvider.GetSpendableBalance(sdk.UnwrapSDKContext(ctx), addr, balances[i].Denom)
		if err != nil {
			return nil, err
		}
		balances[i] = spendableCoin
	}

	return &banktypes.QuerySpendableBalancesResponse{
		Balances:   balances,
		Pagination: balancesRes.Pagination,
	}, nil
}

// SpendableBalanceByDenom implements a gRPC query handler for retrieving an account's spendable balance for a specific
// denom, including asset ft frozen coins.
func (k BaseKeeperWrapper) SpendableBalanceByDenom(
	ctx context.Context, req *banktypes.QuerySpendableBalanceByDenomRequest,
) (*banktypes.QuerySpendableBalanceByDenomResponse, error) {
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrapf(cosmoserrors.ErrInvalidAddress, "invalid address %s", req.Address)
	}

	balanceRes, err := k.BaseKeeper.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: req.Address,
		Denom:   req.Denom,
	})
	if err != nil {
		return nil, err
	}

	if balanceRes.Balance == nil {
		return &banktypes.QuerySpendableBalanceByDenomResponse{}, nil
	}

	spendableCoin, err := k.ftProvider.GetSpendableBalance(sdk.UnwrapSDKContext(ctx), addr, balanceRes.Balance.Denom)
	if err != nil {
		return nil, err
	}

	return &banktypes.QuerySpendableBalanceByDenomResponse{
		Balance: lo.ToPtr(spendableCoin),
	}, nil
}

func (k BaseKeeperWrapper) isSmartContract(ctx sdk.Context, addr sdk.AccAddress) bool {
	return wasm.IsSmartContract(ctx, addr, k.wasmKeeper)
}
