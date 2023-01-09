package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

type IOPuts struct {
	Address string
	Coins   sdk.Coins
}

type AccAmount struct {
	Account string
	Amount  sdk.Int
}

type GroupedMultisend struct {
	Inputs  []AccAmount
	Outputs []AccAmount
}

func (k Keeper) applyFeatures2(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	inp := make([]IOPuts, len(inputs))
	for i, in := range inputs {
		inp[i] = IOPuts(in)
	}

	outp := make([]IOPuts, len(outputs))
	for i, out := range outputs {
		outp[i] = IOPuts(out)
	}

	squashedInp := squashIOPuts(inp)
	squashedOutp := squashIOPuts(outp)

	groupedInp := groupByDenom(squashedInp)
	groupedOutp := groupByDenom(squashedOutp)

	return k.applyRates(ctx, groupedInp, groupedOutp)
}

func squashIOPuts(inputs []IOPuts) []IOPuts {
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

	return lo.MapToSlice(accAddressCoins, func(accAddress string, coins sdk.Coins) IOPuts {
		return IOPuts{
			Address: accAddress,
			Coins:   coins,
		}
	})
}

func groupByDenom(ioputs []IOPuts) map[string][]AccAmount {
	denomToAccounts := make(map[string][]AccAmount)

	for _, ioput := range ioputs {
		for _, coin := range ioput.Coins {
			newAccAmount := AccAmount{
				Account: ioput.Address,
				Amount:  coin.Amount,
			}
			accounts, ok := denomToAccounts[coin.Denom]
			if ok {
				accounts = append(accounts, newAccAmount)
			} else {
				denomToAccounts[coin.Denom] = []AccAmount{newAccAmount}
			}
		}
	}

	return denomToAccounts
}

func (k Keeper) applyRates(ctx sdk.Context, inputsAmounts, outputAmounts map[string][]AccAmount) error {
	for denom, inpAmounts := range inputsAmounts {
		ftd, err := k.GetTokenDefinition(ctx, denom)
		if types.ErrFTNotFound.Is(err) {
			return nil
		}

		outpAmounts := outputAmounts[denom]
		adjustedBurnRate := CalculateAdjustedRate(ftd.BurnRate, ftd.Issuer, inpAmounts, outpAmounts)
		adjustedCommissionRate := CalculateAdjustedRate(ftd.SendCommissionRate, ftd.Issuer, inpAmounts, outpAmounts)

		for _, inpAmount := range inpAmounts {
			if inpAmount.Account != ftd.Issuer {
				// TODO: Move calculations to a single place.
				burnAmount := adjustedBurnRate.Mul(sdk.NewDecFromInt(inpAmount.Amount)).Ceil().RoundInt()
				if err := k.burn(ctx, sdk.AccAddress(inpAmount.Account), ftd, burnAmount); err != nil {
					return err
				}

				sendCommissionAmount := adjustedCommissionRate.Mul(sdk.NewDecFromInt(inpAmount.Amount)).Ceil().RoundInt()
				coins := sdk.NewCoins(sdk.NewCoin(ftd.Denom, sendCommissionAmount))
				err = k.bankKeeper.SendCoins(ctx, sdk.AccAddress(inpAmount.Account), sdk.AccAddress(ftd.Issuer), coins)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func CalculateAdjustedRate(initialRate sdk.Dec, issuer string, inputAmounts, outputAmounts []AccAmount) sdk.Dec {
	inputSum := lo.Reduce(inputAmounts, func(sum sdk.Int, accAmount AccAmount, index int) sdk.Int {
		return sum.Add(accAmount.Amount)
	}, sdk.ZeroInt())

	issuerSumFunc := func(sum sdk.Int, accAmount AccAmount, index int) sdk.Int {
		if accAmount.Account == issuer {
			return sum.Add(accAmount.Amount)
		}
		return sum
	}
	inputSumIssuer := lo.Reduce(inputAmounts, issuerSumFunc, sdk.ZeroInt())
	outputSumIssuer := lo.Reduce(outputAmounts, issuerSumFunc, sdk.ZeroInt())

	maxIssuer := inputSumIssuer
	if outputSumIssuer.GT(maxIssuer) {
		maxIssuer = outputSumIssuer
	}

	chargableAmount := inputSum.Sub(maxIssuer)
	return initialRate.MulInt(chargableAmount).QuoInt(inputSum)
}
