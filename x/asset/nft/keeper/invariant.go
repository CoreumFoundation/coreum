package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

// Invariant names.
const (
	OriginalClassExistsInvariantName = "original-class-exists"
	WhitelistingInvariantName        = "whitelisting"
	FreezingInvariantName            = "freezing"
)

// RegisterInvariants registers the bank module invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, OriginalClassExistsInvariantName, OriginalClassExistsInvariant(k))
}

// OriginalClassExistsInvariant checks that all the registered Classes have counterpart on the original Cosmos SDK NFT module.
func OriginalClassExistsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg           string
			notFoundCount int
		)

		classDefinitions, _, err := k.GetClassDefinitions(ctx, &query.PageRequest{Limit: query.MaxLimit})
		if err != nil {
			panic(err)
		}

		for _, classDef := range classDefinitions {
			found := k.nftKeeper.HasClass(ctx, classDef.ID)
			if !found {
				notFoundCount++
				msg += fmt.Sprintf("\t%s class does not have counterpart on original nft module", classDef.ID)
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, OriginalClassExistsInvariantName,
			fmt.Sprintf("number of missing original definitions %d\n%s", notFoundCount, msg),
		), notFoundCount != 0
	}
}
