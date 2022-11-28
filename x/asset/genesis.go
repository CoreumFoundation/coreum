package asset

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/x/asset/keeper"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// InitGenesis initializes the asset module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Init fungible token definitions
	for _, definition := range genState.FungibleTokenDefinitions {
		k.SetFungibleTokenDefinition(ctx, definition)
	}

	// Init frozen balances
	for _, frozenBalance := range genState.FrozenBalances {
		address := sdk.MustAccAddressFromBech32(frozenBalance.Address)
		k.SetFrozenBalances(ctx, address, frozenBalance.Coins)
	}

	// Init Symbols
	for _, symbolIndex := range genState.Symbols {
		address := sdk.MustAccAddressFromBech32(symbolIndex.Address)
		for _, sl := range symbolIndex.Symbols {
			err := k.StoreSymbol(ctx, sl, address)
			if err != nil {
				panic(err)
			}
		}
	}
}

// ExportGenesis returns the asset module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// Export fungible token definitions
	fungibleTokenDefinitions, _, err := k.GetFungibleTokenDefinitions(ctx, &query.PageRequest{Limit: query.MaxLimit})
	if err != nil {
		panic(err)
	}

	// Export symbol index
	symbols, _, err := k.GetAllSymbols(ctx, &query.PageRequest{Limit: query.DefaultLimit})
	if err != nil {
		panic(err)
	}

	// Export frozen balances
	balances, _, err := k.GetAccountsFrozenBalances(ctx, &query.PageRequest{Limit: query.MaxLimit})
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		FungibleTokenDefinitions: fungibleTokenDefinitions,
		FrozenBalances:           balances,
		Symbols:                  symbols,
	}
}
