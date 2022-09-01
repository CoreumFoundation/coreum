package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

func DefaultGenesisState() *GenesisState {
	model := DefaultModel()
	return &GenesisState{
		Params:      model,
		MinGasPrice: sdk.NewCoin("cosmos", model.InitialGasPrice),
	}
}

// Validate validates genesis parameters
func (m *GenesisState) Validate() error {
	if err := m.MinGasPrice.Validate(); err != nil {
		return errors.WithStack(err)
	}
	return m.Params.Validate()
}
