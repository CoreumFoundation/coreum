package types

const (
	// ModuleName defines the module name
	// "fee" is already taken by feegrant module
	ModuleName = "coreumfee"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// TransientStoreKey defines the transient module store key
	TransientStoreKey = "transient_" + ModuleName
)
