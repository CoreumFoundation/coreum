package benchmarks

import (
	"fmt"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/ibc-go/v3/testing/simapp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
)

func createSimApp(b *testing.B) *app.App {
	simapp.FlagEnabledValue = true
	_, db, dir, logger, _, err := simapp.SetupSimulation("coreum-app-sim", "Simulation")
	require.NoError(b, err, "simulation setup failed")

	b.Cleanup(func() {
		db.Close()
		err = os.RemoveAll(dir)
		require.NoError(b, err)
	})

	encoding := config.NewEncodingConfig(app.ModuleBasics)
	network, err := config.NetworkByChainID(config.Devnet)
	if err != nil {
		panic(err)
	}

	app.ChosenNetwork = network
	simApp := app.New(
		logger,
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		0,
		encoding,
		simapp.EmptyAppOptions{},
	)

	return simApp
}

func Benchmark100KDenomBankSend(b *testing.B) {
	simApp := createSimApp(b)
	bankKeeper := simApp.BankKeeper
	sdkContext := simApp.NewUncachedContext(false, types.Header{})
	singleCoinDenom := config.TokenSymbolDev
	coins := sdk.NewCoins(sdk.NewCoin(singleCoinDenom, sdk.NewInt(1_000_000_000)))
	err := bankKeeper.MintCoins(sdkContext, minttypes.ModuleName, coins)
	assert.NoError(b, err)

	b.ResetTimer()
	b.Run("test-single-send", func(b *testing.B) {
		b.StopTimer()
		var addresses []sdk.AccAddress
		for i := 0; i < b.N; i++ {
			address := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
			addresses = append(addresses, address)
			amount := sdk.NewCoins(sdk.NewCoin(singleCoinDenom, sdk.NewInt(10)))
			err = bankKeeper.SendCoinsFromModuleToAccount(sdkContext, minttypes.ModuleName, address, amount)
		}

		amount := sdk.NewCoins(sdk.NewCoin(singleCoinDenom, sdk.NewInt(10)))
		b.StartTimer()
		for i := 0; i < b.N; i++ {
			fromAddress := addresses[i]
			toAddress := addresses[(i+1)%len(addresses)]
			err = bankKeeper.SendCoins(sdkContext, fromAddress, toAddress, amount)
			assert.NoError(b, err)
		}
	})

	b.Run("test-100k-denom-send", func(b *testing.B) {
		b.StopTimer()
		var denoms []string
		for i := 0; i < 100_000; i++ {
			denom := fmt.Sprintf("test-denom-%d", i)
			denoms = append(denoms, denom)
			coins := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1_000_000_000)))
			err = bankKeeper.MintCoins(sdkContext, minttypes.ModuleName, coins)
			assert.NoError(b, err)
		}

		var addresses []sdk.AccAddress
		for i := 0; i < b.N; i++ {
			address := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
			addresses = append(addresses, address)

			denom := denoms[b.N%len(denoms)]
			amount := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(10)))
			err = bankKeeper.SendCoinsFromModuleToAccount(sdkContext, minttypes.ModuleName, address, amount)
		}

		b.StartTimer()
		for i := 0; i < b.N; i++ {
			fromAddress := addresses[i]
			toAddress := addresses[(i+1)%len(addresses)]
			denom := denoms[b.N%len(denoms)]
			amount := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(10)))
			err = bankKeeper.SendCoins(sdkContext, fromAddress, toAddress, amount)
			assert.NoError(b, err)
		}
	})
}

func Benchmark100KDenomBankModuleSend(b *testing.B) {
	simApp := createSimApp(b)
	bankKeeper := simApp.BankKeeper
	sdkContext := simApp.NewUncachedContext(false, types.Header{})
	singleCoinDenom := config.TokenSymbolDev
	coins := sdk.NewCoins(sdk.NewCoin(singleCoinDenom, sdk.NewInt(1_000_000_000)))
	err := bankKeeper.MintCoins(sdkContext, minttypes.ModuleName, coins)
	assert.NoError(b, err)

	b.ResetTimer()
	b.Run("test-single-module-send", func(b *testing.B) {
		b.StopTimer()
		var addresses []sdk.AccAddress
		for i := 0; i < b.N; i++ {
			address := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
			addresses = append(addresses, address)
		}

		b.StartTimer()
		for i := 0; i < b.N; i++ {
			address := addresses[i]
			amount := sdk.NewCoins(sdk.NewCoin(singleCoinDenom, sdk.NewInt(10)))
			err = bankKeeper.SendCoinsFromModuleToAccount(sdkContext, minttypes.ModuleName, address, amount)
			assert.NoError(b, err)
		}
	})

	b.Run("test-100k-denom-module-send", func(b *testing.B) {
		b.StopTimer()
		var denoms []string
		for i := 0; i < 100_000; i++ {
			denom := fmt.Sprintf("test-denom-%d", i)
			denoms = append(denoms, denom)
			coins := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1_000_000_000)))
			err = bankKeeper.MintCoins(sdkContext, minttypes.ModuleName, coins)
			assert.NoError(b, err)
		}

		var addresses []sdk.AccAddress
		for i := 0; i < b.N; i++ {
			address := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
			addresses = append(addresses, address)
		}

		b.StartTimer()
		for i := 0; i < b.N; i++ {
			address := addresses[i]
			denom := denoms[b.N%len(denoms)]
			amount := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(10)))
			err = bankKeeper.SendCoinsFromModuleToAccount(sdkContext, minttypes.ModuleName, address, amount)
			assert.NoError(b, err)
		}
	})
}

func Benchmark100KDenomBankGetSupply(b *testing.B) {
	simApp := createSimApp(b)
	bankKeeper := simApp.BankKeeper
	sdkContext := simApp.NewUncachedContext(false, types.Header{})

	singleCoinDenom := config.TokenSymbolDev
	coin := sdk.NewCoin(singleCoinDenom, sdk.NewInt(1_000_000_000))
	coins := sdk.NewCoins(coin)
	err := bankKeeper.MintCoins(sdkContext, minttypes.ModuleName, coins)
	assert.NoError(b, err)
	b.ResetTimer()
	b.Run("test-single-get-supply", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			supply := bankKeeper.GetSupply(sdkContext, singleCoinDenom)
			assert.NoError(b, err)
			assert.EqualValues(b, coin.String(), supply.String())
		}
	})

	var denoms []string
	mintValue := sdk.NewInt(1_000_000_000)
	for i := 0; i < 100_000; i++ {
		denom := fmt.Sprintf("test-denom-%d", i)
		denoms = append(denoms, denom)
		coins := sdk.NewCoins(sdk.NewCoin(denom, mintValue))
		err := bankKeeper.MintCoins(sdkContext, minttypes.ModuleName, coins)
		assert.NoError(b, err)
	}

	b.Run("test-100k-get-supply", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			denom := denoms[b.N%len(denoms)]
			supply := bankKeeper.GetSupply(sdkContext, denom)
			assert.EqualValues(b, mintValue, supply.Amount, "denom: %s", supply.Denom)
		}
	})
}
