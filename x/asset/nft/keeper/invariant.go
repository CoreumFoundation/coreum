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
	ir.RegisterRoute(types.ModuleName, WhitelistingInvariantName, WhitelistingInvariant(k))
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
				found := k.nftKeeper.HasNFT(ctx, frozen.ClassID, nftID)
				if !found {
					violationsCount++
					msg += fmt.Sprintf("\t frozen nft (%s/%s) not registered on the original nft module", frozen.ClassID, nftID)
				}

				classDefinition, err := k.GetClassDefinition(ctx, frozen.ClassID)
				if types.ErrClassNotFound.Is(err) {
					violationsCount++
					msg += fmt.Sprintf("\t class definition not found for frozen nft(%s/%s)", frozen.ClassID, nftID)
				} else if err != nil {
					panic(err)
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

// WhitelistingInvariant checks that all whitelisted NFTs have counterpart on the original Cosmos SDK NFT module.
func WhitelistingInvariant(k Keeper) sdk.Invariant {
	return func(ctx sdk.Context) (string, bool) {
		var (
			msg             string
			violationsCount int
		)

		_, whitelistedNFTs, err := k.GetAllWhitelisted(ctx, &query.PageRequest{Limit: query.MaxLimit})
		if err != nil {
			panic(err)
		}
		for _, whitelisted := range whitelistedNFTs {
			found := k.nftKeeper.HasNFT(ctx, whitelisted.ClassID, whitelisted.NftID)
			if !found {
				violationsCount++
				msg += fmt.Sprintf("\t(%s/%s) nft not registered on the original nft module", whitelisted.ClassID, whitelisted.NftID)
			}
			classDefinition, err := k.GetClassDefinition(ctx, whitelisted.ClassID)

			if types.ErrClassNotFound.Is(err) {
				violationsCount++
				msg += fmt.Sprintf("\t class definition not found for whitelisted nft(%s/%s)", whitelisted.ClassID, whitelisted.NftID)
			} else if err != nil {
				panic(err)
			}

			if !classDefinition.IsFeatureEnabled(types.ClassFeature_whitelisting) {
				violationsCount++
				msg += fmt.Sprintf("\t whitelisting is disabled, but (%s/%s) is whitelisted \n", whitelisted.ClassID, whitelisted.NftID)
			}
		}

		return sdk.FormatInvariant(
			types.ModuleName, WhitelistingInvariantName,
			fmt.Sprintf("number of invariant violations %d\n%s", violationsCount, msg),
		), violationsCount != 0
	}
}

// OriginalClassExistsInvariant checks that all are registered Classes have counterpart on the original Cosmos SDK NFT module.
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
