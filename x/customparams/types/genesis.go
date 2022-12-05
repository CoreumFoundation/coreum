package types

// DefaultGenesisState returns genesis state with default values
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		StakingParams: DefaultStakingParams(),
	}
}

// Validate validates genesis parameters
func (m *GenesisState) Validate() error {
	return m.StakingParams.ValidateBasic()
}
