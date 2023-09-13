package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/v3/x/asset/nft/types"
)

// Invariant names.
const (
	OriginalClassExistsInvariantName = "original-class-exists"
	FreezingInvariantName            = "freezing"
	BurntNFTInvariantName            = "burnt-nft"
)

// RegisterInvariants registers the bank module invariants.
func RegisterInvariants(ir sdk.InvariantRegistry, k Keeper) {
	ir.RegisterRoute(types.ModuleName, OriginalClassExistsInvariantName, OriginalClassExistsInvariant(k))
	ir.RegisterRoute(types.ModuleName, FreezingInvariantName, FreezingInvariant(k))
	ir.RegisterRoute(types.ModuleName, BurntNFTInvariantName, BurntNFTInvariant(k))
}

// FreezingInvariant checks that all frozen NFTs have counterpart on the original Cosmos SDK nft module.
func FreezingInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg             string
			violationsCount int
		)

		frozenNFTs, _, err := k.GetFrozenNFTs(ctx, &query.PageRequest{Limit: query.MaxLimit})
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

// OriginalClassExistsInvariant checks that all the registered Classes have counterpart on the original Cosmos SDK nft module.
func OriginalClassExistsInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg           string
			notFoundCount int
		)

		err := k.IterateAllClassDefinitions(ctx, func(classDef types.ClassDefinition) (bool, error) {
			found := k.nftKeeper.HasClass(ctx, classDef.ID)
			if !found {
				notFoundCount++
				msg += fmt.Sprintf("\t%s class does not have counterpart on original nft module", classDef.ID)
			}

			return false, nil
		})
		if err != nil {
			panic(err)
		}

		return sdk.FormatInvariant(
			types.ModuleName, OriginalClassExistsInvariantName,
			fmt.Sprintf("number of missing original definitions %d\n%s", notFoundCount, msg),
		), notFoundCount != 0
	}
}

// BurntNFTInvariant checks that all burnt NFT registered in assetnft module don't exist in the Cosmos SDK nft module.
func BurntNFTInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg            string
			violationCount int
		)

		burntNFTs, _, err := k.GetBurntNFTs(ctx, &query.PageRequest{Limit: query.MaxLimit})
		if err != nil {
			panic(err)
		}

		for _, burntNFT := range burntNFTs {
			for _, id := range burntNFT.NftIDs {
				if k.nftKeeper.HasNFT(ctx, burntNFT.ClassID, id) {
					violationCount++
					msg += fmt.Sprintf("\t burnt NFT exists in the nft module, classID %s, ID: %s", burntNFT.ClassID, id)
				}
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, BurntNFTInvariantName,
			fmt.Sprintf("number of found not burnt NFTs %d\n%s", violationCount, msg),
		), violationCount != 0
	}
}
