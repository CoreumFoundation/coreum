package nft

const (
	// ModuleName module name.
	// The name is updated from "nft" to "cnft" to keep an ability to migrate to sdk native module.
	ModuleName = "cnft"

	// CLIName is used to name command in CLI.
	CLIName = "nft"

	// StoreKey is the default store key for nft.
	StoreKey = ModuleName

	// RouterKey is the message route for nft.
	RouterKey = ModuleName
)
