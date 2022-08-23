package types

const (
	// ModuleName defines the module name
	ModuleName = "fee"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// TransientStoreKey defines the transient module store key
	TransientStoreKey = "transient_" + ModuleName
)
