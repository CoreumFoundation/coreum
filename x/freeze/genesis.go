package freeze

import (
	"github.com/CoreumFoundation/coreum/x/freeze/keeper"
	"github.com/CoreumFoundation/coreum/x/freeze/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init

	for _, frozenCoin := range genState.FrozenCoins {
		acc, err := sdk.AccAddressFromBech32(frozenCoin.Account)
		if err != nil {
			panic(err)
		}

		for _, coin := range frozenCoin.Coins {
			k.FreezeCoin(ctx, acc, coin)
		}
	}
}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	frozenCoins, err := k.ListFrozenCoins(ctx)
	if err != nil {
		panic(err)
	}

	for acc, coins := range frozenCoins {
		genesis.FrozenCoins = append(genesis.FrozenCoins, &types.AccFrozenCoins{
			Account: acc,
			Coins:   coins,
		})
	}

	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
