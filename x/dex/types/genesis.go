package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DefaultGenesis returns the default Token genesis state.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic genesis state validation returning an error upon any failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.ValidateBasic(); err != nil {
		return err
	}
	denoms := make(map[string]struct{})
	for _, ob := range gs.OrderBooks {
		denoms[ob.Data.BaseDenom] = struct{}{}
		denoms[ob.Data.QuoteDenom] = struct{}{}
	}
	usedSequence := make(map[uint64]struct{})
	for _, order := range gs.Orders {
		if _, ok := usedSequence[order.Sequence]; ok {
			return sdkerrors.Wrapf(ErrInvalidInput, "duplicate order sequence %d", order.Sequence)
		}
		usedSequence[order.Sequence] = struct{}{}

		if _, ok := denoms[order.BaseDenom]; !ok {
			return sdkerrors.Wrapf(ErrInvalidInput, "base denom %s does not exist in order books", order.BaseDenom)
		}
		if _, ok := denoms[order.QuoteDenom]; !ok {
			return sdkerrors.Wrapf(ErrInvalidInput, "quote denom %s does not exist in order books", order.QuoteDenom)
		}

		order.Sequence = 0
		order.RemainingBaseQuantity = sdkmath.Int{}
		order.RemainingSpendableBalance = sdkmath.Int{}
		order.Reserve = sdk.Coin{}

		if err := order.Validate(); err != nil {
			return err
		}
	}

	return nil
}
