//go:build integrationtests

package export_test

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"

	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	sdkstore "cosmossdk.io/core/store"
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

// TestExportGenesisModuleHashes tests the export of genesis and compares the module hashes
// steps:
// 1. Export genesis and application state from a full node.
// 2. Initialize a new app with the exported genesis.
// 3. Move both apps to the same height by finalizing a block.
// 4. Compare the module hashes of both apps to ensure they match.
func TestExportGenesisModuleHashes(t *testing.T) {
	requireT := require.New(t)

	chainId := string(constant.ChainIDDev)

	// the chain is stopped and the genesis is exported from a full node
	exportedApp, exportedGenesisBuf := exportGenesisAndApp(t, requireT, chainId)

	// the exported genesis is used to initialize a new app to simulate new chain initialization
	initiatedApp, _, _, initChainReq, _ := simapp.NewWithGenesis(exportedGenesisBuf.Bytes())

	// sync heights of both apps stores
	syncAppsHeights(t, requireT, exportedApp, &initiatedApp.App, initChainReq)

	// check that the module hashes of both apps match
	checkModuleStoreMismatches(t, requireT, exportedApp, &initiatedApp.App, initChainReq.InitialHeight)
}

// exportGenesisAndApp exports the genesis and application state from a full node
// and returns the application instance and the exported genesis as a bytes buffer.
func exportGenesisAndApp(t *testing.T, requireT *require.Assertions, chainId string) (*app.App, bytes.Buffer) {
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

	// the home directory of the node is being copied to prevent resource unavailability
	// because of internal locks on the database files
	copiedNodeHome := filepath.Join(znetAppDir, fmt.Sprintf("%d-cored-05-full", currentTs))
	// only copy if the directory does not exist
	_, statErr := os.Stat(copiedNodeHome)
	if os.IsNotExist(statErr) {
		err := copyDir(fullNodeHome, copiedNodeHome)
		requireT.NoError(err, "failed to copy data directory")
	} else {
		requireT.NoError(statErr, "failed to stat copied data directory")
		t.Logf("Copied node home already exists, skipping copy: %s", copiedNodeHome)
	}
	copiedNodeChainHome := filepath.Join(copiedNodeHome, chainId)
	copiedDBDir := filepath.Join(copiedNodeChainHome, "data")
	nodeDb, err := dbm.NewDB("application", dbm.GoLevelDBBackend, copiedDBDir)
	requireT.NoError(err, "failed to open node DB at %s", copiedDBDir)
	t.Log("Copied node DB directory:", copiedDBDir)

	// initialize the application with the copied home directory

	network, err := config.NetworkConfigByChainID(constant.ChainID(chainId))
	if err != nil {
		panic(errors.WithStack(err))
	}
	app.ChosenNetwork = network
	network.SetSDKConfig()

	// this is a temporary app equivalent to the actual running chain exported app
	chainFullNodeApp := app.New(logger, nodeDb, nil, false, simtestutil.AppOptionsMap{
		flags.FlagHome:            copiedNodeChainHome,
		server.FlagInvCheckPeriod: time.Millisecond * 100,
	})

	fullNodeBin := filepath.Join(coredBinDir, "cored")
	exportJSONPath := filepath.Join(znetAppDir, fmt.Sprintf("%d-exported-genesis.json", currentTs))
	var exportBuf bytes.Buffer

	// Filter out ignored modules
	var modulesToExport []string
	for _, m := range chainFullNodeApp.ModuleManager.ModuleNames() {
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

	return chainFullNodeApp, exportBuf
}

func syncAppsHeights(t *testing.T, requireT *require.Assertions, exportedApp *app.App, initiatedApp *app.App, initChainReq *abci.RequestInitChain) {
	// load the latest version from the exported app
	// the initial height is the height that need to gets finalized in the initiated app
	nodeAppStateHeight := initChainReq.InitialHeight - 1
	err := exportedApp.LoadVersion(nodeAppStateHeight)
	requireT.NoError(err, "failed to load version %d from exported app", nodeAppStateHeight)

	// finalize new block for the exported app
	_, err = exportedApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: initChainReq.InitialHeight,
	})
	require.NoError(t, err)
	_, err = exportedApp.Commit()
	require.NoError(t, err)

	// finalize new block for the initiated app
	_, err = initiatedApp.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: initChainReq.InitialHeight,
	})
	require.NoError(t, err)

	_, err = initiatedApp.Commit()
	require.NoError(t, err)
}

