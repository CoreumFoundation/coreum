// Package simapp contains utils to bootstrap the chain.
package simapp

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/log"
	pruningtypes "cosmossdk.io/store/pruning/types"
	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	cmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/proto/tendermint/crypto"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"

	"github.com/CoreumFoundation/coreum/v6/app"
	"github.com/CoreumFoundation/coreum/v6/pkg/config"
	"github.com/CoreumFoundation/coreum/v6/pkg/config/constant"
)

const appHash = "sim-app-hash"

// Settings for the simapp initialization.
type Settings struct {
	db     dbm.DB
	logger log.Logger
}

var sdkConfigOnce = &sync.Once{}

// Option represents simapp customisations.
type Option func(settings Settings) Settings

// WithCustomDB returns the simapp Option to run with different DB.
func WithCustomDB(db dbm.DB) Option {
	return func(s Settings) Settings {
		s.db = db
		return s
	}
}

// WithCustomLogger returns the simapp Option to run with different logger.
func WithCustomLogger(logger log.Logger) Option {
	return func(s Settings) Settings {
		s.logger = logger
		return s
	}
}

// App is a simulation app wrapper.
type App struct {
	app.App
}

// New creates application instance with in-memory database and disabled logging.
func New(options ...Option) *App {
	settings := Settings{
		db:     dbm.NewMemDB(),
		logger: log.NewNopLogger(),
	}

	for _, option := range options {
		settings = option(settings)
	}

	sdkConfigOnce.Do(func() {
		network, err := config.NetworkConfigByChainID(constant.ChainIDDev)
		if err != nil {
			panic(err)
		}

		app.ChosenNetwork = network
		network.SetSDKConfig()
	})

	coreApp := app.New(settings.logger, settings.db, nil, true, simtestutil.NewAppOptionsWithFlagHome(tempDir()))
	pubKey, err := cryptocodec.ToCmtPubKeyInterface(ed25519.GenPrivKey().PubKey())
	if err != nil {
		panic(fmt.Sprintf("can't generate validator pub key genesisState: %v", err))
	}
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	senderPrivateKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivateKey.PubKey().Address().Bytes(), senderPrivateKey.PubKey(), 0, 0)

	defaultGenesis := coreApp.DefaultGenesis()
	genesisState, err := simtestutil.GenesisStateWithValSet(
		coreApp.AppCodec(),
		defaultGenesis,
		valSet,
		[]authtypes.GenesisAccount{acc},
	)
	if err != nil {
		panic(fmt.Sprintf("can't generate genesis state with wallet, err: %s", err))
	}

	stateBytes, err := cmjson.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(errors.Errorf("can't Marshal genesisState: %s", err))
	}

	_, err = coreApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: simtestutil.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
		Time:            time.Now(),
	})
	if err != nil {
		panic(errors.Errorf("can't init chain: %s", err))
	}

	simApp := &App{*coreApp}

	return simApp
}

func NewWithGenesis(genesisBytes []byte, options ...Option) (App, string, map[string]json.RawMessage, *abci.RequestInitChain, *abci.ResponseInitChain) {
	homeDir := tempDir()

	settings := Settings{
		db:     dbm.NewMemDB(),
		logger: log.NewNopLogger(),
	}

	for _, option := range options {
		settings = option(settings)
	}

	initChainReq, appState, err := convertExportedGenesisToInitChain(genesisBytes)
	if err != nil {
		panic(errors.Errorf("can't convert genesis bytes to init chain: %s", err))
	}

	coreApp := app.New(
		settings.logger,
		settings.db,
		nil,
		true,
		simtestutil.NewAppOptionsWithFlagHome(homeDir),
		baseapp.SetChainID(initChainReq.ChainId),
		baseapp.SetPruning(pruningtypes.NewPruningOptions(pruningtypes.PruningNothing)),
	)

	initChainRes, err := coreApp.InitChain(initChainReq)
	if err != nil {
		panic(errors.Errorf("can't init chain: %s", err))
	}

	return App{*coreApp}, homeDir, appState, initChainReq, initChainRes
}

// BeginNextBlock begins new SimApp block and returns the ctx of the new block.
func (s *App) BeginNextBlock() (sdk.Context, sdk.BeginBlock, error) {
	header := tmproto.Header{
		Height:  s.LastBlockHeight() + 1,
		Time:    time.Now(),
		AppHash: []byte(appHash),
	}
	ctx := s.NewContextLegacy(false, header)
	beginBlock, err := s.BeginBlocker(ctx)
	return ctx, beginBlock, err
}

