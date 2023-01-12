package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

func (k Keeper) applyFeatures(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	inputGroup := make(groupedByDenomAccountBalances)
	for _, in := range inputs {
		inputGroup.add(in.Address, in.Coins)
	}

	outputGroup := make(groupedByDenomAccountBalances)
	for _, out := range outputs {
		outputGroup.add(out.Address, out.Coins)
	}

	return k.applyRates(ctx, inputGroup, outputGroup)
}

type accountBalanceMap map[string]sdk.Int

type groupedByDenomAccountBalances map[string]accountBalanceMap

func (g groupedByDenomAccountBalances) add(address string, coins sdk.Coins) {
	for _, coin := range coins {
		entry, ok := g[coin.Denom]
		if !ok {
			entry = make(map[string]sdk.Int)
		}
		oldAmount, ok := entry[address]
		if !ok {
			oldAmount = sdk.ZeroInt()
		}

		oldAmount = oldAmount.Add(coin.Amount)
		entry[address] = oldAmount
		g[coin.Denom] = entry
	}
}

func (k Keeper) applyRates(ctx sdk.Context, inputs, outputs groupedByDenomAccountBalances) error {
	for denom, inOps := range inputs {
		ftd, err := k.GetTokenDefinition(ctx, denom)
		if types.ErrTokenNotFound.Is(err) {
			return nil
		}

		outOps := outputs[denom]

		burnShares := CalculateRateShares(ftd.BurnRate, ftd.Issuer, inOps, outOps)
		for account, amount := range burnShares {
			if err := k.burn(ctx, sdk.MustAccAddressFromBech32(account), ftd, amount); err != nil {
				return err
			}
		}

		commissionShares := CalculateRateShares(ftd.SendCommissionRate, ftd.Issuer, inOps, outOps)
		for account, amount := range commissionShares {
			coins := sdk.NewCoins(sdk.NewCoin(ftd.Denom, amount))
			issuer := sdk.MustAccAddressFromBech32(ftd.Issuer)
			if err := k.bankKeeper.SendCoins(ctx, sdk.MustAccAddressFromBech32(account), issuer, coins); err != nil {
				return err
			}
		}

		for account, amount := range inOps {
			if err := k.isCoinSpendable(ctx, sdk.MustAccAddressFromBech32(account), ftd, amount); err != nil {
				return err
			}
		}

		for account, amount := range outOps {
			if err := k.isCoinReceivable(ctx, sdk.MustAccAddressFromBech32(account), ftd, amount); err != nil {
				return err
			}
		}
	}

	return nil
}

// CalculateRateShares calculates how the burn or commission share amount should be split between different parties
func CalculateRateShares(rate sdk.Dec, issuer string, inOps, outOps accountBalanceMap) map[string]sdk.Int {
	// The algorithm is as following. we first get the minimum of total inputs and outputs which are not
	// from the issuer. We then multiply by the rate to get applicable total amount. we then split this amount
	// between non-issuer senders, proportional to their input value.
	if rate.IsNil() || !rate.IsPositive() {
		return nil
	}

	nonIssuerSum := func(values accountBalanceMap) sdk.Int {
		sum := sdk.ZeroInt()
		for account, amount := range values {
			if account != issuer {
				sum = sum.Add(amount)
			}
		}
		return sum
	}

	inputSumNonIssuer := nonIssuerSum(inOps)
	outputSumNonIssuer := nonIssuerSum(outOps)

	minNonIssuer := inputSumNonIssuer
	if outputSumNonIssuer.LT(minNonIssuer) {
		minNonIssuer = outputSumNonIssuer
	}

	shares := make(accountBalanceMap, 0)
	if minNonIssuer.IsPositive() {
		for account, amount := range inOps {
			if account != issuer {
				finalShare := rate.MulInt(minNonIssuer.Mul(amount)).QuoInt(inputSumNonIssuer).Ceil().RoundInt()
				shares[account] = finalShare
			}
		}
	}
	return shares
}
