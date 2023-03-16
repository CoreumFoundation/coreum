package upgrade

import (
	store "github.com/cosmos/cosmos-sdk/store/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// Upgrade defines the common structure for the chain upgrades.
type Upgrade struct {
	Name          string
	StoreUpgrades store.StoreUpgrades
	Upgrade       upgradetypes.UpgradeHandler
}
