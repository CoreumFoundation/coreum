package asset_test

import (
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

func TestExportGenesis(t *testing.T) {
	assertT := assert.New(t)
	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	assetKeeper := testApp.AssetKeeper
	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	denom := types.BuildFungibleTokenDenom("ABC", issuer)
	denom2 := types.BuildFungibleTokenDenom("ABC2", issuer)

	var balances []types.Balance
	for i := 0; i < 20; i++ {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		coins := sdk.NewCoins(
			sdk.NewCoin(denom, sdk.NewInt(rand.Int63())),
			sdk.NewCoin(denom2, sdk.NewInt(rand.Int63())),
		)
		balances = append(balances, types.Balance{Address: addr.String(), Coins: coins})
		assetKeeper.SetFrozenBalances(ctx, addr, coins)
	}

	genState := asset.ExportGenesis(ctx, assetKeeper)
	assertT.ElementsMatch(balances, genState.FrozenBalances)
}

func TestImportGenesis(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)
	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	assetKeeper := testApp.AssetKeeper
	issuer := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	denom := types.BuildFungibleTokenDenom("ABC", issuer)
	denom2 := types.BuildFungibleTokenDenom("ABC2", issuer)

	var balances []types.Balance
	for i := 0; i < 20; i++ {
		addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
		balances = append(balances,
			types.Balance{
				Address: addr.String(),
				Coins: sdk.NewCoins(
					sdk.NewCoin(denom, sdk.NewInt(rand.Int63())),
					sdk.NewCoin(denom2, sdk.NewInt(rand.Int63())),
				)})
	}

	genState := types.GenesisState{
		FrozenBalances: balances,
	}

	asset.InitGenesis(ctx, assetKeeper, genState)

	for _, balance := range balances {
		address, err := sdk.AccAddressFromBech32(balance.Address)
		requireT.NoError(err)
		coins, _, err := assetKeeper.GetFrozenBalances(ctx, address, nil)
		requireT.NoError(err)
		assertT.EqualValues(balance.Coins.String(), coins.String())
	}
}
