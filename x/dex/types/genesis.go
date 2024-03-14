package types

// DefaultGenesis returns the default Token genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// TODO: Implement
	return nil
}
