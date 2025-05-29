//go:build integrationtests

package export_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"cosmossdk.io/log"
	"cosmossdk.io/store/iavl"
	"cosmossdk.io/store/rootmulti"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/store/wrapper"
	"github.com/CoreumFoundation/coreum/v6/app"
	integrationtests "github.com/CoreumFoundation/coreum/v6/integration-tests"
	"github.com/CoreumFoundation/coreum/v6/testutil/simapp"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	cosmosiavl "github.com/cosmos/iavl"
	"github.com/stretchr/testify/require"
)

func executeCommand(ctx context.Context, buf *bytes.Buffer, name string, args []string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func storeJSONToFile(requireT *require.Assertions, filename string, buf []byte) {
	outFileReal, err := os.Create(filename)
	requireT.NoError(err, "failed to create temp file for realAppExported")
	defer outFileReal.Close()
	_, err = outFileReal.Write(buf)
	requireT.NoError(err, "failed to write buf to realAppExported file")
}

func openDB(rootDir string, backendType dbm.BackendType) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return dbm.NewDB("application", backendType, dataDir)
}

// AppOptionsMap is a stub implementing AppOptions which can get data from a map
type AppOptionsMap map[string]interface{}

func (m AppOptionsMap) Get(key string) interface{} {
	v, ok := m[key]
	if !ok {
		return interface{}(nil)
	}

	return v
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

func TestXxx(t *testing.T) {
	refCtx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	// currentHeight := infoRes.Block.Header.Height
	// t.Logf("App hash from infoRes: %d", appHashFromInfo)
	homeDir, err := os.UserHomeDir()
	requireT.NoError(err, "failed to get user home directory")
	t.Logf("Home directory: %s", homeDir)

	// validatorHome := filepath.Join(homeDir, ".crust", "znet", "znet", "app", "cored-00-val")
	validatorHome := os.Getenv("CORED_VAL_HOME")
	t.Logf("Validator home directory: %s", validatorHome)

	fullHome := os.Getenv("CORED_FULL_HOME")
	t.Logf("Full home directory: %s", fullHome)

	logBuffer := new(bytes.Buffer)
	logger := log.NewLogger(logBuffer, log.ColorOption(false))

	var exportBuf bytes.Buffer
	err = executeCommand(
		refCtx,
		&exportBuf,
		filepath.Join(os.Getenv("CORED_BIN_PATH"), "cored"),
		[]string{"export", "--chain-id", chain.ClientContext.ChainID(), "--home", validatorHome},
	)
	requireT.NoError(err)

	storeJSONToFile(requireT, "realappexported.json", exportBuf.Bytes())
	t.Logf("realAppExported written to: %s", "realappexported.json")

	newApp, _, newAppState, initChain, initChainResponse, newAppDb := simapp.NewWithGenesis(exportBuf.Bytes())

	newAppctx := newApp.NewContext(false)
	customStakingParams, err := newApp.CustomParamsKeeper.GetStakingParams(newAppctx)
	requireT.NoError(err, "failed to get staking params from new app")
	requireT.Equal(customStakingParams.MinSelfDelegation.Int64(), int64(10000000), "staking params should not be empty in new app")

	validators, err := newApp.StakingKeeper.GetAllValidators(newAppctx)
	requireT.NoError(err, "failed to get all validators from new app")
	requireT.NotEmpty(validators, "validators should not be empty in new app")

	// ctxNew := newApp.NewContextLegacy(false, tmproto.Header{})

	// header := tmproto.Header{
	// 	Height:  newApp.LastBlockHeight() + 1,
	// 	Time:    time.Now(),
	// 	AppHash: []byte(newApp.LastCommitID().Hash),
	// }
	// ctx := newApp.NewContextLegacy(false, header)

	// requireT.NoError(err, "failed to begin next block in new app")
	// newApp.EndBlocker(ctx)

	newAppStateBytes, err := json.Marshal(newAppState)
	requireT.NoError(err, "failed to marshal newAppState to JSON")
	storeJSONToFile(requireT, "newappstate.json", newAppStateBytes)
	t.Logf("newappstate written to: %s", "newappstate.json")

	srcDir := filepath.Join(fullHome, initChain.ChainId)
	dstDir := filepath.Join(".", initChain.ChainId)
	err = os.RemoveAll(dstDir)
	requireT.NoError(err, "failed to remove existing destination directory")
	err = exec.Command("cp", "-a", srcDir, dstDir).Run()
	requireT.NoError(err, fmt.Sprintf("failed to copy directory from %s to %s", srcDir, dstDir))
	t.Logf("Copied directory from %s to %s", srcDir, dstDir)

	db, err := openDB(dstDir, dbm.GoLevelDBBackend)
	requireT.NoError(err, "failed to open DB at %s", dstDir)
	appOptions := AppOptionsMap{}
	appOptions["home"] = dstDir // Use a temporary directory for the new app
	appOptions[server.FlagInvCheckPeriod] = time.Millisecond * 100
	commitInfo, err := getModuleHashesAtHeight(app.New(logger, db, nil, false, appOptions), initChain.InitialHeight)
	requireT.NoError(err, "failed to get module hashes at height")
	t.Logf("CommitInfo for height %d: %+v", initChain.InitialHeight, commitInfo)

	// err = newApp.FinalizeBlock()
	_, err = newApp.Commit()
	// newApp.EndBlocker(newAppctx)
	requireT.NoError(err, "failed to commit new app after initialization")

	newAppCmt := newApp.CommitMultiStore().(*rootmulti.Store)
	keysByName := newAppCmt.StoreKeysByName()
	err = ViewKeysInStore(
		initChain.InitialHeight,
		newAppDb,
		&newApp.App,
		keysByName["gov"],
	)
	if err != nil {
		panic(err)
	}

	commitInfoNew, err := getModuleHashesAtHeight(&newApp.App, initChain.InitialHeight)
	requireT.NoError(err, "failed to get module hashes at height")

	t.Logf("New app last commit ID: %X", newApp.LastCommitID().Hash)
	t.Logf("New app last commit ID: %X", commitInfo.Hash())

	// requireT.NoError(err, "failed to commit new app after initialization")
	// newApp.FinalizeBlock()

	t.Logf("New app last commit ID: %X", newApp.LastCommitID().Hash)

	t.Logf("CommitInfo for height %d: %+v", initChain.InitialHeight, commitInfoNew)

	// requireT.Equal(commitInfo, commitInfoNew, "CommitInfo from original and new app should be equal")

	// newAppExported, err := newApp.ExportAppStateAndValidators(false, nil, nil)
	// outFile, err := os.Create("newappexported.json")
	// requireT.NoError(err, "failed to create temp file for newAppExported")
	// defer outFile.Close()
	// encoder := json.NewEncoder(outFile)
	// encoder.SetIndent("", "  ")
	// err = encoder.Encode(newAppExported)
	// requireT.NoError(err, "failed to encode newAppExported to JSON")
	// t.Logf("newAppExported written to: %s", outFile.Name())

	// requireT.NoError(err, "failed to export app state and validators")

	// realExported, err := ConvertJSONToExportedApp(buf.Bytes())
	// requireT.NoError(err, "failed to convert JSON to ExportedApp")
	// requireT.Equal(newAppExported.AppState, realExported.AppState, "Exported height should be equal to initial height + 1")

	t.Logf("New app block header initial height: %d", initChain.InitialHeight)
	t.Logf("Init chain app hash: %X", initChainResponse.AppHash)
	t.Logf("New app CommitID hash: %X", newApp.LastCommitID().Hash)

	// t.Logf("New app block header app hash: %X", newApp.LastCommitID().Hash)

	time.Sleep(5 * time.Second) // Give some time for the chain to be ready

	tmcQueryClient := cmtservice.NewServiceClient(chain.ClientContext)
	currBlock, err := tmcQueryClient.GetLatestBlock(refCtx, &cmtservice.GetLatestBlockRequest{})
	requireT.NoError(err)
	t.Logf("Height from currBlock: %d", currBlock.SdkBlock.Header.Height)

	tmQueryClient := cmtservice.NewServiceClient(chain.ClientContext)
	refBlock, err := tmQueryClient.GetBlockByHeight(refCtx, &cmtservice.GetBlockByHeightRequest{Height: initChain.InitialHeight})
	requireT.NoError(err)
	t.Logf("Height from refBlock: %d", refBlock.SdkBlock.Header.Height)
	t.Logf("App hash from refBlock: %X", refBlock.SdkBlock.Header.AppHash)

	time.Sleep(5 * time.Second) // Give some time for the chain to be ready

	refNextBlock, err := tmQueryClient.GetBlockByHeight(refCtx, &cmtservice.GetBlockByHeightRequest{Height: initChain.InitialHeight + 1})
	requireT.NoError(err)
	t.Logf("Height from refNextBlock: %d", refNextBlock.SdkBlock.Header.Height)
	t.Logf("App hash from refNextBlock: %X", refNextBlock.SdkBlock.Header.AppHash)

	// latestBlockHeader, err := chain.LatestBlockHeader(ctx)

	// appHashFromInfo := latestBlockHeader.AppHash

	// newContext, _, err := newApp.BeginNextBlock()
	// requireT.NoError(err)
	// newApp.FinalizeBlock()
	// _, err = newApp.Commit()
	// requireT.NoError(err)

	// header := tmproto.Header{
	// 	Height:  newApp.LastBlockHeight() + 1,
	// 	Time:    refNextBlock.SdkBlock.Header.Time,
	// 	AppHash: newApp.LastCommitID().Hash,
	// }
	// beginBlockctx := newApp.NewContextLegacy(false, header)
	// _, err = newApp.BeginBlocker(beginBlockctx)
	// requireT.NoError(err)
	// newApp.Commit()
	// newApp.FinalizeBlock()

	// t.Logf("New app block header app hash: %X", beginBlockctx.BlockHeader().AppHash)
	// t.Logf("New app last block height: %d", newApp.LastBlockHeight())
	// t.Logf("New app last commit ID: %X", newApp.LastCommitID().Hash)

	// newContext, _, err := newApp.BeginNextBlock()
	// requireT.NoError(err)
	// newApp.FinalizeBlock()
	// _, err = newApp.Commit()
	// requireT.NoError(err)

	// t.Logf("New app block header app hash: %X", newContext.BlockHeader().AppHash)
	// t.Logf("New app last block height: %d", newApp.LastBlockHeight())
	// t.Logf("New app last commit ID: %X", newApp.LastCommitID().Hash)
	// tmQueryClient = cmtservice.NewServiceClient(chain.ClientContext)
	// simappNextBlock, err := tmQueryClient.GetBlockByHeight(simappCtx, &cmtservice.GetBlockByHeightRequest{Height: initChain.InitialHeight + 1})
	// requireT.NoError(err)

	// commitInfo := newApp.CommitMultiStore().LastCommitID()
	// appHash := commitInfo.Hash
	// t.Logf("App hash: %X", appHash)
	// panic(appHash)
	// panic("TODO: implement export test")

	requireT.Equal(newApp.LastCommitID().Hash, refNextBlock.SdkBlock.Header.AppHash, "App hash from new app should match the app hash from the reference block %X != %X", newApp.LastCommitID().Hash, refNextBlock.SdkBlock.Header.AppHash)
	// requireT.Equal(appHashFromInfo, appHash)

}

// ConvertJSONToExportedApp reads a JSON file and converts it to ExportedApp
func ConvertJSONToExportedApp(bz []byte) (*servertypes.ExportedApp, error) {
	// Define a struct for the relevant fields
	var raw struct {
		AppState  json.RawMessage `json:"app_state"`
		Consensus struct {
			Validators []cmttypes.GenesisValidator `json:"validators"`
			Params     cmtproto.ConsensusParams    `json:"params"`
		} `json:"consensus"`
		InitialHeight json.RawMessage `json:"initial_height"`
	}

	if err := json.Unmarshal(bz, &raw); err != nil {
		return nil, err
	}

	// Parse height (can be string or int in JSON)
	var height int64
	if err := json.Unmarshal(raw.InitialHeight, &height); err != nil {
		var heightStr string
		if err := json.Unmarshal(raw.InitialHeight, &heightStr); err == nil {
			height, _ = strconv.ParseInt(heightStr, 10, 64)
		}
	}

	return &servertypes.ExportedApp{
		AppState:        raw.AppState,
		Validators:      raw.Consensus.Validators,
		Height:          height,
		ConsensusParams: raw.Consensus.Params,
	}, nil
}

func ViewKeysInStore(
	height int64,
	db dbm.DB,
	tempApp *app.App,
	key storetypes.StoreKey,
) error {
	cmt := tempApp.CommitMultiStore().(*rootmulti.Store)
	if cmt.GetCommitKVStore(key).GetStoreType() != storetypes.StoreTypeIAVL {
		return nil
	}

	err := cmt.LoadVersion(height)
	if err != nil {
		return err
	}

	keysByName := cmt.StoreKeysByName()

	// the 's/k:' prefix cretes subspace for the iavl tree to avoid key collisions.
	modulePrefix := "s/k:" + key.Name() + "/"
	moduleDB := dbm.NewPrefixDB(db, []byte(modulePrefix))

	tree := cosmosiavl.NewMutableTree(
		wrapper.NewDBWrapper(moduleDB),
		iavl.DefaultIAVLCacheSize,
		false,
		tempApp.Logger(),
		cosmosiavl.InitialVersionOption(0),
	)

	_, err = tree.LoadVersion(height)
	if err != nil {
		return err
	}

	h1 := tree.WorkingHash()
	fmt.Printf("tree working hash: %x \n", h1)
	fmt.Printf("cmt working hash: h1:%x \n", cmt.WorkingHash())
	kv1 := map[string][]byte{}
	_, err = tree.Iterate(func(key, value []byte) bool {
		kv1[string(key)] = value
		return false
	})
	if err != nil {
		return err
	}

	str1 := cmt.GetCommitKVStore(keysByName[key.Name()]).(*iavl.Store)

	iter := str1.Iterator(nil, nil)
	dkv := map[string][]byte{}
	for iter.Valid() {
		dkv[string(iter.Key())] = iter.Value()
		iter.Next()
	}

	// TODO: add sample code for comparing the values.
	// compare values on two maps
	for k, v := range kv1 {
		dv, found := dkv[k]
		if !found {
			return errors.New("could find key:" + k)
		}

		if bytes.Equal(v, dv) {
			return errors.New("values of key:" + k + "mismatch." + "v1:" + string(v) + ",v2:" + string(dv))
		}
	}

	return nil
}
