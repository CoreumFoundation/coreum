package types

import (
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type AssetFTKeeper interface {
	GetDefinition(ctx sdk.Context, denom string) (assetfttypes.Definition, error)
}
