package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// ParamSubspace represents a subscope of methods exposed by param module to store and retrieve parameters
type ParamSubspace interface {
	GetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
	SetParamSet(ctx sdk.Context, ps paramtypes.ParamSet)
}

// Keeper is the asset module keeper.
type Keeper struct {
	cdc           codec.BinaryCodec
	paramSubspace ParamSubspace
	storeKey      sdk.StoreKey
	bankKeeper    types.BankKeeper
}

// NewKeeper creates a new instance of the Keeper.
func NewKeeper(
	cdc codec.BinaryCodec,
	paramSubspace ParamSubspace,
	storeKey sdk.StoreKey,
	bankKeeper types.BankKeeper,
) Keeper {
	return Keeper{
		cdc:           cdc,
		paramSubspace: paramSubspace,
		storeKey:      storeKey,
		bankKeeper:    bankKeeper,
	}
}

// SetParams sets the parameters of the model
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramSubspace.SetParamSet(ctx, &params)
}

// GetParams gets the parameters of the model
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	var params types.Params
	k.paramSubspace.GetParamSet(ctx, &params)
	return params
}

// BeforeSendCoins checks that a transfer request is allowed or not
//
// TODO: we should try to express this function in terms of BeforeInputOutputCoins so
// we will have a single place to enforce our logic
func (k Keeper) BeforeSendCoins(ctx sdk.Context, fromAddress, toAddress sdk.AccAddress, coins sdk.Coins) error {
	for _, coin := range coins {
		ft, err := k.GetTokenDefinition(ctx, coin.Denom)
		if err != nil {
			if types.ErrInvalidDenom.Is(err) || types.ErrFTNotFound.Is(err) {
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
		if err := k.applySendCommissionRate(ctx, ft, fromAddress, toAddress, coin); err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) applyBurnRate(ctx sdk.Context, ft types.FTDefinition, fromAddress, toAddress sdk.AccAddress, coin sdk.Coin) error {
	if !ft.BurnRate.IsNil() && ft.BurnRate.IsPositive() && ft.Issuer != fromAddress.String() && ft.Issuer != toAddress.String() {
		coinToBurn := ft.CalculateBurnRateAmount(coin)
		err := k.burn(ctx, fromAddress, ft, coinToBurn)
		if err != nil {
			return err
		}
	}

	return nil
}

func (k Keeper) applySendCommissionRate(ctx sdk.Context, ft types.FTDefinition, fromAddress, toAddress sdk.AccAddress, coin sdk.Coin) error {
	if !ft.SendCommissionRate.IsNil() && ft.SendCommissionRate.IsPositive() && ft.Issuer != fromAddress.String() && ft.Issuer != toAddress.String() {
		sendCommission := ft.CalculateSendCommissionRateAmount(coin)
		err := k.bankKeeper.SendCoins(ctx, fromAddress, sdk.MustAccAddressFromBech32(ft.Issuer), sdk.NewCoins(sdk.NewCoin(coin.Denom, sendCommission)))
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
			ft, err := k.GetTokenDefinition(ctx, coin.Denom)
			if types.ErrInvalidDenom.Is(err) || types.ErrFTNotFound.Is(err) {
				continue
			}

			if err != nil {
				return err
			}

			if err := k.isCoinSpendable(ctx, inAddress, ft, coin.Amount); err != nil {
				return err
			}

			if !ft.BurnRate.IsNil() && ft.BurnRate.IsPositive() && ft.Issuer != inAddress.String() {
				coinsToBurn := ft.CalculateBurnRateAmount(coin)
				err := k.burn(ctx, inAddress, ft, coinsToBurn)
				if err != nil {
					return err
				}
			}
			if !ft.SendCommissionRate.IsNil() && ft.SendCommissionRate.IsPositive() && ft.Issuer != inAddress.String() {
				sendCommissionRate := ft.CalculateSendCommissionRateAmount(coin)
				err := k.bankKeeper.SendCoins(ctx, inAddress, sdk.MustAccAddressFromBech32(ft.Issuer), sdk.NewCoins(sdk.NewCoin(coin.Denom, sendCommissionRate)))
				if err != nil {
					return err
				}
			}
		}
	}

	for _, out := range outputs {
		outAddress, err := sdk.AccAddressFromBech32(out.Address)
		if err != nil {
			return err
		}

		for _, coin := range out.Coins {
			ft, err := k.GetTokenDefinition(ctx, coin.Denom)
			if types.ErrInvalidDenom.Is(err) || types.ErrFTNotFound.Is(err) {
				continue
			}
			if err != nil {
				return err
			}

			if err := k.isCoinReceivable(ctx, outAddress, ft, coin.Amount); err != nil {
				return err
			}
		}
	}

	return nil
}

// Logger returns the Keeper logger.
func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

// Mint mints new fungible token
func (k Keeper) Mint(ctx sdk.Context, sender sdk.AccAddress, coin sdk.Coin) error {
	ft, err := k.GetTokenDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	err = k.checkFeatureAllowed(sender, ft, types.TokenFeature_mint) //nolint:nosnakecase
	if err != nil {
		return err
	}

	return k.mint(ctx, ft, coin.Amount, sender)
}

// Burn burns fungible token
func (k Keeper) Burn(ctx sdk.Context, sender sdk.AccAddress, coin sdk.Coin) error {
	ft, err := k.GetTokenDefinition(ctx, coin.Denom)
	if err != nil {
		return sdkerrors.Wrapf(err, "not able to get token info for denom:%s", coin.Denom)
	}

	err = k.checkFeatureAllowed(sender, ft, types.TokenFeature_burn) //nolint:nosnakecase
	if err != nil {
		return err
	}

	return k.burn(ctx, sender, ft, coin.Amount)
}

func (k Keeper) checkFeatureAllowed(sender sdk.AccAddress, ft types.FTDefinition, feature types.TokenFeature) error {
	if !ft.IsFeatureEnabled(feature) {
		return sdkerrors.Wrapf(types.ErrFeatureNotActive, "denom:%s, feature:%s", ft.Denom, feature)
	}

	if ft.Issuer != sender.String() {
		return sdkerrors.Wrapf(sdkerrors.ErrUnauthorized, "address %s is unauthorized to perform this operation", sender.String())
	}

	return nil
}
