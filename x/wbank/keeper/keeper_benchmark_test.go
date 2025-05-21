package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/pkg/config"
	"github.com/CoreumFoundation/coreum/v6/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/v6/testutil/simapp"
)

func Benchmark100KDenomBankSend(b *testing.B) {
	requireT := require.New(b)

	simApp := createSimApp(b)
	bankKeeper := simApp.BankKeeper
	sdkContext := simApp.NewUncachedContext(false, types.Header{})
	chainConfig, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	requireT.NoError(err)
	singleCoinDenom := chainConfig.Denom()
	coins := sdk.NewCoins(sdk.NewCoin(singleCoinDenom, sdkmath.NewInt(1_000_000_000)))
	requireT.NoError(bankKeeper.MintCoins(sdkContext, minttypes.ModuleName, coins))

	testAction := func(b *testing.B, numberOfDenoms int) {
		b.StopTimer()
		denoms := make([]string, numberOfDenoms)
		for i := range numberOfDenoms {
			denoms[i] = fmt.Sprintf("test-denom-%d", i)
			coins := sdk.NewCoins(sdk.NewCoin(denoms[i], sdkmath.NewInt(1_000_000_000)))
			requireT.NoError(bankKeeper.MintCoins(sdkContext, minttypes.ModuleName, coins))
		}

		addresses := make([]sdk.AccAddress, b.N)
		for i := range b.N {
			address := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
			addresses[i] = address

			denom := denoms[b.N%len(denoms)]
			amount := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(10)))
			requireT.NoError(bankKeeper.SendCoinsFromModuleToAccount(sdkContext, minttypes.ModuleName, address, amount))
		}

		b.StartTimer()
		for i := range b.N {
			fromAddress := addresses[i]
			toAddress := addresses[(i+1)%len(addresses)]
			denom := denoms[b.N%len(denoms)]
			amount := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(10)))
			requireT.NoError(bankKeeper.SendCoins(sdkContext, fromAddress, toAddress, amount))
		}
	}

	b.ResetTimer()
	b.Run("test-single-send", func(b *testing.B) {
		testAction(b, 1)
	})

	b.Run("test-100k-denom-send", func(b *testing.B) {
		testAction(b, 100_000)
	})
}

func Benchmark100KDenomBankGetSupply(b *testing.B) {
	requireT := require.New(b)

	simApp := createSimApp(b)
	bankKeeper := simApp.BankKeeper
	sdkContext := simApp.NewUncachedContext(false, types.Header{})

	chainConfig, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	requireT.NoError(err)
	singleCoinDenom := chainConfig.Denom()
	coin := sdk.NewCoin(singleCoinDenom, sdkmath.NewInt(1_000_000_000))
	coins := sdk.NewCoins(coin)
	requireT.NoError(bankKeeper.MintCoins(sdkContext, minttypes.ModuleName, coins))
	b.Run("test-single-get-supply", func(b *testing.B) {
		for range b.N {
			supply := bankKeeper.GetSupply(sdkContext, singleCoinDenom)
			assert.Equal(b, coin.String(), supply.String())
		}
	})

	var denoms []string
	mintValue := sdkmath.NewInt(1_000_000_000)
	for i := range 100_000 {
		denom := fmt.Sprintf("test-denom-%d", i)
		denoms = append(denoms, denom)
		coins := sdk.NewCoins(sdk.NewCoin(denom, mintValue))
		requireT.NoError(bankKeeper.MintCoins(sdkContext, minttypes.ModuleName, coins))
	}

	b.Run("test-100k-get-supply", func(b *testing.B) {
		for range b.N {
			denom := denoms[b.N%len(denoms)]
			supply := bankKeeper.GetSupply(sdkContext, denom)
			assert.Equal(b, mintValue, supply.Amount, "denom: %s", supply.Denom)
		}
	})
}

func createSimApp(b *testing.B) *simapp.App {
	db, err := dbm.NewDB("simulation", dbm.GoLevelDBBackend, b.TempDir())
	require.NoError(b, err)

	b.Cleanup(func() {
		db.Close()
	})

	return simapp.New(simapp.WithCustomDB(db))
}
