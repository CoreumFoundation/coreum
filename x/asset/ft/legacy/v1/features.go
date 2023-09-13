package v1

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
)

// FTKeeper represents ft keeper.
type FTKeeper interface {
	IterateAllDefinitions(ctx sdk.Context, cb func(types.Definition) (bool, error)) error
	SetDefinition(ctx sdk.Context, issuer sdk.AccAddress, subunit string, definition types.Definition)
}

// MigrateFeatures migrates asset ft features state from v1 to v2.
// It removes features which are outside the allowed scope.
func MigrateFeatures(ctx sdk.Context, keeper FTKeeper) error {
	allowedFeatures := map[types.Feature]struct{}{
		types.Feature_minting:      {},
		types.Feature_burning:      {},
		types.Feature_freezing:     {},
		types.Feature_whitelisting: {},
	}

	return keeper.IterateAllDefinitions(ctx, func(def types.Definition) (bool, error) {
		present := map[types.Feature]struct{}{}
		newFeatures := make([]types.Feature, 0, len(def.Features))
		for _, f := range def.Features {
			if _, exists := allowedFeatures[f]; !exists {
				continue
			}
			if _, exists := present[f]; exists {
				continue
			}
			present[f] = struct{}{}
			newFeatures = append(newFeatures, f)
		}
		if len(newFeatures) < len(def.Features) {
			subunit, issuer, err := types.DeconstructDenom(def.Denom)
			if err != nil {
				return false, err
			}

			def.Features = newFeatures
			keeper.SetDefinition(ctx, issuer, subunit, def)
		}
		return false, nil
	})
}
