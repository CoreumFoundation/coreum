package types

// Validate validates genesis parameters
func (m *GenesisState) Validate() error {
	return m.Params.Validate()
}
