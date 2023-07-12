// Package simapp contains utils to bootstrap the chain.
package simapp

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/ibc-go/v4/testing/simapp/helpers"
	"github.com/pkg/errors"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	"github.com/CoreumFoundation/coreum/v2/app"
	"github.com/CoreumFoundation/coreum/v2/pkg/config"
	"github.com/CoreumFoundation/coreum/v2/pkg/config/constant"
)

var defaultConsensusParams = &abci.ConsensusParams{
	Block: &abci.BlockParams{
		MaxBytes: 200000,
		MaxGas:   2000000,
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour,
		MaxBytes:        10000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
}

// Option represents simapp customisations.
type Option func() tmdb.DB

// WithCustomDB returns the simapp Option to run with different DB.
func WithCustomDB(db tmdb.DB) Option {
	return func() tmdb.DB {
		return db
	}
}

// App is a simulation app wrapper.
type App struct {
	app.App
}

// New creates application instance with in-memory database and disabled logging.
func New(options ...Option) *App {
	var db tmdb.DB

	db = tmdb.NewMemDB()
	logger := log.NewNopLogger()

	for _, option := range options {
		customDB := option()
		if customDB != nil {
			db = customDB
		}
	}

	network, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	if err != nil {
		panic(err)
	}

	app.ChosenNetwork = network
	encoding := config.NewEncodingConfig(app.ModuleBasics)

	coreApp := app.New(logger, db, nil, true, map[int64]bool{}, "", 0, encoding,
		simapp.EmptyAppOptions{})

	genesisState := app.ModuleBasics.DefaultGenesis(encoding.Codec)
	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	if err != nil {
		panic(fmt.Sprintf("can't Marshal genesisState: %v", err))
	}
	coreApp.InitChain(abci.RequestInitChain{
		ConsensusParams: defaultConsensusParams,
		AppStateBytes:   stateBytes,
	})

	simApp := &App{*coreApp}

	return simApp
}

// BeginNextBlock begins new SimApp block and returns the ctx of the new block.
func (s *App) BeginNextBlock(blockTime time.Time) sdk.Context {
	if blockTime.IsZero() {
		blockTime = time.Now()
	}
	header := tmproto.Header{Height: s.LastBlockHeight() + 1, Time: blockTime}
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

	txGen := config.NewEncodingConfig(app.ModuleBasics).TxConfig

	tx, err := helpers.GenTx(
		txGen,
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

	return s.Deliver(txGen.TxEncoder(), tx)
}
