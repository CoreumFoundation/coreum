package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

const (
	// NonNegativeBalancesInvariantName is non negative balances invariant name.
	NonNegativeBalancesInvariantName = "non-negative-balances"
	// BankMetadataMatchesInvariantName is bank metadata matches name.
	BankMetadataMatchesInvariantName = "bank-metadata-matches"
)

// RegisterInvariants registers the bank module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, NonNegativeBalancesInvariantName, NonNegativeBalancesInvariant(k))
	ir.RegisterRoute(types.ModuleName, BankMetadataMatchesInvariantName, BankMetadataMatchesInvariant(k))
}

// NonNegativeBalancesInvariant checks that all accounts in the application have non-negative feature balances.
func NonNegativeBalancesInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg   string
			count int
		)

		err := k.IterateAllFrozenBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coin) bool {
			if balance.IsNegative() {
				count++
				msg += fmt.Sprintf("\t%s has a negative frozen balance of %s\n", addr, balance)
			}

			return false
		})
		if err != nil {
			count++
			msg += fmt.Sprintf("catched error on IterateAllFrozenBalances %s\n", err)
		}

		err = k.IterateAllWhitelistedBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coin) bool {
			if balance.IsNegative() {
				count++
				msg += fmt.Sprintf("\t%s has a negative whitelisted balance of %s\n", addr, balance)
			}

			return false
		})
		if err != nil {
			count++
			msg += fmt.Sprintf("catched error on IterateAllWhitelistedBalances %s\n", err)
		}

		return sdk.FormatInvariant(
			types.ModuleName, NonNegativeBalancesInvariantName,
			fmt.Sprintf("amount of negative balances found %d\n%s", count, msg),
		), count != 0
	}
}

// BankMetadataMatchesInvariant checks that all fungible tokens demons are in the bank as well.
func BankMetadataMatchesInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg   string
			count int
		)

		k.IterateAllTokenDefinitions(ctx, func(definition types.FTDefinition) bool {
			_, found := k.bankKeeper.GetDenomMetaData(ctx, definition.Denom)
			if !found {
				count++
				msg += fmt.Sprintf("\t%s denom doesn't have corresponding bank metadata", definition.Denom)
			}

			return false
		})

		return sdk.FormatInvariant(
			types.ModuleName, BankMetadataMatchesInvariantName,
			fmt.Sprintf("number of missing metadata entries %d\n%s", count, msg),
		), count != 0
	}
}
