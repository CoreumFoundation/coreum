package app_test

import (
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	simulationtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	clientcli "github.com/cosmos/cosmos-sdk/x/simulation/client/cli"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/app"
	"github.com/CoreumFoundation/coreum/v6/pkg/config"
	"github.com/CoreumFoundation/coreum/v6/pkg/config/constant"
	testutilconstant "github.com/CoreumFoundation/coreum/v6/testutil/constant"
)

func init() {
	clientcli.GetSimulatorFlags()
}

// FullSimulation run the chain simulation
// Running as go test:
//
//	`go test ./app -run TestFullAppSimulation -v -Enabled=true \
//		-Verbose=true -NumBlocks=100 -BlockSize=200 -Commit=true -Period=5`.
func TestFullAppSimulation(t *testing.T) {
	if !clientcli.FlagEnabledValue {
		t.Skip()
		return
	}

	cfg := clientcli.NewConfigFromFlags()
	cfg.ChainID = testutilconstant.SimAppChainID

	db, dir, logger, _, err := simtestutil.SetupSimulation(
		cfg,
		"goleveldb-app-sim",
		"Simulation",
		clientcli.FlagVerboseValue,
		clientcli.FlagEnabledValue,
	)
	require.NoError(t, err, "simulation setup failed")

	t.Cleanup(func() {
		db.Close()
		err = os.RemoveAll(dir)
		require.NoError(t, err)
	})

	network, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	if err != nil {
		panic(err)
	}
	network.SetSDKConfig()
	app.ChosenNetwork = network

	appOptions := make(simtestutil.AppOptionsMap, 0)
	appOptions[flags.FlagHome] = dir // ensure a unique folder
	appOptions[server.FlagInvCheckPeriod] = clientcli.FlagPeriodValue

	simApp := app.New(
		logger,
		db,
		nil,
		true,
		appOptions,
		fauxMerkleModeOpt,
		baseapp.SetChainID(cfg.ChainID),
	)

	// Run randomized simulations
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
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
	require.NoError(t, err)
	require.NoError(t, simErr)

	if cfg.Commit {
		simtestutil.PrintStats(db)
	}
}

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}
