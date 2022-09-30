package benchmarks

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/ibc-go/v3/testing/simapp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
)

func Benchmark100KDenomBankTransfer(b *testing.B) {
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

	bankKeeper := simApp.BankKeeper

	sdkContext := simApp.NewUncachedContext(false, types.Header{})
	singleCoinDenom := config.TokenSymbolDev
	coins := sdk.NewCoins(sdk.NewCoin(singleCoinDenom, sdk.NewInt(1000*1000_000)))
	err = bankKeeper.MintCoins(sdkContext, minttypes.ModuleName, coins)
	assert.NoError(b, err)

	ctx := sdk.WrapSDKContext(sdkContext)
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	b.Cleanup(cancel)
	totalSupply, err := bankKeeper.TotalSupply(ctx, &banktypes.QueryTotalSupplyRequest{})
	assert.NoError(b, err)
	assert.EqualValues(b, coins, totalSupply.Supply)

	b.ResetTimer()

	b.Run("test-single-transfer", func(b *testing.B) {
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

	b.Run("test-100k-denom-transfer", func(b *testing.B) {
		b.StopTimer()
		var denoms []string
		for i := 0; i < 100_000; i++ {
			denom := fmt.Sprintf("test-denom-%d", i)
			denoms = append(denoms, denom)
			coins := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(1000_000_000)))
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