// BeginNextBlockAtTime begins new SimApp block and returns the ctx of the new block with given time.
func (s *App) BeginNextBlockAtTime(blockTime time.Time) (sdk.Context, sdk.BeginBlock, error) {
	header := tmproto.Header{
		Height:  s.LastBlockHeight() + 1,
		Time:    blockTime,
		AppHash: []byte(appHash),
	}
	ctx := s.NewContextLegacy(false, header)
	beginBlock, err := s.BeginBlocker(ctx)
	return ctx, beginBlock, err
}

// BeginNextBlockAtHeight begins new SimApp block and returns the ctx of the new block with given hight.
func (s *App) BeginNextBlockAtHeight(height int64) (sdk.Context, sdk.BeginBlock, error) {
	header := tmproto.Header{
		Height:  height,
		Time:    time.Now(),
		AppHash: []byte(appHash),
	}
	ctx := s.NewContextLegacy(false, header)
	beginBlock, err := s.BeginBlocker(ctx)
	return ctx, beginBlock, err
}

// FinalizeBlock ends the current block and commit the state and creates a new block.
func (s *App) FinalizeBlock() error {
	_, err := s.App.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: s.LastBlockHeight() + 1,
		Hash:   s.LastCommitID().Hash,
		Time:   time.Now(),
	})
	return err
}

// FinalizeBlockAtTime ends the current block and commit the state and creates a new block at specified time.
func (s *App) FinalizeBlockAtTime(blockTime time.Time) error {
	_, err := s.App.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: s.LastBlockHeight() + 1,
		Hash:   s.LastCommitID().Hash,
		Time:   blockTime,
	})
	return err
}

// GenAccount creates a new account and registers it in the App.
func (s *App) GenAccount(ctx sdk.Context) (sdk.AccAddress, *secp256k1.PrivKey) {
	privateKey := secp256k1.GenPrivKey()
	accountAddress := sdk.AccAddress(privateKey.PubKey().Address())
	account := s.AccountKeeper.NewAccountWithAddress(ctx, accountAddress)
	s.AccountKeeper.SetAccount(ctx, account)

	return accountAddress, privateKey
}

// FundAccount mints and sends the coins to the provided App account.
func (s *App) FundAccount(ctx sdk.Context, address sdk.AccAddress, balances sdk.Coins) error {
	if err := s.BankKeeper.MintCoins(ctx, minttypes.ModuleName, balances); err != nil {
		return errors.Wrap(err, "can't mint in simapp")
	}

	if err := s.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, address, balances); err != nil {
		return errors.Wrap(err, "can't send funding coins in simapp")
	}

	return nil
}

// SendTx sends the tx to the simApp.
func (s *App) SendTx(
	ctx sdk.Context,
	feeAmt sdk.Coin,
	gas uint64,
	priv cryptotypes.PrivKey,
	messages ...sdk.Msg,
) (sdk.GasInfo, *sdk.Result, error) {
	tx, err := s.GenTx(
		ctx, feeAmt, gas, priv, messages...,
	)
	if err != nil {
		return sdk.GasInfo{}, nil, err
	}

	txCfg := s.TxConfig()
	return s.SimDeliver(txCfg.TxEncoder(), tx)
}

