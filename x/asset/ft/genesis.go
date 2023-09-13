package ft

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/v3/x/asset/ft/keeper"
	"github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
)

// InitGenesis initializes the asset module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	// Init fungible token definitions
	for _, token := range genState.Tokens {
		if err := token.Validate(); err != nil {
			panic(err)
		}

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
			Version:            token.Version,
		}

		k.SetDefinition(ctx, issuer, subunit, definition)

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
		if err := types.ValidateAssetCoins(frozenBalance.Coins); err != nil {
			panic(err)
		}
		address := sdk.MustAccAddressFromBech32(frozenBalance.Address)
		k.SetFrozenBalances(ctx, address, frozenBalance.Coins)
	}

	// Init whitelisted balances
	for _, whitelistedBalance := range genState.WhitelistedBalances {
		if err := types.ValidateAssetCoins(whitelistedBalance.Coins); err != nil {
			panic(err)
		}
		address := sdk.MustAccAddressFromBech32(whitelistedBalance.Address)
		k.SetWhitelistedBalances(ctx, address, whitelistedBalance.Coins)
	}

	// Init pending version upgrades
	if err := k.ImportPendingTokenUpgrades(ctx, genState.PendingTokenUpgrades); err != nil {
		panic(err)
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

	pendingTokenUpgrades, err := k.ExportPendingTokenUpgrades(ctx)
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		Params:               k.GetParams(ctx),
		Tokens:               tokens,
		FrozenBalances:       frozenBalances,
		WhitelistedBalances:  whitelistedBalances,
		PendingTokenUpgrades: pendingTokenUpgrades,
	}
}
