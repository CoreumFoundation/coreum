package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	customparamstypes "github.com/CoreumFoundation/coreum/x/customparams/types"
)

// CustomParamsKeeper defines the custom params keeper interface required for the module.
type CustomParamsKeeper interface {
	GetStakingParams(ctx sdk.Context) customparamstypes.StakingParams
}