// GenTx generates a tx from messages.
func (s *App) GenTx(
	ctx sdk.Context,
	feeAmt sdk.Coin,
	gas uint64,
	priv cryptotypes.PrivKey,
	messages ...sdk.Msg,
) (sdk.Tx, error) {
	signerAddress := sdk.AccAddress(priv.PubKey().Address())
	account := s.AccountKeeper.GetAccount(ctx, signerAddress)
	if account == nil {
		return nil, errors.Errorf(
			"the account %s doesn't exist, check that it's created or state committed",
			signerAddress,
		)
	}
	accountNum := account.GetAccountNumber()
	accountSeq := account.GetSequence()

	txCfg := s.TxConfig()

	tx, err := simtestutil.GenSignedMockTx(
		rand.New(rand.NewSource(time.Now().UnixNano())),
		txCfg,
		messages,
		sdk.NewCoins(feeAmt),
		gas,
		s.ChainID(),
		[]uint64{accountNum},
		[]uint64{accountSeq},
		priv,
	)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

// MintAndSendCoin mints coin to the mint module and sends them to the recipient.
func (s *App) MintAndSendCoin(
	t *testing.T,
	sdkCtx sdk.Context,
	recipient sdk.AccAddress,
	coins sdk.Coins,
) {
	require.NoError(
		t, s.BankKeeper.MintCoins(sdkCtx, minttypes.ModuleName, coins),
	)
	require.NoError(
		t, s.BankKeeper.SendCoinsFromModuleToAccount(sdkCtx, minttypes.ModuleName, recipient, coins),
	)
}

func tempDir() string {
	dir, err := os.MkdirTemp("", "cored")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(dir) //nolint:errcheck // we don't care

	return dir
}

// CopyContextWithMultiStore returns a sdk.Context with a copied MultiStore.
func CopyContextWithMultiStore(sdkCtx sdk.Context) sdk.Context {
	return sdkCtx.WithMultiStore(sdkCtx.MultiStore().CacheWrap().(storetypes.MultiStore))
}

func convertExportedGenesisToInitChain(jsonBytes []byte) (*abci.RequestInitChain, map[string]json.RawMessage, error) {
	var export struct {
		InitialHeight int64                      `json:"initial_height"`
		GenesisTime   string                     `json:"genesis_time"`
		ChainID       string                     `json:"chain_id"`
		AppState      map[string]json.RawMessage `json:"app_state"`
		Consensus     struct {
			Params struct {
				Block struct {
					MaxBytes string `json:"max_bytes"`
					MaxGas   string `json:"max_gas"`
				} `json:"block"`
				Evidence struct {
					MaxAgeNumBlocks string `json:"max_age_num_blocks"`
					MaxAgeDuration  string `json:"max_age_duration"`
					MaxBytes        string `json:"max_bytes"`
				} `json:"evidence"`
				Validator struct {
					PubKeyTypes []string `json:"pub_key_types"`
				} `json:"validator"`
				Version struct {
					App string `json:"app"`
				} `json:"version"`
				ABCI struct {
					VoteExtensionsEnableHeight string `json:"vote_extensions_enable_height"`
				} `json:"abci"`
			} `json:"params"`
			Validators []struct {
				Address string `json:"address"`
				PubKey  struct {
					Type  string `json:"type"`
					Value string `json:"value"`
				} `json:"pub_key"`
				Power string `json:"power"`
				Name  string `json:"name"`
			} `json:"validators"`
		} `json:"consensus"`
	}
	if err := json.Unmarshal(jsonBytes, &export); err != nil {
		return nil, nil, err
	}

	// Marshal app_state to bytes
	appStateBytes, err := json.Marshal(export.AppState)
	if err != nil {
		return nil, nil, err
	}

	// Parse genesis_time
	genesisTime, err := time.Parse(time.RFC3339Nano, export.GenesisTime)
	if err != nil {
		return nil, nil, err
	}

	// Build ConsensusParams
	consensusParams := &tmproto.ConsensusParams{
		Block: &tmproto.BlockParams{
			MaxBytes: mustParseInt64(export.Consensus.Params.Block.MaxBytes),
			MaxGas:   mustParseInt64(export.Consensus.Params.Block.MaxGas),
		},
		Evidence: &tmproto.EvidenceParams{
			MaxAgeNumBlocks: mustParseInt64(export.Consensus.Params.Evidence.MaxAgeNumBlocks),
			MaxAgeDuration:  mustParseDuration(export.Consensus.Params.Evidence.MaxAgeDuration),
			MaxBytes:        mustParseInt64(export.Consensus.Params.Evidence.MaxBytes),
		},
		Validator: &tmproto.ValidatorParams{
			PubKeyTypes: export.Consensus.Params.Validator.PubKeyTypes,
		},
		Version: &tmproto.VersionParams{
			App: mustParseUint64(export.Consensus.Params.Version.App),
		},
		Abci: &tmproto.ABCIParams{
			VoteExtensionsEnableHeight: mustParseInt64(export.Consensus.Params.ABCI.VoteExtensionsEnableHeight),
		},
	}

	// Build Validators
	var validators []abci.ValidatorUpdate
	for _, v := range export.Consensus.Validators {
		pubKey, err := base64.StdEncoding.DecodeString(v.PubKey.Value)
		if err != nil {
			return nil, nil, err
		}
		validators = append(validators, abci.ValidatorUpdate{
			PubKey: crypto.PublicKey{
				Sum: &crypto.PublicKey_Ed25519{Ed25519: pubKey},
			},
			Power: mustParseInt64(v.Power),
		})
	}

	return &abci.RequestInitChain{
		Time:            genesisTime,
		ChainId:         export.ChainID,
		ConsensusParams: consensusParams,
		Validators:      validators,
		AppStateBytes:   appStateBytes,
		InitialHeight:   export.InitialHeight,
	}, export.AppState, nil
}

// Helper functions
func mustParseInt64(s string) int64 {
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}
func mustParseUint64(s string) uint64 {
	v, _ := strconv.ParseUint(s, 10, 64)
	return v
}
func mustParseDuration(s string) time.Duration {
	v, _ := time.ParseDuration(s + "ns")
	return v
}
