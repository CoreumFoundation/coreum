package keeper

import "github.com/CoreumFoundation/coreum/x/asset/types"

// isFeatureEnabled checks weather a feature is present on a list of token features
func isFeatureEnabled(features []types.FungibleTokenFeature, feature types.FungibleTokenFeature) bool {
	for _, o := range features {
		if o == feature {
			return true
		}
	}
	return false
}
