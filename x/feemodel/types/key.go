package types

const (
	// ModuleName defines the module name.
	ModuleName = "feemodel"

	// StoreKey defines the primary module store key.
	StoreKey = ModuleName

	// TransientStoreKey defines the transient module store key.
	TransientStoreKey = "transient_" + ModuleName

	// RouterKey defines the module's message routing key.
	RouterKey = ModuleName
)

// Store key prefixes.
var (
	// ParamsKey defines the key to store parameters of the module, set via governance.
	ParamsKey = []byte{0x01}
)
