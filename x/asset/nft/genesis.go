package nft

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/x/asset/nft/keeper"
	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	for _, definition := range genState.ClassDefinitions {
		k.SetClassDefinition(ctx, definition)
	}
	k.SetParams(ctx, genState.Params)

	for _, frozen := range genState.FrozenNFTs {
		for _, nftID := range frozen.NftIDs {
			if err := k.SetFrozen(ctx, frozen.ClassID, nftID, true); err != nil {
				panic(err)
			}
		}
	}

	for _, whitelisted := range genState.WhitelistedNFTAccounts {
		for _, account := range whitelisted.Accounts {
			if err := k.SetWhitelisting(
				ctx,
				whitelisted.ClassID,
				whitelisted.NftID,
				sdk.MustAccAddressFromBech32(account),
				true,
			); err != nil {
				panic(err)
			}
		}
	}

	for _, burnt := range genState.BurntNFTs {
		for _, nftID := range burnt.NftIDs {
			if err := k.SetBurnt(ctx, burnt.ClassID, nftID); err != nil {
				panic(err)
			}
		}
	}
}

// ExportGenesis returns the module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	classDefinitions, _, err := k.GetClassDefinitions(ctx, &query.PageRequest{Limit: query.MaxLimit})
	if err != nil {
		panic(err)
	}

	_, frozen, err := k.GetFrozenNFTs(ctx, &query.PageRequest{Limit: query.MaxLimit})
	if err != nil {
		panic(err)
	}

	_, whitelisted, err := k.GetAllWhitelisted(ctx, &query.PageRequest{Limit: query.MaxLimit})
	if err != nil {
		panic(err)
	}

	_, burnt, err := k.GetBurntNFTs(ctx, &query.PageRequest{Limit: query.MaxLimit})
	if err != nil {
		panic(err)
	}

	return &types.GenesisState{
		ClassDefinitions:       classDefinitions,
		Params:                 k.GetParams(ctx),
		FrozenNFTs:             frozen,
		WhitelistedNFTAccounts: whitelisted,
		BurntNFTs:              burnt,
	}
}
