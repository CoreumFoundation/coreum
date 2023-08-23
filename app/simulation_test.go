package app_test

import (
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simulationtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	clientcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v2/app"
	"github.com/CoreumFoundation/coreum/v2/pkg/config"
	"github.com/CoreumFoundation/coreum/v2/pkg/config/constant"
	testutilconstant "github.com/CoreumFoundation/coreum/v2/testutil/constant"
)

func init() {
	clientcli.GetSimulatorFlags()
}

// BenchmarkSimulation run the chain simulation
// Running using starport command:
// `starport chain simulate -v --numBlocks 200 --blockSize 50`
// Running as go benchmark test:
// `go test -benchmem -run=^$ -bench ^BenchmarkSimulation ./app -NumBlocks=200 -BlockSize 50`.
func BenchmarkSimulation(b *testing.B) {
	clientcli.FlagEnabledValue = true
	clientcli.FlagCommitValue = true

	cfg := clientcli.NewConfigFromFlags()
	cfg.ChainID = testutilconstant.SimAppChainID

	db, dir, logger, _, err := simtestutil.SetupSimulation(cfg, "goleveldb-app-sim", "Simulation", true, true)
	require.NoError(b, err, "simulation setup failed")

	b.Cleanup(func() {
		db.Close()
		err = os.RemoveAll(dir)
		require.NoError(b, err)
	})

	network, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	if err != nil {
		panic(err)
	}
	network.SetSDKConfig()

	app.ChosenNetwork = network
	simApp := app.New(
		logger,
		db,
		nil,
		true,
		simtestutil.EmptyAppOptions{},
		baseapp.SetChainID(cfg.ChainID),
	)

	// Run randomized simulations
	_, simParams, simErr := simulation.SimulateFromSeed(
		b,
		os.Stdout,
		simApp.GetBaseApp(),
		simtestutil.AppStateFn(simApp.AppCodec(), simApp.SimulationManager(), simApp.DefaultGenesis()),
		simulationtypes.RandomAccounts,
		simtestutil.SimulationOperations(simApp, simApp.AppCodec(), cfg),
		simApp.ModuleAccountAddrs(),
		cfg,
		simApp.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simtestutil.CheckExportSimulation(simApp, cfg, simParams)
	require.NoError(b, err)
	require.NoError(b, simErr)

	if cfg.Commit {
		simtestutil.PrintStats(db)
	}
}
