package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
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
func (k Keeper) BeforeSendCoins(ctx sdk.Context, fromAddress, toAddress sdk.AccAddress, coins sdk.Coins) error {
	return k.BeforeInputOutputCoins(
		ctx,
		[]banktypes.Input{{Address: fromAddress.String(), Coins: coins}},
		[]banktypes.Output{{Address: toAddress.String(), Coins: coins}},
	)
}

// MultiSendIterationInfo is used to gather information about multi send, and will be used to calculate
// burn_rate and send_commission_rate must be calculate in multi send
type MultiSendIterationInfo struct {
	FT                 types.FTDefinition
	NonIssuerInputSum  sdk.Int
	NonIssuerOutputSum sdk.Int
	NonIssuerSenders   map[string]sdk.Int
	Senders            map[string]sdk.Int
	Receivers          map[string]sdk.Int
}

// CalculateBurnRateShares returns the coins to be burned
func (info MultiSendIterationInfo) CalculateBurnRateShares() (map[string]sdk.Int, map[string]sdk.Int) {
	var minNonIssuerIOAmount sdk.Int
	if info.NonIssuerInputSum.LT(info.NonIssuerOutputSum) {
		minNonIssuerIOAmount = info.NonIssuerInputSum
	} else {
		minNonIssuerIOAmount = info.NonIssuerOutputSum
	}

	burnShares := map[string]sdk.Int{}
	burnAmount := info.FT.BurnRate.MulInt(minNonIssuerIOAmount)
	if burnAmount.IsPositive() {
		for sendAccount, sendAmount := range info.NonIssuerSenders {
			burnShare := burnAmount.Mul(sdk.NewDecFromInt(sendAmount)).Quo(sdk.NewDecFromInt(info.NonIssuerInputSum)).Ceil().RoundInt()
			burnShares[sendAccount] = burnShare
		}
	}

	commissionShares := map[string]sdk.Int{}
	commissionAmount := info.FT.SendCommissionRate.MulInt(minNonIssuerIOAmount)
	if commissionAmount.IsPositive() {
		for sendAccount, sendAmount := range info.NonIssuerSenders {
			commissionShare := commissionAmount.Mul(sdk.NewDecFromInt(sendAmount)).Quo(sdk.NewDecFromInt(info.NonIssuerInputSum)).Ceil().RoundInt()
			commissionShares[sendAccount] = commissionShare
		}
	}

	return burnShares, commissionShares
}

func (k Keeper) iterateInputOutputs(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) (map[string]MultiSendIterationInfo, error) {
	iterationMap := map[string]MultiSendIterationInfo{}
	iterateCoin := func(coin sdk.Coin, address sdk.AccAddress, isInput bool) error {
		ft, err := k.GetTokenDefinition(ctx, coin.Denom)
		if types.ErrFTNotFound.Is(err) {
			return nil
		}

		if err != nil {
			return err
		}

		info, ok := iterationMap[ft.Denom]
		if !ok {
			info = MultiSendIterationInfo{
				FT:                 ft,
				NonIssuerInputSum:  sdk.NewInt(0),
				NonIssuerOutputSum: sdk.NewInt(0),
				NonIssuerSenders:   map[string]sdk.Int{},
				Senders:            map[string]sdk.Int{},
				Receivers:          map[string]sdk.Int{},
			}
		}

		//nolint:nestif // I could not find a way to fix the complexity here without sacrificing performance
		if isInput {
			oldAmount, ok := info.Senders[address.String()]
			if !ok {
				oldAmount = sdk.NewInt(0)
			}

			newAmount := oldAmount.Add(coin.Amount)
			if err := k.isCoinSpendable(ctx, address, ft, newAmount); err != nil {
				return err
			}

			info.Senders[address.String()] = newAmount
			if ft.Issuer != address.String() {
				info.NonIssuerSenders[address.String()] = newAmount
				info.NonIssuerInputSum = info.NonIssuerInputSum.Add(coin.Amount)
			}
		} else {
			oldAmount, ok := info.Receivers[address.String()]
			if !ok {
				oldAmount = sdk.NewInt(0)
			}
			newAmount := oldAmount.Add(coin.Amount)
			info.Receivers[address.String()] = newAmount
			if err := k.isCoinReceivable(ctx, address, ft, newAmount); err != nil {
				return err
			}

			if ft.Issuer != address.String() {
				info.NonIssuerOutputSum = info.NonIssuerOutputSum.Add(coin.Amount)
			}
		}

		iterationMap[ft.Denom] = info
		return nil
	}

	for _, in := range inputs {
		inAddress, err := sdk.AccAddressFromBech32(in.Address)
		if err != nil {
			return nil, err
		}

		for _, coin := range in.Coins {
			err := iterateCoin(coin, inAddress, true)
			if err != nil {
				return nil, err
			}
		}
	}

	for _, out := range outputs {
		outAddress, err := sdk.AccAddressFromBech32(out.Address)
		if err != nil {
			return nil, err
		}

		for _, coin := range out.Coins {
			err := iterateCoin(coin, outAddress, false)
			if err != nil {
				return nil, err
			}
		}
	}

	return iterationMap, nil
}

// BeforeInputOutputCoins extends InputOutputCoins method of the bank keeper
func (k Keeper) BeforeInputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	splitIntoMap, err := k.iterateInputOutputs(ctx, inputs, outputs)
	if err != nil {
		return err
	}

	for _, splitInfo := range splitIntoMap {
		burnShares, commissionShares := splitInfo.CalculateBurnRateShares()
		for account, burnShare := range burnShares {
			senderAccAddress, err := sdk.AccAddressFromBech32(account)
			if err != nil {
				return err
			}

			err = k.burn(ctx, senderAccAddress, splitInfo.FT, burnShare)
			if err != nil {
				return err
			}
		}

		for account, commissionShare := range commissionShares {
			senderAccAddress, err := sdk.AccAddressFromBech32(account)
			if err != nil {
				return err
			}

			coins := sdk.NewCoins(sdk.NewCoin(splitInfo.FT.Denom, commissionShare))
			err = k.bankKeeper.SendCoins(ctx, senderAccAddress, sdk.MustAccAddressFromBech32(splitInfo.FT.Issuer), coins)
			if err != nil {
				return err
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
			if types.ErrFTNotFound.Is(err) {
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
