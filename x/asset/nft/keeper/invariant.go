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
	FreezingInvariantName            = "freezing"
)

// RegisterInvariants registers the bank module invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, OriginalClassExistsInvariantName, OriginalClassExistsInvariant(k))
	ir.RegisterRoute(types.ModuleName, FreezingInvariantName, FreezingInvariant(k))
}

// FreezingInvariant checks that all frozen NFTs have counterpart on the original Cosmos SDK NFT module.
func FreezingInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg             string
			violationsCount int
		)

		_, frozenNFTs, err := k.GetFrozenNFTs(ctx, &query.PageRequest{Limit: query.MaxLimit})
		if err != nil {
			panic(err)
		}
		for _, frozen := range frozenNFTs {
			for _, nftID := range frozen.NftIDs {
				classDefinition, err := k.GetClassDefinition(ctx, frozen.ClassID)
				if types.ErrClassNotFound.Is(err) {
					violationsCount++
					msg += fmt.Sprintf("\t class definition not found for frozen nft(%s/%s)", frozen.ClassID, nftID)
				} else if err != nil {
					panic(err)
				}

				if !k.nftKeeper.HasNFT(ctx, frozen.ClassID, nftID) {
					violationsCount++
					msg += fmt.Sprintf("\t nft not found for frozen nft(%s/%s)", frozen.ClassID, nftID)
				}

				if !classDefinition.IsFeatureEnabled(types.ClassFeature_freezing) {
					violationsCount++
					msg += fmt.Sprintf("\t freezing is disabled, but (%s/%s) is frozen \n", frozen.ClassID, nftID)
				}
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, FreezingInvariantName,
			fmt.Sprintf("number of invariant violation %d\n%s", violationsCount, msg),
		), violationsCount != 0
	}
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
