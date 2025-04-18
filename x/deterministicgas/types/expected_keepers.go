package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	assetfttypes "github.com/CoreumFoundation/coreum/v6/x/asset/ft/types"
)

// AssetFTKeeper is the expected keeper from the assetft module.
type AssetFTKeeper interface {
	GetDefinition(ctx sdk.Context, denom string) (assetfttypes.Definition, error)
}
