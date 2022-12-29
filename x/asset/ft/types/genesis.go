package types

// DefaultGenesis returns the default FT genesis state.
func DefaultGenesis() *GenesisState {
	// TODO(dhil) replace with real implementation
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	// TODO(dhil) replace with real implementation
	return gs.Params.ValidateBasic()
}
