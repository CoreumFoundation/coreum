package v3

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

// MigrateDefinitions migrates asset ft definitions state from v2 to v3.
// It sets admin the same as issuer for all previously issued tokens.
func MigrateDefinitions(ctx sdk.Context, keeper FTKeeper) error {
	return keeper.IterateAllDefinitions(ctx, func(def types.Definition) (bool, error) {
		subunit, issuer, err := types.DeconstructDenom(def.Denom)
		if err != nil {
			return false, err
		}

		def.Admin = def.Issuer
		if err := keeper.SetDefinition(ctx, issuer, subunit, def); err != nil {
			return false, err
		}
		return false, nil
	})
}