func checkModuleStoreMismatches(t *testing.T, requireT *require.Assertions, nodeApp *app.App, newApp *app.App, height int64) {
	var mismatches []string

	// ensure the app contexts are created for the specified height
	nodeAppCtx := nodeApp.NewUncachedContext(false, cmtproto.Header{Height: height})
	newAppCtx := newApp.NewUncachedContext(false, cmtproto.Header{Height: height})

	for _, moduleName := range nodeApp.ModuleManager.ModuleNames() {
		// skip the module if it is in the ignored list
		if _, ok := ignoredModuleNames[moduleName]; ok {
			continue
		}

		// auth module store name is different from its module name
		// so we use a special case for it
		storeName := moduleName
		if moduleName == authtypes.StoreKey {
			storeName = "acc"
		}

		// list the prefixes to ignore for the module
		var modulePrefixesToIgnore [][]byte
		if prefixes, ok := ignoredPrefixes[moduleName]; ok {
			modulePrefixesToIgnore = prefixes
		}

		var nodeAppKvStore, newAppKvStore sdkstore.KVStore
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
			if newAppKvStore != nil {
				// means that the module has a KVStore in the initiated app, but not in the exported app
				mismatches = append(mismatches, fmt.Sprintf("KVStore %s not found in exported app", storeName))
			}
			// means that the module does not have a KVStore in both apps, so we skip the comparison
			continue
		}

		// compare the KVStores of the exported app and the initiated app
		// and append any mismatches to the list
		err := compareKVStores(nodeAppKvStore, newAppKvStore, modulePrefixesToIgnore)
		if err != nil {
			mismatches = append(mismatches, fmt.Sprintf("failed to compare %s KV stores: %v", storeName, err))
		}
	}

	requireT.Equal(0, len(mismatches), "KVStore mismatches:\n%s", strings.Join(mismatches, "\n"))
}

func compareKVStores(exportedAppStore, initiatedAppStore sdkstore.KVStore, ignorePrefixes [][]byte) error {
	// build maps of key-value pairs for exported app store
	exportedMap := make(map[string][]byte)
	iter1, err := exportedAppStore.Iterator(nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create iterator for exported app store: %w", err)
	}
	defer iter1.Close()
	for ; iter1.Valid(); iter1.Next() {
		exportedMap[string(iter1.Key())] = append([]byte(nil), iter1.Value()...)
	}

	// build maps of key-value pairs for initiated app store
	initiatedMap := make(map[string][]byte)
	iter2, err := initiatedAppStore.Iterator(nil, nil)
	if err != nil {
		return fmt.Errorf("failed to create iterator for initiated app store: %w", err)
	}
	defer iter2.Close()
	for ; iter2.Valid(); iter2.Next() {
		initiatedMap[string(iter2.Key())] = append([]byte(nil), iter2.Value()...)
	}

	var mismatches []string

	// Check for keys in exportedMap not in initiatedMap or with different values
	for k, v := range exportedMap {
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

		// check if the key  exists in the initiated app store
		// if the key is not found, append the mismatch to the list
		nv, ok := initiatedMap[k]
		if !ok {
			mismatches = append(mismatches, fmt.Sprintf("key %q missing in initiated app store", k))
			continue
		}
		// check if the value matches
		// if the value is not equal, append the mismatch to the list
		if !bytes.Equal(v, nv) {
			mismatches = append(mismatches, fmt.Sprintf("value mismatch for key %q: %q vs %q", k, v, nv))
		}
	}

	// Check for extra keys in initiated app store
	for k := range initiatedMap {
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

		// check if the key exists in the exported app store
		// if the key is not found, append the mismatch to the list
		if _, ok := exportedMap[k]; !ok {
			mismatches = append(mismatches, fmt.Sprintf("extra key %X in new store", []byte(k)))
		}
	}

	// If there are any mismatches, return an error with the details
	if len(mismatches) > 0 {
		return fmt.Errorf("KVStore mismatches:\n%s", strings.Join(mismatches, "\n"))
	}
	return nil
}

// copyDir recursively copies a directory tree from src to dst
func copyDir(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Recursively copy subdirectory
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// copyFile copies a single file from src to dst
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func executeCommand(ctx context.Context, buf *bytes.Buffer, name string, args []string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
