package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// Keeper is the asset module keeper.
type Keeper struct {
	cdc        codec.BinaryCodec
	storeKey   sdk.StoreKey
	bankKeeper types.BankKeeper
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(cdc codec.BinaryCodec, storeKey sdk.StoreKey, bankKeeper types.BankKeeper) Keeper {
	return Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		bankKeeper: bankKeeper,
	}
}

// BeforeSendCoins checks that a transfer request is allowed or not
//
// TODO: we should try to express this function in terms of BeforeInputOutputCoins so
// we will have a single place to enforce our logic
func (k Keeper) BeforeSendCoins(ctx sdk.Context, fromAddress, toAddress sdk.AccAddress, coins sdk.Coins) error {
	for _, coin := range coins {
		ft, err := k.GetFungibleTokenDefinition(ctx, coin.Denom)
		if err != nil {
			if types.ErrFungibleTokenNotFound.Is(err) {
				continue
			}
			return err
		}
		if err := k.isCoinSpendable(ctx, fromAddress, ft, coin.Amount); err != nil {
			return err
		}
		if err := k.isCoinReceivable(ctx, toAddress, ft, coin.Amount); err != nil {
			return err
		}
		if err := k.applyBurnRate(ctx, ft, fromAddress, toAddress, coin); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) applyBurnRate(ctx sdk.Context, ft types.FungibleTokenDefinition, fromAddress, toAddress sdk.AccAddress, coin sdk.Coin) error {
	if !ft.BurnRate.IsNil() && ft.BurnRate.IsPositive() && ft.Issuer != fromAddress.String() && ft.Issuer != toAddress.String() {
		coinToBurn := ft.CalculateBurnRateAmount(coin)
		err := k.burnFungibleToken(ctx, fromAddress, ft, coinToBurn)
		if err != nil {
			return err
		}
	}

	return nil
}

// BeforeInputOutputCoins extends InputOutputCoins method of the bank keeper
func (k Keeper) BeforeInputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	for _, in := range inputs {
		inAddress, err := sdk.AccAddressFromBech32(in.Address)
		if err != nil {
			return err
		}

		for _, coin := range in.Coins {
			ft, err := k.GetFungibleTokenDefinition(ctx, coin.Denom)
			if types.ErrFungibleTokenNotFound.Is(err) {
				continue
			}

			if err != nil {
				return err
			}

			if !ft.BurnRate.IsNil() && ft.BurnRate.IsPositive() && ft.Issuer != inAddress.String() {
				coinsToBurn := ft.CalculateBurnRateAmount(coin)
				err = k.burnFungibleToken(ctx, inAddress, ft, coinsToBurn)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Logger returns the Keeper logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// MintFungibleToken mints new fungible token
func (k Keeper) MintFungibleToken(ctx sdk.Context, sender sdk.AccAddress, coin sdk.Coin) error {
	ft, err := k.GetFungibleTokenDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	err = k.checkFeatureAllowed(sender, ft, types.FungibleTokenFeature_mint) //nolint:nosnakecase
	if err != nil {
		return err
	}

	return k.mintFungibleToken(ctx, ft, coin.Amount, sender)
}

// BurnFungibleToken burns fungible token
func (k Keeper) BurnFungibleToken(ctx sdk.Context, sender sdk.AccAddress, coin sdk.Coin) error {
	ft, err := k.GetFungibleTokenDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	err = k.checkFeatureAllowed(sender, ft, types.FungibleTokenFeature_burn) //nolint:nosnakecase
	if err != nil {
		return err
	}

	return k.burnFungibleToken(ctx, sender, ft, coin.Amount)
}

func (k Keeper) checkFeatureAllowed(sender sdk.AccAddress, ft types.FungibleTokenDefinition, feature types.FungibleTokenFeature) error {
	if !ft.IsFeatureEnabled(feature) {
		return sdkerrors.Wrapf(types.ErrFeatureNotActive, "denom:%s, feature:%s", ft.Denom, feature)
	}

	if ft.Issuer != sender.String() {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "address %s is unauthorized to perform this operation", sender.String())
	}

	return nil
}
