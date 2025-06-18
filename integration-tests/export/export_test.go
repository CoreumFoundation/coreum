//go:build integrationtests

package export_test

import (
	"bytes"
	"context"
	"flag"
	"fmt"

	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	nftkeeper "cosmossdk.io/x/nft/keeper"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/server"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	ibchost "github.com/cosmos/ibc-go/v8/modules/core/24-host"
	ibcexported "github.com/cosmos/ibc-go/v8/modules/core/exported"

	"github.com/CoreumFoundation/coreum/v6/app"
	"github.com/CoreumFoundation/coreum/v6/pkg/config"
	"github.com/CoreumFoundation/coreum/v6/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/v6/testutil/simapp"
)

var runUnsafe bool

func init() {
	flag.BoolVar(&runUnsafe, "run-unsafe", false, "run unsafe tests for example ones related to governance")
}

// ignoredModuleNames defines module names that should be ignored in entire process.
var ignoredModuleNames = map[string]struct{}{
	// TODO: fix Error calling the VM: Cache error: Error opening Wasm file for reading
	"wasm": {},
}

// ignoredPrefixes defines prefixes of keys that should be ignored during KVStore comparison.
// These keys are typically used for internal state management and do not affect the exported genesis.
var ignoredPrefixes = map[string][][]byte{
	stakingtypes.StoreKey: {
		stakingtypes.UnbondingIDKey,
		stakingtypes.UnbondingIndexKey,
		stakingtypes.UnbondingTypeKey,
		stakingtypes.UnbondingQueueKey,
		stakingtypes.HistoricalInfoKey,
	},
	distributiontypes.StoreKey: {
		distributiontypes.FeePoolKey,
		distributiontypes.ProposerKey,
	},
	nftkeeper.StoreKey: {
		nftkeeper.ClassTotalSupply,
	},
	ibcexported.StoreKey: {
		ibchost.KeyClientStorePrefix,
	},
}

func TestExportGenesisModuleHashes(t *testing.T) {
	requireT := require.New(t)

	chainId := string(constant.ChainIDDev)

	coredBinDir := os.Getenv("CORED_BIN_DIR")
	t.Logf("Cored binary directory: %s", coredBinDir)

	znetHomeDir := os.Getenv("ZNET_HOME_DIR")
	t.Logf("ZNet home directory: %s", znetHomeDir)

	znetAppDir := filepath.Join(znetHomeDir, "app")
	fullNodeHome := filepath.Join(znetAppDir, "cored-05-full")
	t.Logf("Full home directory: %s", fullNodeHome)

	logBuffer := new(bytes.Buffer)
	logger := log.NewLogger(logBuffer, log.ColorOption(false))

	currentTs := time.Now().Unix()

	copiedNodeHome := filepath.Join(znetAppDir, fmt.Sprintf("%d-cored-05-full", currentTs))
	// Only copy if the directory does not exist
	_, statErr := os.Stat(copiedNodeHome)
	if os.IsNotExist(statErr) {
		err := executeCommand(
			t.Context(),
			new(bytes.Buffer),
			"cp",
			[]string{"-a", fullNodeHome, copiedNodeHome},
		)
		requireT.NoError(err, "failed to copy data directory")
	} else {
		requireT.NoError(statErr, "failed to stat copied data directory")
		t.Logf("Copied node home already exists, skipping copy: %s", copiedNodeHome)
	}
	copiedNodeChainHome := filepath.Join(copiedNodeHome, chainId)
	copiedDBDir := filepath.Join(copiedNodeChainHome, "data")
	// requireT.NoError(err, "failed to copy data directory")
	nodeDb, err := dbm.NewDB("application", dbm.GoLevelDBBackend, copiedDBDir)
	requireT.NoError(err, "failed to open node DB at %s", copiedDBDir)
	t.Log("Copied node DB directory:", copiedDBDir)

	network, err := config.NetworkConfigByChainID(constant.ChainID(chainId))
	if err != nil {
		panic(errors.WithStack(err))
	}
	app.ChosenNetwork = network
	network.SetSDKConfig()

	nodeApp := app.New(logger, nodeDb, nil, false, simtestutil.AppOptionsMap{
		flags.FlagHome:            copiedNodeChainHome,
		server.FlagInvCheckPeriod: time.Millisecond * 100,
	})

	fullNodeBin := filepath.Join(coredBinDir, "cored")
	exportJSONPath := filepath.Join(znetAppDir, fmt.Sprintf("%d-exported-genesis.json", currentTs))
	var exportBuf bytes.Buffer

	// Filter out ignored modules
	var modulesToExport []string
	for _, m := range nodeApp.ModuleManager.ModuleNames() {
		if _, ignored := ignoredModuleNames[m]; !ignored {
			modulesToExport = append(modulesToExport, m)
		}
	}

	// File does not exist, run export command and write to file
	err = executeCommand(
		t.Context(),
		&exportBuf,
		fullNodeBin,
		[]string{
			"export",
			fmt.Sprintf("--log_level=%s", "disabled"),
			fmt.Sprintf("--modules-to-export=%s", strings.Join(modulesToExport, ",")),
			fmt.Sprintf("--chain-id=%s", chainId),
			fmt.Sprintf("--home=%s", fullNodeHome),
		},
	)
	requireT.NoError(err, "failed to execute export command")
	err = os.WriteFile(exportJSONPath, exportBuf.Bytes(), 0644)
	requireT.NoError(err, "failed to write export JSON to file")
	t.Logf("Wrote export JSON to %s", exportJSONPath)

	newApp, _, _, initChainReq, _ := simapp.NewWithGenesis(exportBuf.Bytes())

	// load the latest version from the node app
	// the initial height is the height that need to gets finalized in the new app
	nodeAppStateHeight := initChainReq.InitialHeight - 1
	err = nodeApp.LoadVersion(nodeAppStateHeight)
	requireT.NoError(err, "failed to load version %d from node app", nodeAppStateHeight)

	// finalize new block for the node app
	_, err = nodeApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: initChainReq.InitialHeight,
	})
	require.NoError(t, err)
	_, err = nodeApp.Commit()
	require.NoError(t, err)

	// finalize new block for the new app
	_, err = newApp.App.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: initChainReq.InitialHeight,
	})
	require.NoError(t, err)

	_, err = newApp.Commit()
	require.NoError(t, err)

	err = checkModuleStoreMismatches(t, nodeApp, &newApp.App, initChainReq.InitialHeight)
	requireT.NoError(err, "failed to compare module stores between node app and new app")
}

