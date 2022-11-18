package types

// DefaultGenesisState returns genesis state with default values
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate validates genesis parameters
func (m *GenesisState) Validate() error {
	if err := m.Params.ValidateBasic(); err != nil {
		return err
	}

	return nil
}
