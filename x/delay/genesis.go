package delay

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v6/x/delay/keeper"
	"github.com/CoreumFoundation/coreum/v6/x/delay/types"
)

// DefaultGenesis returns the default genesis state.
func DefaultGenesis() *types.GenesisState {
	return &types.GenesisState{}
}

// InitGenesis initializes the state from a provided genesis.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	if err := k.ImportDelayedItems(ctx, genState.DelayedItems); err != nil {
		panic(err)
	}
	if err := k.ImportBlockItems(ctx, genState.BlockItems); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	delayedItems, err := k.ExportDelayedItems(ctx)
	if err != nil {
		panic(err)
	}
	blockItems, err := k.ExportBlockItems(ctx)
	if err != nil {
		panic(err)
	}
	return &types.GenesisState{
		DelayedItems: delayedItems,
		BlockItems:   blockItems,
	}
}
