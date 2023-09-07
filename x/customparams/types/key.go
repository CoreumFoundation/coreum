package types

const (
	// ModuleName defines the module name.
	ModuleName = "customparams"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key.
	RouterKey = ModuleName

	// CustomParamsStaking defines the params space key to store the staking custom params.
	CustomParamsStaking = "customparamsstaking"
)

// Store key prefixes.
var (
	// StakingParamsKey defines the key to store parameters of the module, set via governance.
	StakingParamsKey = []byte{0x30}
)
