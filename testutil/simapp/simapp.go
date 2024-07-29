// Package simapp contains utils to bootstrap the chain.
package simapp

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"cosmossdk.io/log"
	sdkmath "cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/json"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
	dbm "github.com/cosmos/cosmos-db"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/app"
	"github.com/CoreumFoundation/coreum/v4/pkg/config"
	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
)

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
	pubKey, err := cryptocodec.ToTmPubKeyInterface(ed25519.GenPrivKey().PubKey())
	if err != nil {
		panic(fmt.Sprintf("can't generate validator pub key genesisState: %v", err))
	}
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})
	senderPrivateKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivateKey.PubKey().Address().Bytes(), senderPrivateKey.PubKey(), 0, 0)

	defaultGenesis := app.ModuleBasics.DefaultGenesis(coreApp.AppCodec())
	genesisState, err := simtestutil.GenesisStateWithValSet(
		coreApp.AppCodec(),
		defaultGenesis,
		valSet,
		[]authtypes.GenesisAccount{acc},
	)
	if err != nil {
		panic(fmt.Sprintf("can't generate genesis state with wallet, err: %s", err))
	}

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(errors.Errorf("can't Marshal genesisState: %s", err))
	}

	_, err = coreApp.InitChain(&abci.RequestInitChain{
		ConsensusParams: simtestutil.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})
	if err != nil {
		panic(errors.Errorf("can't init chain: %s", err))
	}

	simApp := &App{*coreApp}

	return simApp
}

// BeginNextBlock begins new SimApp block and returns the ctx of the new block.
func (s *App) BeginNextBlock(blockTime time.Time) (sdk.Context, sdk.BeginBlock, error) {
	if blockTime.IsZero() {
		blockTime = time.Now()
	}
	header := tmproto.Header{Height: s.App.LastBlockHeight() + 1, Time: blockTime}
	ctx := s.NewContextLegacy(false, header)
	beginBlock, err := s.App.BeginBlocker(ctx)
	return ctx, beginBlock, err
}

// FinalizeBlock ends the current block and commit the state and creates a new block.
func (s *App) FinalizeBlock() error {
	_, err := s.App.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: s.LastBlockHeight() + 1,
		Hash:   s.LastCommitID().Hash,
	})
	return err
}

// GenAccount creates a new account and registers it in the App.
func (s *App) GenAccount(ctx sdk.Context) (sdk.AccAddress, *secp256k1.PrivKey) {
	privateKey := secp256k1.GenPrivKey()
	accountAddress := sdk.AccAddress(privateKey.PubKey().Address())
	account := s.App.AccountKeeper.NewAccountWithAddress(ctx, accountAddress)
	s.App.AccountKeeper.SetAccount(ctx, account)

	return accountAddress, privateKey
}

// FundAccount mints and sends the coins to the provided App account.
func (s *App) FundAccount(ctx sdk.Context, address sdk.AccAddress, balances sdk.Coins) error {
	if err := s.App.BankKeeper.MintCoins(ctx, minttypes.ModuleName, balances); err != nil {
		return errors.Wrap(err, "can't mint in simapp")
	}

	if err := s.App.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, address, balances); err != nil {
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

	txCfg := config.NewEncodingConfig(app.ModuleBasics).TxConfig
	return s.App.SimDeliver(txCfg.TxEncoder(), tx)
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
	account := s.App.AccountKeeper.GetAccount(ctx, signerAddress)
	if account == nil {
		return nil, errors.Errorf(
			"the account %s doesn't exist, check that it's created or state committed",
			signerAddress,
		)
	}
	accountNum := account.GetAccountNumber()
	accountSeq := account.GetSequence()

	txCfg := config.NewEncodingConfig(app.ModuleBasics).TxConfig

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

// SimulateFundAndSendTx simulates the tx, funds account and sends the tx to the simApp.
func (s *App) SimulateFundAndSendTx(
	ctx sdk.Context,
	priv cryptotypes.PrivKey,
	messages ...sdk.Msg,
) (sdk.GasInfo, *sdk.Result, error) {
	simTx, err := s.GenTx(
		ctx,
		sdk.NewCoin(constant.DenomDev, sdkmath.ZeroInt()),
		0,
		priv,
		messages...,
	)
	if err != nil {
		return sdk.GasInfo{}, nil, err
	}
	txCfg := config.NewEncodingConfig(app.ModuleBasics).TxConfig
	txBytes, err := txCfg.TxEncoder()(simTx)
	if err != nil {
		return sdk.GasInfo{}, nil, err
	}

	simGas, _, err := s.Simulate(txBytes)
	if err != nil {
		return sdk.GasInfo{}, nil, err
	}
	targetGas := sdkmath.NewInt(int64(simGas.GasUsed * 2))
	minGasPrice := s.App.FeeModelKeeper.GetMinGasPrice(ctx)
	fee := sdk.NewCoin(minGasPrice.Denom, minGasPrice.Amount.MulInt(targetGas).MulInt64(2).RoundInt())

	accountAddress := sdk.AccAddress(priv.PubKey().Address())
	err = s.FundAccount(ctx, accountAddress, sdk.NewCoins(fee))
	if err != nil {
		return sdk.GasInfo{}, nil, err
	}

	return s.SendTx(
		ctx,
		fee,
		targetGas.Uint64(),
		priv,
		messages...,
	)
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
	defer os.RemoveAll(dir)

	return dir
}
