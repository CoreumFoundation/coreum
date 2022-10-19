// Package simapp contains utils to bootstrap the chain.
package simapp

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/simapp"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
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

// New creates application instance with in-memory database and disabled logging.
func New() *app.App {
	db := tmdb.NewMemDB()
	logger := log.NewNopLogger()

	network, err := config.NetworkByChainID(config.ChainIDDev)
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

	return coreApp
}
