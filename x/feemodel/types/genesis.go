package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

// DefaultGenesisState returns genesis state with default values
func DefaultGenesisState() *GenesisState {
	params := DefaultParams()
	return &GenesisState{
		Params:      params,
		MinGasPrice: sdk.NewCoin(sdk.DefaultBondDenom, params.InitialGasPrice),
	}
}

// Validate validates genesis parameters
func (m *GenesisState) Validate() error {
	if err := m.MinGasPrice.Validate(); err != nil {
		return errors.WithStack(err)
	}
	return m.Params.Validate()
}
