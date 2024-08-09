package upgrade

import (
	store "cosmossdk.io/store/types"
	upgradetypes "cosmossdk.io/x/upgrade/types"
)

// Upgrade defines the common structure for the chain upgrades.
type Upgrade struct {
	Name          string
	StoreUpgrades store.StoreUpgrades
	Upgrade       upgradetypes.UpgradeHandler
}
