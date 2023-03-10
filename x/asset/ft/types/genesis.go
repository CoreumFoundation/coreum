package types

// DefaultGenesis returns the default Token genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	for _, token := range gs.Tokens {
		if err := token.Validate(); err != nil {
			return err
		}
	}

	for _, balance := range gs.FrozenBalances {
		if err := ValidateAssetCoins(balance.Coins); err != nil {
			return err
		}
	}

	for _, balance := range gs.WhitelistedBalances {
		if err := ValidateAssetCoins(balance.Coins); err != nil {
			return err
		}
	}

	return gs.Params.ValidateBasic()
}

// Validate checks all the fields are valid.
func (token Token) Validate() error {
	_, _, err := DeconstructDenom(token.Denom)
	if err != nil {
		return err
	}

	if err := ValidateSymbol(token.Symbol); err != nil {
		return err
	}

	if err := ValidateSubunit(token.Subunit); err != nil {
		return err
	}

	if err := ValidatePrecision(token.Precision); err != nil {
		return err
	}

	if err := ValidateSendCommissionRate(token.SendCommissionRate); err != nil {
		return err
	}

	return ValidateBurnRate(token.BurnRate)
}
