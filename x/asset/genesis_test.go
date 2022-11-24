package asset_test

import (
	"fmt"
	"math/rand"
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
					types.FungibleTokenFeature_freezable, //nolint:nosnakecase // proto enum
				},
			})
	}

	// fungible token frozen balances
	var fungibleTokenFrozenBalances []types.Balance
	for i := 0; i < 5; i++ {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		fungibleTokenFrozenBalances = append(fungibleTokenFrozenBalances,
			types.Balance{
				Address: addr.String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin(fungibleTokenDefinitions[0].Denom, sdk.NewInt(rand.Int63())),
					sdk.NewCoin(fungibleTokenDefinitions[1].Denom, sdk.NewInt(rand.Int63())),
				),
			})
	}

	genState := types.GenesisState{
		FrozenBalances:           fungibleTokenFrozenBalances,
		FungibleTokenDefinitions: fungibleTokenDefinitions,
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

	// check that export is equal import
	exportedGenState := asset.ExportGenesis(ctx, assetKeeper)
	assertT.ElementsMatch(genState.FrozenBalances, exportedGenState.FrozenBalances)
	assertT.ElementsMatch(genState.FungibleTokenDefinitions, exportedGenState.FungibleTokenDefinitions)
}
