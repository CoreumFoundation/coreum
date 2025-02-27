package v4

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
)

// FTKeeper represents ft keeper.
type FTKeeper interface {
	IterateAllDefinitions(ctx sdk.Context, cb func(types.Definition) (bool, error)) error
	SetDefinition(ctx sdk.Context, issuer sdk.AccAddress, subunit string, definition types.Definition) error
}

// ParamsKeeper specifies methods of params keeper required by the migration.
type ParamsKeeper interface {
	GetSubspace(s string) (paramstypes.Subspace, bool)
}

// MigrateDefinitions migrates asset ft definitions state.
func MigrateDefinitions(ctx sdk.Context, keeper FTKeeper) error {
	definitions := []types.Definition{}
	err := keeper.IterateAllDefinitions(ctx, func(def types.Definition) (bool, error) {
		// for extension without ibc we add it because we apply the ft validation for the extension starting
		// from the current version
		if def.IsFeatureEnabled(types.Feature_extension) &&
			!def.IsFeatureEnabled(types.Feature_ibc) {
			def.Features = append(def.Features, types.Feature_ibc)
		}

		if !def.IsFeatureEnabled(types.Feature_dex_unified_ref_amount_change) {
			def.Features = append(def.Features, types.Feature_dex_unified_ref_amount_change)
		}
		definitions = append(definitions, def)

		return false, nil
	})

	if err != nil {
		return err
	}

	for _, def := range definitions {
		subunit, issuer, err := types.DeconstructDenom(def.Denom)
		if err != nil {
			return err
		}
		if err := keeper.SetDefinition(ctx, issuer, subunit, def); err != nil {
			return err
		}
	}

	return nil
}
