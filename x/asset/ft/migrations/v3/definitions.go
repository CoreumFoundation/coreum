package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

// FTKeeper represents ft keeper.
type FTKeeper interface {
	IterateAllDefinitions(ctx sdk.Context, cb func(types.Definition) (bool, error)) error
	SetDefinition(ctx sdk.Context, issuer sdk.AccAddress, subunit string, definition types.Definition)
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
		keeper.SetDefinition(ctx, issuer, subunit, def)
		return false, nil
	})
}
