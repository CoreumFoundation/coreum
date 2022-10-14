package types

const (
	// ModuleName defines the module name
	ModuleName = "snapshot"

	// StoreKey defines the primary module store key
	StoreKey = ModuleName

	// RouterKey defines the module's message routing key
	RouterKey = ModuleName

	// QuerierRoute defines the module's query routing key
	QuerierRoute = ModuleName
)

var (
	SnapshotsKey               = []byte{0x00}
	PendingSnapshotSubkey      = []byte{0x00}
	CurrentValueSnapshotSubkey = []byte{0x01}
	DataSubkey                 = []byte{0x02}

	FreezeRequestsKey      = []byte{0x01}
	FreezeRequestsIndexKey = []byte{0x02}
	FrozenKey              = []byte{0x03}
)
