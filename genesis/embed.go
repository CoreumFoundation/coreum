package genesis

import (
	_ "embed"
)

var (
	//go:embed coreum-mainnet-1.json
	MainnetGenesis []byte

	//go:embed coreum-testnet-1.json
	TestnetGenesis []byte
)
