package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
)

const (
	// FreezingInvariantName is frozen balances invariant name.
	FreezingInvariantName = "freezing"
	// WhitelistingInvariantName is whitelisted balances invariant name.
	WhitelistingInvariantName = "whitelisting"
	// BankMetadataExistsInvariantName is bank metadata exist name.
	BankMetadataExistsInvariantName = "bank-metadata-exist"
)

// RegisterInvariants registers the bank module invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, FreezingInvariantName, FreezingInvariant(k))
	ir.RegisterRoute(types.ModuleName, WhitelistingInvariantName, WhitelistingInvariant(k))
	ir.RegisterRoute(types.ModuleName, BankMetadataExistsInvariantName, BankMetadataExistInvariant(k))
}

// FreezingInvariant checks that all accounts in the application have non-negative frozen balances.
func FreezingInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			count int
			msg   string
		)

		definitions := make(map[string]types.Definition)
		err := k.IterateAccountsFrozenBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coin) bool {
			count, msg = applyFeatureBalanceInvariant(ctx, k, addr, balance, types.Feature_freezing, definitions, count, msg)
			return false
		})
		if err != nil {
			count++
			msg += fmt.Sprintf("can't iterate over frozen balances %s\n", err)
		}

		return sdk.FormatInvariant(
			types.ModuleName, FreezingInvariantName,
			fmt.Sprintf("amount of invalid frozen balances found: %d\n%s", count, msg),
		), count != 0
	}
}

// WhitelistingInvariant checks that all accounts in the application have non-negative whitelisted balances.
func WhitelistingInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			count int
			msg   string
		)

		definitions := make(map[string]types.Definition)
		err := k.IterateAccountsWhitelistedBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coin) bool {
			count, msg = applyFeatureBalanceInvariant(ctx, k, addr, balance, types.Feature_whitelisting, definitions, count, msg)
			return false
		})
		if err != nil {
			count++
			msg += fmt.Sprintf("can't iterate over whitelisted balances %s\n", err)
		}

		return sdk.FormatInvariant(
			types.ModuleName, WhitelistingInvariantName,
			fmt.Sprintf("amount of invalid whitelisted balances found: %d\n%s", count, msg),
		), count != 0
	}
}

// BankMetadataExistInvariant checks that all fungible tokens demons are in the bank as well.
func BankMetadataExistInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg   string
			count int
		)

		err := k.IterateAllDefinitions(ctx, func(definition types.Definition) (bool, error) {
			_, found := k.bankKeeper.GetDenomMetaData(ctx, definition.Denom)
			if !found {
				count++
				msg += fmt.Sprintf("\t%s denom doesn't have corresponding bank metadata", definition.Denom)
			}

			return false, nil
		})
		if err != nil {
			// impossible
			panic(err)
		}

		return sdk.FormatInvariant(
			types.ModuleName, BankMetadataExistsInvariantName,
			fmt.Sprintf("number of missing metadata entries %d\n%s", count, msg),
		), count != 0
	}
}

func applyFeatureBalanceInvariant(
	ctx sdk.Context,
	k Keeper,
	addr sdk.AccAddress,
	balance sdk.Coin,
	feature types.Feature,
	definitions map[string]types.Definition,
	count int,
	msg string,
) (int, string) {
	if balance.IsNegative() {
		count++
		msg += fmt.Sprintf("\taddress %s has a negative %s balance: %s\n", addr, feature, balance)
	}

	definition, ok := definitions[balance.Denom]
	if !ok {
		var err error
		definition, err = k.GetDefinition(ctx, balance.Denom)
		if err != nil {
			count++
			msg += fmt.Sprintf("\t definition for the %s denom not found\n", balance.Denom)
			return count, msg
		}
		definitions[balance.Denom] = definition
	}

	if !definition.IsFeatureEnabled(feature) && balance.IsPositive() {
		count++
		msg += fmt.Sprintf("\t feature %s is disabled, but %s balance %s is positive\n", feature.String(), feature.String(), balance)
	}

	return count, msg
}
