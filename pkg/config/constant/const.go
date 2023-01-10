package constant

const (
	// CoinType is the CORE coin type as defined in SLIP44 (https://github.com/satoshilabs/slips/blob/master/slip-0044.md)
	CoinType uint32 = 990
)

// ChainID represents predefined chain ID
type ChainID string

// Predefined chain ids
const (
	ChainIDMain ChainID = "coreum-mainnet-1"
	ChainIDDev  ChainID = "coreum-devnet-1"
)

// Denom names
const (
	DenomDev         = "udevcore"
	DenomDevDisplay  = "devcore"
	DenomTest        = "utestcore"
	DenomTestDisplay = "testcore"
	DenomMain        = "ucore"
	DenomMainDisplay = "core"
)

// Address prefixes
const (
	AddressPrefixDev  = "devcore"
	AddressPrefixTest = "testcore"
	AddressPrefixMain = "core"
)
