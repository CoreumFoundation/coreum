package asset_test

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/crypto/ed25519"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestImportAndExportGenesis(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	testApp := simapp.New()

	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	assetKeeper := testApp.AssetKeeper
	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	// prepare the genesis data

	// fungible token definitions
	var fungibleTokenDefinitions []types.FungibleTokenDefinition
	for i := 0; i < 5; i++ {
		fungibleTokenDefinitions = append(fungibleTokenDefinitions,
			types.FungibleTokenDefinition{
				Denom:  types.BuildFungibleTokenDenom(fmt.Sprintf("ABC%d", i), issuer),
				Issuer: issuer.String(),
				Features: []types.FungibleTokenFeature{
					types.FungibleTokenFeature_freeze,    //nolint:nosnakecase // proto enum
					types.FungibleTokenFeature_whitelist, //nolint:nosnakecase // proto enum
				},
			})
	}

	// fungible token frozen balances
	var fungibleTokenFrozenBalances []types.FungibleTokenBalance
	for i := 0; i < 5; i++ {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		fungibleTokenFrozenBalances = append(fungibleTokenFrozenBalances,
			types.FungibleTokenBalance{
				Address: addr.String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin(fungibleTokenDefinitions[0].Denom, sdk.NewInt(rand.Int63())),
					sdk.NewCoin(fungibleTokenDefinitions[1].Denom, sdk.NewInt(rand.Int63())),
				),
			})
	}

	// fungible token whitelisted balances
	var fungibleTokenWhitelistedBalances []types.FungibleTokenBalance
	for i := 0; i < 5; i++ {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		fungibleTokenWhitelistedBalances = append(fungibleTokenWhitelistedBalances,
			types.FungibleTokenBalance{
				Address: addr.String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin(fungibleTokenDefinitions[0].Denom, sdk.NewInt(rand.Int63())),
					sdk.NewCoin(fungibleTokenDefinitions[1].Denom, sdk.NewInt(rand.Int63())),
				),
			})
	}

	genState := types.GenesisState{
		FungibleTokens: types.FungibleTokenState{
			TokenDefinitions:    fungibleTokenDefinitions,
			FrozenBalances:      fungibleTokenFrozenBalances,
			WhitelistedBalances: fungibleTokenWhitelistedBalances,
		},
	}

	// init the keeper
	asset.InitGenesis(ctx, assetKeeper, genState)

	// assert the keeper state

	// fungible token definitions
	for _, definition := range fungibleTokenDefinitions {
		storedDefinition, err := assetKeeper.GetFungibleTokenDefinition(ctx, definition.Denom)
		requireT.NoError(err)
		assertT.EqualValues(definition, storedDefinition)
	}

	// fungible token frozen balances
	for _, balance := range fungibleTokenFrozenBalances {
		address, err := sdk.AccAddressFromBech32(balance.Address)
		requireT.NoError(err)
		coins, _, err := assetKeeper.GetFrozenBalances(ctx, address, nil)
		requireT.NoError(err)
		assertT.EqualValues(balance.Coins.String(), coins.String())
	}

	// fungible token whitelisted balances
	for _, balance := range fungibleTokenWhitelistedBalances {
		address, err := sdk.AccAddressFromBech32(balance.Address)
		requireT.NoError(err)
		coins, _, err := assetKeeper.GetWhitelistedBalances(ctx, address, nil)
		requireT.NoError(err)
		assertT.EqualValues(balance.Coins.String(), coins.String())
	}

	// check that export is equal import
	exportedGenState := asset.ExportGenesis(ctx, assetKeeper)

	sort.Slice(genState.FungibleTokens.TokenDefinitions, func(i, j int) bool {
		return strings.Compare(
			genState.FungibleTokens.TokenDefinitions[i].Denom,
			genState.FungibleTokens.TokenDefinitions[j].Denom,
		) < 0
	})

	sort.Slice(genState.FungibleTokens.FrozenBalances, func(i, j int) bool {
		return strings.Compare(
			genState.FungibleTokens.FrozenBalances[i].Coins.String()+genState.FungibleTokens.FrozenBalances[i].Address,
			genState.FungibleTokens.FrozenBalances[j].Coins.String()+genState.FungibleTokens.FrozenBalances[j].Address,
		) < 0
	})

	sort.Slice(genState.FungibleTokens.WhitelistedBalances, func(i, j int) bool {
		return strings.Compare(
			genState.FungibleTokens.WhitelistedBalances[i].Coins.String()+genState.FungibleTokens.WhitelistedBalances[i].Address,
			genState.FungibleTokens.WhitelistedBalances[j].Coins.String()+genState.FungibleTokens.WhitelistedBalances[j].Address,
		) < 0
	})

	sort.Slice(exportedGenState.FungibleTokens.TokenDefinitions, func(i, j int) bool {
		return strings.Compare(
			exportedGenState.FungibleTokens.TokenDefinitions[i].Denom,
			exportedGenState.FungibleTokens.TokenDefinitions[j].Denom,
		) < 0
	})

	sort.Slice(exportedGenState.FungibleTokens.FrozenBalances, func(i, j int) bool {
		return strings.Compare(
			exportedGenState.FungibleTokens.FrozenBalances[i].Coins.String()+exportedGenState.FungibleTokens.FrozenBalances[i].Address,
			exportedGenState.FungibleTokens.FrozenBalances[j].Coins.String()+exportedGenState.FungibleTokens.FrozenBalances[j].Address,
		) < 0
	})

	sort.Slice(exportedGenState.FungibleTokens.WhitelistedBalances, func(i, j int) bool {
		return strings.Compare(
			exportedGenState.FungibleTokens.WhitelistedBalances[i].Coins.String()+exportedGenState.FungibleTokens.WhitelistedBalances[i].Address,
			exportedGenState.FungibleTokens.WhitelistedBalances[j].Coins.String()+exportedGenState.FungibleTokens.WhitelistedBalances[j].Address,
		) < 0
	})

	assertT.EqualValues(genState.FungibleTokens, exportedGenState.FungibleTokens)
}