func executeCommand(ctx context.Context, buf *bytes.Buffer, name string, args []string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func compareKVStores(nodeStore, newStore corestore.KVStore, ignorePrefixes [][]byte) error {
	// Build maps of key-value pairs for both stores
	nodeMap := make(map[string][]byte)
	iter1, err := nodeStore.Iterator(nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create iterator for node store: %w", err)
	}
	defer iter1.Close()
	for ; iter1.Valid(); iter1.Next() {
		nodeMap[string(iter1.Key())] = append([]byte(nil), iter1.Value()...)
	}

	newMap := make(map[string][]byte)
	iter2, err := newStore.Iterator(nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create iterator for new store: %w", err)
	}
	defer iter2.Close()
	for ; iter2.Valid(); iter2.Next() {
		newMap[string(iter2.Key())] = append([]byte(nil), iter2.Value()...)
	}

	var mismatches []string

	// Check for keys in nodeMap not in newMap or with different values
	for k, v := range nodeMap {
		// Skip keys with any of the ignorePrefixes
		ignored := false
		for _, prefix := range ignorePrefixes {
			if bytes.HasPrefix([]byte(k), prefix) {
				ignored = true
				break
			}
		}
		if ignored {
			continue
		}

		nv, ok := newMap[k]
		if !ok {
			mismatches = append(mismatches, fmt.Sprintf("key %q missing in new store", k))
			continue
		}
		if !bytes.Equal(v, nv) {
			mismatches = append(mismatches, fmt.Sprintf("value mismatch for key %q: %q vs %q", k, v, nv))
		}
	}

	// Check for extra keys in newMap
	for k := range newMap {
		// Skip keys with any of the ignorePrefixes
		ignored := false
		for _, prefix := range ignorePrefixes {
			if bytes.HasPrefix([]byte(k), prefix) {
				ignored = true
				break
			}
		}
		if ignored {
			continue
		}

		if _, ok := nodeMap[k]; !ok {
			mismatches = append(mismatches, fmt.Sprintf("extra key %X in new store", []byte(k)))
		}
	}

	if len(mismatches) > 0 {
		return fmt.Errorf("KVStore mismatches:\n%s", strings.Join(mismatches, "\n"))
	}
	return nil
}

func checkModuleStoreMismatches(t *testing.T, nodeApp *app.App, newApp *app.App, height int64) error {
	var mismatches []string

	nodeAppCtx := nodeApp.NewUncachedContext(false, cmtproto.Header{Height: height})
	newAppCtx := newApp.NewUncachedContext(false, cmtproto.Header{Height: height})

	for _, moduleName := range nodeApp.ModuleManager.ModuleNames() {
		if _, ok := ignoredModuleNames[moduleName]; ok {
			continue
		}

		storeName := moduleName
		if moduleName == authtypes.StoreKey {
			storeName = "acc"
		}

		var ignorePrefixes [][]byte
		if prefixes, ok := ignoredPrefixes[moduleName]; ok {
			ignorePrefixes = prefixes
		}

		var nodeAppKvStore, newAppKvStore corestore.KVStore
		// panic happens if the store is not registered in the app,
		func() {
			defer func() {
				if r := recover(); r != nil {
					t.Logf("Recovered from panic while opening KVStore for %s: %v", storeName, r)
				}
			}()
			nodeAppKvStoreService := runtime.NewKVStoreService(nodeApp.GetKey(storeName))
			nodeAppKvStore = nodeAppKvStoreService.OpenKVStore(nodeAppCtx)
			newAppKvStoreService := runtime.NewKVStoreService(newApp.GetKey(storeName))
			newAppKvStore = newAppKvStoreService.OpenKVStore(newAppCtx)
		}()

		if nodeAppKvStore == nil {
			if newAppKvStore == nil {
				// means that the module does not have a KVStore in both apps, so we skip the comparison
				continue
			} else {
				// means that the module has a KVStore in the new app, but not in the node app
				mismatches = append(mismatches, fmt.Sprintf("KVStore %s not found in node app", storeName))
				continue
			}
		}

		err := compareKVStores(nodeAppKvStore, newAppKvStore, ignorePrefixes)
		if err != nil {
			mismatches = append(mismatches, fmt.Sprintf("failed to compare %s KV stores: %v", storeName, err))
		}
	}

	if len(mismatches) > 0 {
		return fmt.Errorf("KVStore mismatches:\n%s", strings.Join(mismatches, "\n"))
	}

	return nil
}
