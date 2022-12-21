package ft

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/x/asset/ft/keeper"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// TODO(yaroslav): Add global freezing logic to genesis once coreum #268 is merged.

// InitGenesis initializes the asset module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Init fungible token definitions
	for _, ft := range genState.Tokens {
		issuerAddress := sdk.MustAccAddressFromBech32(ft.Issuer)
		definition := types.FTDefinition{
			Denom:    ft.Denom,
			Issuer:   ft.Issuer,
			Features: ft.Features,
			BurnRate: ft.BurnRate,
		}
		k.SetTokenDefinition(ctx, definition)
		err := k.StoreSymbol(ctx, ft.Symbol, issuerAddress)
		if err != nil {
			panic(err)
		}
		if ft.GloballyFrozen {
			k.SetGlobalFreeze(ctx, ft.Denom, true)
		}
	}

	// Init frozen balances
	for _, frozenBalance := range genState.FrozenBalances {
		address := sdk.MustAccAddressFromBech32(frozenBalance.Address)
		k.SetFrozenBalances(ctx, address, frozenBalance.Coins)
	}

	// Init whitelisted balances
	for _, whitelistedBalance := range genState.WhitelistedBalances {
		address := sdk.MustAccAddressFromBech32(whitelistedBalance.Address)
		k.SetWhitelistedBalances(ctx, address, whitelistedBalance.Coins)
	}
}

// ExportGenesis returns the asset module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	// Export fungible token definitions
	tokens, _, err := k.GetTokens(ctx, &query.PageRequest{Limit: query.MaxLimit})
	if err != nil {
		panic(err)
	}

	// Export frozen balances
	frozenBalances, _, err := k.GetAccountsFrozenBalances(ctx, &query.PageRequest{Limit: query.MaxLimit})
	if err != nil {
		panic(err)
	}

	// Export whitelisted balances
	whitelistedBalances, _, err := k.GetAccountsWhitelistedBalances(ctx, &query.PageRequest{Limit: query.MaxLimit})
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Tokens:              tokens,
		FrozenBalances:      frozenBalances,
		WhitelistedBalances: whitelistedBalances,
	}
}
