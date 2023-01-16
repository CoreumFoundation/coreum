package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

func (k Keeper) applyFeatures(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	groupInputs := make(groupedByDenomAccountOperations)
	for _, in := range inputs {
		groupInputs.add(in.Address, in.Coins)
	}

	groupOutputs := make(groupedByDenomAccountOperations)
	for _, out := range outputs {
		groupOutputs.add(out.Address, out.Coins)
	}

	return k.applyRates(ctx, groupInputs, groupOutputs)
}

type accountOperationMap map[string]sdk.Int

type groupedByDenomAccountOperations map[string]accountOperationMap

func (g groupedByDenomAccountOperations) add(address string, coins sdk.Coins) {
	for _, coin := range coins {
		accountBalances, ok := g[coin.Denom]
		if !ok {
			accountBalances = make(map[string]sdk.Int)
		}
		oldAmount, ok := accountBalances[address]
		if !ok {
			oldAmount = sdk.ZeroInt()
		}

		oldAmount = oldAmount.Add(coin.Amount)
		accountBalances[address] = oldAmount
		g[coin.Denom] = accountBalances
	}
}

func (k Keeper) applyRates(ctx sdk.Context, inputs, outputs groupedByDenomAccountOperations) error {
	for denom, inOps := range inputs {
		ftd, err := k.GetTokenDefinition(ctx, denom)
		if types.ErrInvalidDenom.Is(err) || types.ErrTokenNotFound.Is(err) {
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
		issuer := sdk.MustAccAddressFromBech32(ftd.Issuer)
		for account, amount := range commissionShares {
			coins := sdk.NewCoins(sdk.NewCoin(ftd.Denom, amount))
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
func CalculateRateShares(rate sdk.Dec, issuer string, inOps, outOps accountOperationMap) map[string]sdk.Int {
	// Since burning & send commission are not applied when sending to/from FT issuer we can't simply apply original burn rate or send commission rate when bank multisend with issuer in inputs or outputs.
	// To recalculate new adjusted amount we split whole "commission" between all non-issuer senders proportionally to amount they send.

	// Examples
	// burn_rate: 10%

	// inputs:
	// 75, 75
	// 25 <-- issuer

	// outputs:
	// 50
	// 100 <-- issuer
	// 25

	// In this case commissioned amount is: min(non_issuer_inputs, non_issuer_outputs) = min(75+75, 50+25) = 75
	// Expected commission: 75 * 10% = 7.5
	// And now we divide it proportionally between all input sender: 7.5 / 150 * 75 = 3.75
	// As result each sender is expected to pay 3.75 of commission.
	// Note that if we used original rate it would be 75 * 10% = 7.5
	// Here is the final formula we use to calculate adjusted burn/commission amount for multisend txs:
	// amount * rate * min(non_issuer_inputs_sum, non_issuer_outputs_sum) / non_issuer_inputs_sum
	if rate.IsNil() || !rate.IsPositive() {
		return nil
	}

	inputSumNonIssuer := nonIssuerSum(inOps, issuer)
	outputSumNonIssuer := nonIssuerSum(outOps, issuer)

	minNonIssuer := inputSumNonIssuer
	if outputSumNonIssuer.LT(minNonIssuer) {
		minNonIssuer = outputSumNonIssuer
	}

	if !minNonIssuer.IsPositive() {
		return nil
	}

	shares := make(accountOperationMap, 0)
	for account, amount := range inOps {
		if account != issuer {
			// in order to reduce precision errors, we first multiply all sdk.Ints, and then multiply sdk.Decs, and then divide
			finalShare := rate.MulInt(minNonIssuer.Mul(amount)).QuoInt(inputSumNonIssuer).Ceil().RoundInt()
			shares[account] = finalShare
		}
	}
	return shares
}

func nonIssuerSum(ops accountOperationMap, issuer string) sdk.Int {
	sum := sdk.ZeroInt()
	for account, amount := range ops {
		if account != issuer {
			sum = sum.Add(amount)
		}
	}
	return sum
}
