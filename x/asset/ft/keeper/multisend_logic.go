package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// AccAmount is a paring between Account and Amount
type AccAmount struct {
	Account sdk.AccAddress
	Amount  sdk.Int
}

func (k Keeper) applyFeatures(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	denomToAccountsInput := make(groupByDenomContainer)
	for _, in := range inputs {
		denomToAccountsInput.add(in.Address, in.Coins)
	}

	denomToAccountsOutput := make(groupByDenomContainer)
	for _, out := range outputs {
		denomToAccountsOutput.add(out.Address, out.Coins)
	}

	return k.applyRates(ctx, denomToAccountsInput.flatten(), denomToAccountsOutput.flatten())
}

type groupByDenomContainer map[string]map[string]sdk.Int

func (g groupByDenomContainer) add(address string, coins sdk.Coins) {
	for _, coin := range coins {
		io, ok := g[coin.Denom]
		if !ok {
			io = make(map[string]sdk.Int)
		}
		oldAmount, ok := io[address]
		if !ok {
			oldAmount = sdk.ZeroInt()
		}

		oldAmount = oldAmount.Add(coin.Amount)
		io[address] = oldAmount
		g[coin.Denom] = io
	}
}

func (g groupByDenomContainer) flatten() map[string][]AccAmount {
	result := make(map[string][]AccAmount)
	for denom, balances := range g {
		for acc, amount := range balances {
			accAddress, err := sdk.AccAddressFromBech32(acc)
			if err != nil {
				panic(err)
			}
			result[denom] = append(result[denom], AccAmount{Account: accAddress, Amount: amount})
		}
	}

	return result
}

func (k Keeper) applyRates(ctx sdk.Context, inputsAmounts, outputAmounts map[string][]AccAmount) error {
	for denom, inputAmounts := range inputsAmounts {
		ftd, err := k.GetTokenDefinition(ctx, denom)
		if types.ErrFTNotFound.Is(err) {
			return nil
		}

		outputAmounts := outputAmounts[denom]

		burnShares := CalculateRateShares(ftd.BurnRate, ftd.Issuer, inputAmounts, outputAmounts)
		for _, burnShare := range burnShares {
			if err := k.burn(ctx, burnShare.Account, ftd, burnShare.Amount); err != nil {
				return err
			}
		}

		commissionShares := CalculateRateShares(ftd.SendCommissionRate, ftd.Issuer, inputAmounts, outputAmounts)
		for _, commissionShare := range commissionShares {
			if err := k.burn(ctx, commissionShare.Account, ftd, commissionShare.Amount); err != nil {
				return err
			}
		}

		for _, in := range inputAmounts {
			if err := k.isCoinSpendable(ctx, in.Account, ftd, in.Amount); err != nil {
				return err
			}
		}

		for _, out := range outputAmounts {
			if err := k.isCoinReceivable(ctx, out.Account, ftd, out.Amount); err != nil {
				return err
			}
		}
	}

	return nil
}

// CalculateRateShares calculates how the burn amount should be split between different parties
func CalculateRateShares(rate sdk.Dec, issuer string, inputAmounts, outputAmounts []AccAmount) []AccAmount {
	if !rate.IsPositive() {
		return nil
	}

	nonIssuerSumFunc := func(sum sdk.Int, accAmount AccAmount, index int) sdk.Int {
		if accAmount.Account.String() != issuer {
			return sum.Add(accAmount.Amount)
		}
		return sum
	}
	inputSumNonIssuer := lo.Reduce(inputAmounts, nonIssuerSumFunc, sdk.ZeroInt())
	outputSumNonIssuer := lo.Reduce(outputAmounts, nonIssuerSumFunc, sdk.ZeroInt())

	minNonIssuer := inputSumNonIssuer
	if outputSumNonIssuer.LT(minNonIssuer) {
		minNonIssuer = outputSumNonIssuer
	}

	shares := make([]AccAmount, 0)
	if minNonIssuer.IsPositive() {
		for _, inpAmount := range inputAmounts {
			if inpAmount.Account.String() != issuer {
				burnAmount := rate.MulInt(minNonIssuer).Mul(sdk.NewDecFromInt(inpAmount.Amount)).QuoInt(inputSumNonIssuer).Ceil().RoundInt()
				shares = append(shares, AccAmount{Account: inpAmount.Account, Amount: burnAmount})
			}
		}
	}
	return shares
}
