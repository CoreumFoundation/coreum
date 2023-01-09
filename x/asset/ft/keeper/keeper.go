package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/samber/lo"
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
func (k Keeper) BeforeSendCoins(ctx sdk.Context, fromAddress, toAddress sdk.AccAddress, coins sdk.Coins) error {
	return k.applyFeatures(
		ctx,
		[]banktypes.Input{{Address: fromAddress.String(), Coins: coins}},
		[]banktypes.Output{{Address: toAddress.String(), Coins: coins}},
	)
}

// BeforeInputOutputCoins extends InputOutputCoins method of the bank keeper
func (k Keeper) BeforeInputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	return k.applyFeatures(ctx, inputs, outputs)
}

// MultiSendIterationInfo is used to gather information about multi send, and will be used to calculate
// burn_rate and send_commission_rate must be calculated in multi send
type MultiSendIterationInfo struct {
	FT                 types.FTDefinition
	NonIssuerInputSum  sdk.Int
	NonIssuerOutputSum sdk.Int
	NonIssuerSenders   map[string]sdk.Int
	Senders            map[string]sdk.Int
	Receivers          map[string]sdk.Int
}

func (k Keeper) applyFeatures2(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	// squashInputs / squashOutputs
}

func squashInputs(inputs []banktypes.Input) []banktypes.Input {
	accAddressCoins := make(map[string]sdk.Coins)

	for _, input := range inputs {
		accCoins, ok := accAddressCoins[input.Address]
		if ok {
			accCoins = accCoins.Add(input.Coins...)
		} else {
			accCoins = input.Coins
		}

		accAddressCoins[input.Address] = accCoins
	}

	return lo.MapToSlice(accAddressCoins, func(accAddress string, coins sdk.Coins) banktypes.Input {
		return banktypes.Input{
			Address: accAddress,
			Coins:   coins,
		}
	})
}

// CalculateRateShares returns the burn coins and commission coins
func (info MultiSendIterationInfo) CalculateRateShares(rate sdk.Dec) map[string]sdk.Int {
	minNonIssuerIOAmount := info.NonIssuerOutputSum
	if info.NonIssuerInputSum.LT(info.NonIssuerOutputSum) {
		minNonIssuerIOAmount = info.NonIssuerInputSum
	}

	shares := make(map[string]sdk.Int)
	amount := rate.MulInt(minNonIssuerIOAmount)
	if amount.IsPositive() {
		for sendAccount, sendAmount := range info.NonIssuerSenders {
			shares[sendAccount] = amount.Mul(sdk.NewDecFromInt(sendAmount)).Quo(sdk.NewDecFromInt(info.NonIssuerInputSum)).Ceil().RoundInt()
		}
	}

	return shares
}

func (k Keeper) fillInputs(address sdk.AccAddress, info *MultiSendIterationInfo, coin sdk.Coin) error {
	oldAmount, ok := info.Senders[address.String()]
	if !ok {
		oldAmount = sdk.NewInt(0)
	}

	newAmount := oldAmount.Add(coin.Amount)
	info.Senders[address.String()] = newAmount
	if info.FT.Issuer != address.String() {
		info.NonIssuerSenders[address.String()] = newAmount
		info.NonIssuerInputSum = info.NonIssuerInputSum.Add(coin.Amount)
	}

	return nil
}

func (k Keeper) fillOutputs(address sdk.AccAddress, info *MultiSendIterationInfo, coin sdk.Coin) error {
	oldAmount, ok := info.Receivers[address.String()]
	if !ok {
		oldAmount = sdk.NewInt(0)
	}
	newAmount := oldAmount.Add(coin.Amount)
	info.Receivers[address.String()] = newAmount
	if info.FT.Issuer != address.String() {
		info.NonIssuerOutputSum = info.NonIssuerOutputSum.Add(coin.Amount)
	}

	return nil
}

func (k Keeper) iterateInputOutputs(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) (map[string]*MultiSendIterationInfo, error) {
	iterationMap := make(map[string]*MultiSendIterationInfo)
	iterateCoin := func(coin sdk.Coin, address sdk.AccAddress, isInput bool) error {
		info, ok := iterationMap[coin.Denom]
		if !ok {
			ft, err := k.GetTokenDefinition(ctx, coin.Denom)
			if types.ErrFTNotFound.Is(err) {
				return nil
			}

			if err != nil {
				return err
			}
			info = &MultiSendIterationInfo{
				FT:                 ft,
				NonIssuerInputSum:  sdk.NewInt(0),
				NonIssuerOutputSum: sdk.NewInt(0),
				NonIssuerSenders:   make(map[string]sdk.Int),
				Senders:            make(map[string]sdk.Int),
				Receivers:          make(map[string]sdk.Int),
			}
			iterationMap[info.FT.Denom] = info
		}

		if isInput {
			if err := k.fillInputs(address, info, coin); err != nil {
				return err
			}
		} else {
			if err := k.fillOutputs(address, info, coin); err != nil {
				return err
			}
		}

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

func (k Keeper) applyFeatures(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	splitIntoMap, err := k.iterateInputOutputs(ctx, inputs, outputs)
	if err != nil {
		return err
	}

	for _, splitInfo := range splitIntoMap {
		burnShares := splitInfo.CalculateRateShares(splitInfo.FT.BurnRate)
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

		commissionShares := splitInfo.CalculateRateShares(splitInfo.FT.SendCommissionRate)
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

		// we need to check restraints after the burn_rate and send_commission_rate are applied
		for sender, amount := range splitInfo.Senders {
			senderAccAddress, err := sdk.AccAddressFromBech32(sender)
			if err != nil {
				return err
			}
			if err := k.isCoinSpendable(ctx, senderAccAddress, splitInfo.FT, amount); err != nil {
				return err
			}
		}

		for receiver, amount := range splitInfo.Receivers {
			receiverAccAddress, err := sdk.AccAddressFromBech32(receiver)
			if err != nil {
				return err
			}
			if err := k.isCoinReceivable(ctx, receiverAccAddress, splitInfo.FT, amount); err != nil {
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
