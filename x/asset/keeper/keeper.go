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

// InterceptSendCoins checks that a transfer request is allowed or not
func (k Keeper) InterceptSendCoins(ctx sdk.Context, fromAddress, toAddress sdk.AccAddress, coins sdk.Coins) error {
	if err := k.areCoinsSpendable(ctx, fromAddress, coins); err != nil {
		return err
	}

	for _, coin := range coins {
		ft, err := k.GetFungibleTokenDefinition(ctx, coin.Denom)
		if types.ErrFungibleTokenNotFound.Is(err) {
			continue
		}

		if err != nil {
			return err
		}

		if ft.BurnRate > 0 && ft.Issuer != fromAddress.String() && ft.Issuer != toAddress.String() {
			coinsToBurn := ft.CalculateBurnCoin(coin)
			err := k.burnFungibleToken(ctx, coinsToBurn, fromAddress)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// InterceptInputOutputCoins extends InputOutputCoins method of the bank keeper
func (k Keeper) InterceptInputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
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

			if ft.BurnRate > 0 && ft.Issuer != inAddress.String() {
				coinsToBurn := ft.CalculateBurnCoin(coin)
				err = k.burnFungibleToken(ctx, coinsToBurn, inAddress)
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

	return k.mintFungibleToken(ctx, coin.Denom, coin.Amount, sender)
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

	return k.burnFungibleToken(ctx, coin, sender)
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
