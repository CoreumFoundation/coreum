//go:build integrationtests

package export_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"strconv"

	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"

	"github.com/CoreumFoundation/coreum/v6/app"
	integrationtests "github.com/CoreumFoundation/coreum/v6/integration-tests"
	"github.com/CoreumFoundation/coreum/v6/testutil/simapp"
)

func executeCommand(ctx context.Context, buf *bytes.Buffer, name string, args []string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func getModuleHashesAtHeight(app *app.App, height int64) (*storetypes.CommitInfo, error) {
	rms, ok := app.CommitMultiStore().(*rootmulti.Store)
	if !ok {
		return nil, fmt.Errorf("expected rootmulti.Store, got %T", app.CommitMultiStore())
	}

	commitInfoForHeight, err := rms.GetCommitInfo(height)
	if err != nil {
		return nil, err
	}

	// Create a new slice of StoreInfos for storing the modified hashes.
	storeInfos := make([]storetypes.StoreInfo, len(commitInfoForHeight.StoreInfos))

	for i, storeInfo := range commitInfoForHeight.StoreInfos {
		// Convert the hash to a hexadecimal string.
		hash := strings.ToUpper(hex.EncodeToString(storeInfo.CommitId.Hash))

		// Create a new StoreInfo with the modified hash.
		storeInfos[i] = storetypes.StoreInfo{
			Name: storeInfo.Name,
			CommitId: storetypes.CommitID{
				Version: storeInfo.CommitId.Version,
				Hash:    []byte(hash),
			},
		}
	}

	// Sort the storeInfos slice based on the module name.
	sort.Slice(storeInfos, func(i, j int) bool {
		return storeInfos[i].Name < storeInfos[j].Name
	})

	// Create a new CommitInfo with the modified StoreInfos.
	commitInfoForHeight = &storetypes.CommitInfo{
		Version:    commitInfoForHeight.Version,
		StoreInfos: storeInfos,
		Timestamp:  commitInfoForHeight.Timestamp,
	}

	return commitInfoForHeight, nil
}

func TestExportGenesisModuleHashes(t *testing.T) {
	refCtx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	znetHomeDir := os.Getenv("ZNET_HOME_DIR")
	t.Logf("ZNet home directory: %s", znetHomeDir)

	fullNodeHome := filepath.Join(znetHomeDir, "app", "cored-05-full")
	t.Logf("Full home directory: %s", fullNodeHome)

	valNodeHome := filepath.Join(znetHomeDir, "app", "cored-00-val")
	t.Logf("Validator home directory: %s", valNodeHome)

	fullNodeBin := filepath.Join(os.Getenv("CORED_BIN_DIR"), "cored")

	logBuffer := new(bytes.Buffer)
	logger := log.NewLogger(logBuffer, log.ColorOption(false))

	nodeHome := filepath.Join(fullNodeHome, chain.ClientContext.ChainID())
	nodeDb, err := dbm.NewDB("application", dbm.GoLevelDBBackend, filepath.Join(nodeHome, "data"))
	requireT.NoError(err, "failed to open node DB at %s", nodeHome)

	nodeApp := app.New(logger, nodeDb, nil, false, simtestutil.AppOptionsMap{
		flags.FlagHome:            nodeHome,
		server.FlagInvCheckPeriod: time.Millisecond * 100,
	})
	exportHeight := nodeApp.LastBlockHeight()

	var exportBuf bytes.Buffer
	err = executeCommand(
		refCtx,
		&exportBuf,
		fullNodeBin,
		[]string{
			"export",
			"--for-zero-height",
			"--height", strconv.FormatInt(exportHeight, 10),
			"--log_level", "disabled",
			"--chain-id", chain.ClientContext.ChainID(),
			"--home", valNodeHome,
		},
	)
	requireT.NoError(err, "failed to execute export command")

	newApp, _, _, initChainReq, initChainRes := simapp.NewWithGenesis(exportBuf.Bytes())
	newAppCtx := newApp.NewContext(false)

	_, err = newApp.App.FinalizeBlock(&abci.RequestFinalizeBlock{
		Hash:   initChainRes.AppHash,
		Height: initChainReq.InitialHeight + 1,
	})
	require.NoError(t, err)

	_, err = newApp.Commit()
	require.NoError(t, err)

	require.NoError(t, err)
	require.Equal(t, initChainReq.InitialHeight+1, newApp.LastBlockHeight())

	customStakingParams, err := newApp.CustomParamsKeeper.GetStakingParams(newAppCtx)
	requireT.NoError(err, "failed to get staking params from new app")
	requireT.Equal(customStakingParams.MinSelfDelegation.Int64(), int64(10000000), "staking params should not be empty in new app")

	validators, err := newApp.StakingKeeper.GetAllValidators(newAppCtx)
	requireT.NoError(err, "failed to get all validators from new app")
	requireT.NotEmpty(validators, "validators should not be empty in new app")

	nodeAppModuleHashes, err := getModuleHashesAtHeight(nodeApp, exportHeight)
	requireT.NoError(err, "failed to get module hashes at height")

	newAppModuleHashes, err := getModuleHashesAtHeight(&newApp.App, 1)
	requireT.NoError(err, "failed to get module hashes at height")

	// Compare StoreInfos one by one and list mismatches
	// Match StoreInfos by Name, not index
	nodeStoreMap := make(map[string]storetypes.StoreInfo)
	newAppStoreMap := make(map[string]storetypes.StoreInfo)
	for _, si := range nodeAppModuleHashes.StoreInfos {
		nodeStoreMap[si.Name] = si
	}
	for _, si := range newAppModuleHashes.StoreInfos {
		newAppStoreMap[si.Name] = si
	}

	var mismatches []string
	// Check for mismatches and missing in newApp
	for name, ns := range nodeStoreMap {
		newAppStore, ok := newAppStoreMap[name]
		if !ok {
			mismatches = append(mismatches, fmt.Sprintf("Missing StoreInfo in newApp: %s", name))
			continue
		}
		if !bytes.Equal(ns.CommitId.Hash, newAppStore.CommitId.Hash) {
			mismatches = append(mismatches, fmt.Sprintf(
				"Hash mismatch for module %s: nodeApp=%s, newApp=%s",
				name, ns.CommitId.Hash, newAppStore.CommitId.Hash,
			))
		}
	}
	// Check for extra modules in newApp
	for name := range newAppStoreMap {
		if _, ok := nodeStoreMap[name]; !ok {
			mismatches = append(mismatches, fmt.Sprintf("Extra StoreInfo in newApp: %s", name))
		}
	}
	if len(mismatches) > 0 {
		t.Fatalf("StoreInfo mismatches:\n%s", strings.Join(mismatches, "\n"))
	}
}
