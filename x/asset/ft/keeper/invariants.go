package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// RegisterInvariants registers the bank module invariants
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, "non-negative-balances", NonNegativeBalancesInvariant(k))
	ir.RegisterRoute(types.ModuleName, "bank-metadata-matches", BankMetadataMatchesInvariant(k))
}

// NonNegativeBalancesInvariant checks that all accounts in the application have non-negative fungible token specific balances.
func NonNegativeBalancesInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg   string
			count int
		)

		k.IterateAllFrozenBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coin) bool {
			if balance.IsNegative() {
				count++
				msg += fmt.Sprintf("\t%s has a negative frozen balance of %s\n", addr, balance)
			}

			return false
		})

		k.IterateAllWhitelistedBalances(ctx, func(addr sdk.AccAddress, balance sdk.Coin) bool {
			if balance.IsNegative() {
				count++
				msg += fmt.Sprintf("\t%s has a whitelisted frozen balance of %s\n", addr, balance)
			}

			return false
		})

		return sdk.FormatInvariant(
			types.ModuleName, "non-negative-balances",
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
			types.ModuleName, "bank-metadata-matches",
			fmt.Sprintf("amount of broken metadata found %d\n%s", count, msg),
		), count != 0
	}
}
