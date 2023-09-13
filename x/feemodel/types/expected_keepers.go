package types

import (
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

// ParamsKeeper specifies expected methods of params keeper.
type ParamsKeeper interface {
	GetSubspace(s string) (paramstypes.Subspace, bool)
}
