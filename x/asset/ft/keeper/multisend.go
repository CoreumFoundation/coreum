package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// BankOperation is a pairing between Account and Amount
type BankOperation struct {
	Account sdk.AccAddress
	Amount  sdk.Int
}

func (k Keeper) applyFeatures(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	inputGroup := make(groupedByDenomBankOperations)
	for _, in := range inputs {
		inputGroup.add(in.Address, in.Coins)
	}

	outputGroup := make(groupedByDenomBankOperations)
	for _, out := range outputs {
		outputGroup.add(out.Address, out.Coins)
	}

	return k.applyRates(ctx, inputGroup.flatten(), outputGroup.flatten())
}

type groupedByDenomBankOperations map[string]map[string]sdk.Int

func (g groupedByDenomBankOperations) add(address string, coins sdk.Coins) {
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

func (g groupedByDenomBankOperations) flatten() map[string][]BankOperation {
	result := make(map[string][]BankOperation)
	for denom, balances := range g {
		for acc, amount := range balances {
			accAddress, err := sdk.AccAddressFromBech32(acc)
			if err != nil {
				panic(err)
			}
			result[denom] = append(result[denom], BankOperation{Account: accAddress, Amount: amount})
		}
	}

	return result
}

func (k Keeper) applyRates(ctx sdk.Context, inputs, outputs map[string][]BankOperation) error {
	for denom, inOps := range inputs {
		ftd, err := k.GetTokenDefinition(ctx, denom)
		if types.ErrTokenNotFound.Is(err) {
			return nil
		}

		outOps := outputs[denom]

		burnShares := CalculateRateShares(ftd.BurnRate, ftd.Issuer, inOps, outOps)
		for _, burnShare := range burnShares {
			if err := k.burn(ctx, burnShare.Account, ftd, burnShare.Amount); err != nil {
				return err
			}
		}

		commissionShares := CalculateRateShares(ftd.SendCommissionRate, ftd.Issuer, inOps, outOps)
		for _, commissionShare := range commissionShares {
			coins := sdk.NewCoins(sdk.NewCoin(ftd.Denom, commissionShare.Amount))
			issuer := sdk.MustAccAddressFromBech32(ftd.Issuer)
			if err := k.bankKeeper.SendCoins(ctx, commissionShare.Account, issuer, coins); err != nil {
				return err
			}
		}

		for _, in := range inOps {
			if err := k.isCoinSpendable(ctx, in.Account, ftd, in.Amount); err != nil {
				return err
			}
		}

		for _, out := range outOps {
			if err := k.isCoinReceivable(ctx, out.Account, ftd, out.Amount); err != nil {
				return err
			}
		}
	}

	return nil
}

// CalculateRateShares calculates how the burn or commission share amount should be split between different parties
func CalculateRateShares(rate sdk.Dec, issuer string, inOps, outOps []BankOperation) []BankOperation {
	// The algorithm is as following. we first get the minimum of total inputs and outputs which are not
	// from the issuer. We then multiply by the rate to get applicable total amount. we then split this amount
	// between non-issuer senders, proportional to their input value.
	if !rate.IsPositive() {
		return nil
	}

	nonIssuerSumFunc := func(sum sdk.Int, accAmount BankOperation, index int) sdk.Int {
		if accAmount.Account.String() != issuer {
			return sum.Add(accAmount.Amount)
		}
		return sum
	}
	inputSumNonIssuer := lo.Reduce(inOps, nonIssuerSumFunc, sdk.ZeroInt())
	outputSumNonIssuer := lo.Reduce(outOps, nonIssuerSumFunc, sdk.ZeroInt())

	minNonIssuer := inputSumNonIssuer
	if outputSumNonIssuer.LT(minNonIssuer) {
		minNonIssuer = outputSumNonIssuer
	}

	shares := make([]BankOperation, 0)
	if minNonIssuer.IsPositive() {
		for _, inpAmount := range inOps {
			if inpAmount.Account.String() != issuer {
				amount := rate.MulInt(minNonIssuer.Mul(inpAmount.Amount)).QuoInt(inputSumNonIssuer).Ceil().RoundInt()
				shares = append(shares, BankOperation{Account: inpAmount.Account, Amount: amount})
			}
		}
	}
	return shares
}
