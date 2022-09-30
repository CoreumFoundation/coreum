package types

// DefaultGenesis returns the default asset genesis state.
func DefaultGenesis() *GenesisState {
	// TODO(dhil) replace with real implementation
	return &GenesisState{}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// TODO(dhil) replace with real implementation
	return nil
}
