package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

type GovKeeper interface {
	GetParams(ctx sdk.Context) (params v1.Params)
}
