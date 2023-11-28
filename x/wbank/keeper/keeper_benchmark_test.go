package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	dbm "github.com/cometbft/cometbft-db"
	"github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/pkg/config"
	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
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
		for i := 0; i < numberOfDenoms; i++ {
			denoms[i] = fmt.Sprintf("test-denom-%d", i)
			coins := sdk.NewCoins(sdk.NewCoin(denoms[i], sdkmath.NewInt(1_000_000_000)))
			requireT.NoError(bankKeeper.MintCoins(sdkContext, minttypes.ModuleName, coins))
		}

		addresses := make([]sdk.AccAddress, b.N)
		for i := 0; i < b.N; i++ {
			address := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
			addresses[i] = address

			denom := denoms[b.N%len(denoms)]
			amount := sdk.NewCoins(sdk.NewCoin(denom, sdkmath.NewInt(10)))
			requireT.NoError(bankKeeper.SendCoinsFromModuleToAccount(sdkContext, minttypes.ModuleName, address, amount))
		}

		b.StartTimer()
		for i := 0; i < b.N; i++ {
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
		for i := 0; i < b.N; i++ {
			supply := bankKeeper.GetSupply(sdkContext, singleCoinDenom)
			assert.EqualValues(b, coin.String(), supply.String())
		}
	})

	var denoms []string
	mintValue := sdkmath.NewInt(1_000_000_000)
	for i := 0; i < 100_000; i++ {
		denom := fmt.Sprintf("test-denom-%d", i)
		denoms = append(denoms, denom)
		coins := sdk.NewCoins(sdk.NewCoin(denom, mintValue))
		requireT.NoError(bankKeeper.MintCoins(sdkContext, minttypes.ModuleName, coins))
	}

	b.Run("test-100k-get-supply", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			denom := denoms[b.N%len(denoms)]
			supply := bankKeeper.GetSupply(sdkContext, denom)
			assert.EqualValues(b, mintValue, supply.Amount, "denom: %s", supply.Denom)
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
