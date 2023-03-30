package genesis

import (
	_ "embed"
)

var (
	//go:embed coreum-mainnet-1.json
	// MainnetGenesis is mainnet genesis file bytes.
	MainnetGenesis []byte

	//go:embed coreum-testnet-1.json
	// TestnetGenesis is testnet genesis file bytes.
	TestnetGenesis []byte
)
