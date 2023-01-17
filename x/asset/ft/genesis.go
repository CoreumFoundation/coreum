package ft

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/x/asset/ft/keeper"
	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// InitGenesis initializes the asset module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	k.SetParams(ctx, genState.Params)

	// Init fungible token definitions
	for _, token := range genState.Tokens {
		subunit, issuer, err := types.DeconstructDenom(token.Denom)
		if err != nil {
			panic(err)
		}

		definition := types.Definition{
			Denom:              token.Denom,
			Issuer:             token.Issuer,
			Features:           token.Features,
			BurnRate:           token.BurnRate,
			SendCommissionRate: token.SendCommissionRate,
		}

		k.SetTokenDefinition(ctx, issuer, subunit, definition)

		err = k.SetSymbol(ctx, token.Symbol, issuer)
		if err != nil {
			panic(err)
		}
		if token.GloballyFrozen {
			k.SetGlobalFreeze(ctx, token.Denom, true)
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
		Params:              k.GetParams(ctx),
		Tokens:              tokens,
		FrozenBalances:      frozenBalances,
		WhitelistedBalances: whitelistedBalances,
	}
}
