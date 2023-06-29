package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

// NFTKeeper represents nft keeper.
type NFTKeeper interface {
	IterateAllClassDefinitions(ctx sdk.Context, cb func(types.ClassDefinition) (bool, error)) error
	SetClassDefinition(ctx sdk.Context, definition types.ClassDefinition) error
}

// MigrateClassFeatures migrates asset nft class features state from v1 to v2.
// It removes features which are outside the allowed scope.
func MigrateClassFeatures(ctx sdk.Context, keeper NFTKeeper) error {
	return keeper.IterateAllClassDefinitions(ctx, func(classDef types.ClassDefinition) (bool, error) {
		present := map[types.ClassFeature]struct{}{}
		newFeatures := make([]types.ClassFeature, 0, len(classDef.Features))
		for _, f := range classDef.Features {
			if _, exists := types.ClassFeature_name[int32(f)]; !exists {
				continue
			}
			if _, exists := present[f]; exists {
				continue
			}
			present[f] = struct{}{}
			newFeatures = append(newFeatures, f)
		}
		if len(newFeatures) < len(classDef.Features) {
			classDef.Features = newFeatures

			if err := keeper.SetClassDefinition(ctx, classDef); err != nil {
				return false, err
			}
		}
		return false, nil
	})
}
