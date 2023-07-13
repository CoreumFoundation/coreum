package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum/v2/x/asset/ft/types"
	wibctransfertypes "github.com/CoreumFoundation/coreum/v2/x/wibctransfer/types"
)

// BeforeSendCoins checks that a transfer request is allowed or not.
func (k Keeper) BeforeSendCoins(ctx sdk.Context, fromAddress, toAddress sdk.AccAddress, coins sdk.Coins) error {
	return k.applyFeatures(
		ctx,
		[]banktypes.Input{{Address: fromAddress.String(), Coins: coins}},
		[]banktypes.Output{{Address: toAddress.String(), Coins: coins}},
	)
}

// BeforeInputOutputCoins extends InputOutputCoins method of the bank keeper.
func (k Keeper) BeforeInputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	return k.applyFeatures(ctx, inputs, outputs)
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

func (k Keeper) applyFeatures(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	groupInputs := make(groupedByDenomAccountOperations)
	for _, in := range inputs {
		groupInputs.add(in.Address, in.Coins)
	}

	groupOutputs := make(groupedByDenomAccountOperations)
	for _, out := range outputs {
		groupOutputs.add(out.Address, out.Coins)
	}

	return k.applyRules(ctx, groupInputs, groupOutputs)
}

func (k Keeper) applyRules(ctx sdk.Context, inputs, outputs groupedByDenomAccountOperations) error {
	for denom, inOps := range inputs {
		def, err := k.GetDefinition(ctx, denom)
		if types.ErrInvalidDenom.Is(err) || types.ErrTokenNotFound.Is(err) {
			return nil
		}

		outOps := outputs[denom]

		burnShares := k.CalculateRateShares(ctx, def.BurnRate, def.Issuer, inOps, outOps)
		for account, amount := range burnShares {
			if err := k.burnIfSpendable(ctx, sdk.MustAccAddressFromBech32(account), def, amount); err != nil {
				return err
			}
		}

		commissionShares := k.CalculateRateShares(ctx, def.SendCommissionRate, def.Issuer, inOps, outOps)
		issuer := sdk.MustAccAddressFromBech32(def.Issuer)
		for account, amount := range commissionShares {
			coins := sdk.NewCoins(sdk.NewCoin(def.Denom, amount))
			if err := k.bankKeeper.SendCoins(ctx, sdk.MustAccAddressFromBech32(account), issuer, coins); err != nil {
				return err
			}
		}

		for account, amount := range inOps {
			if err := k.isCoinSpendable(ctx, sdk.MustAccAddressFromBech32(account), def, amount); err != nil {
				return err
			}
		}

		for account, amount := range outOps {
			if err := k.isCoinReceivable(ctx, sdk.MustAccAddressFromBech32(account), def, amount); err != nil {
				return err
			}
		}
	}

	return nil
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

// CalculateRateShares calculates how the burn or commission share amount should be split between different parties.
func (k Keeper) CalculateRateShares(ctx sdk.Context, rate sdk.Dec, issuer string, inOps, outOps accountOperationMap) map[string]sdk.Int {
	// Since burning & send commission are not applied when sending to/from token issuer we can't simply apply original burn rate or send commission rate when bank multisend with issuer in inputs or outputs.
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
	if rate.IsNil() || !rate.IsPositive() || wibctransfertypes.IsPurposeIn(ctx) {
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
		// if sender is issuer or IBC escrow
		if account == issuer {
			continue
		}
		// in order to reduce precision errors, we first multiply all sdk.Ints, and then multiply sdk.Decs, and then divide
		finalShare := rate.MulInt(minNonIssuer.Mul(amount)).QuoInt(inputSumNonIssuer).Ceil().RoundInt()
		shares[account] = finalShare
	}

	return shares
}
