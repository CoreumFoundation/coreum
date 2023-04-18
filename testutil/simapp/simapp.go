// Package simapp contains utils to bootstrap the chain.
package simapp

import (
	"math/rand"
	"time"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
)

// Option represents simapp customisations.
type Option func() dbm.DB

// WithCustomDB returns the simapp Option to run with different DB.
func WithCustomDB(db dbm.DB) Option {
	return func() dbm.DB {
		return db
	}
}

// App is a simulation app wrapper.
type App struct {
	app.App
}

// New creates application instance with in-memory database and disabled logging.
func New(options ...Option) *App {
	var db dbm.DB

	db = dbm.NewMemDB()
	logger := log.NewNopLogger()

	for _, option := range options {
		customDB := option()
		if customDB != nil {
			db = customDB
		}
	}

	network, err := config.NetworkByChainID(constant.ChainIDDev)
	if err != nil {
		panic(err)
	}

	app.ChosenNetwork = network
	encoding := config.NewEncodingConfig(app.ModuleBasics)

	coreApp := app.New(logger, db, nil, true, simtestutil.EmptyAppOptions{})

	genesisState := app.ModuleBasics.DefaultGenesis(encoding.Codec)
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(errors.Errorf("can't Marshal genesisState: %v", err))
	}
	coreApp.InitChain(abci.RequestInitChain{
		ConsensusParams: simtestutil.DefaultConsensusParams,
		AppStateBytes:   stateBytes,
	})

	simApp := &App{*coreApp}

	return simApp
}

// BeginNextBlock begins new SimApp block and returns the ctx of the new block.
func (s *App) BeginNextBlock() sdk.Context {
	header := tmproto.Header{Height: s.LastBlockHeight() + 1}
	s.BeginBlock(abci.RequestBeginBlock{Header: header})
	return s.BaseApp.NewContext(false, header)
}

// EndBlockAndCommit ends the current block and commit the state.
func (s *App) EndBlockAndCommit(ctx sdk.Context) {
	s.EndBlocker(ctx, abci.RequestEndBlock{Height: ctx.BlockHeight()})
	s.Commit()
}

// GenAccount creates a new account and registers it in the App.
func (s *App) GenAccount(ctx sdk.Context) (sdk.AccAddress, *secp256k1.PrivKey) {
	privateKey := secp256k1.GenPrivKey()
	accountAddress := sdk.AccAddress(privateKey.PubKey().Address())
	account := s.AccountKeeper.NewAccount(ctx, &authtypes.BaseAccount{
		Address: accountAddress.String(),
	})
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
	signerAddress := sdk.AccAddress(priv.PubKey().Address())
	account := s.AccountKeeper.GetAccount(ctx, signerAddress)
	if account == nil {
		return sdk.GasInfo{}, nil, errors.Errorf("the account %s doesn't exist, check that it's created or state committed", signerAddress)
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
		"",
		[]uint64{accountNum},
		[]uint64{accountSeq},
		priv,
	)
	if err != nil {
		return sdk.GasInfo{}, nil, err
	}

	return s.SimDeliver(txCfg.TxEncoder(), tx)
}
