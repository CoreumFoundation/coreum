package nft_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/nft"
	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

func TestInitAndExportGenesis(t *testing.T) {
	assertT := assert.New(t)

	testApp := simapp.New()

	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	nftKeeper := testApp.AssetNFTKeeper

	// prepare the genesis data

	genState := types.GenesisState{
		Params: types.DefaultParams(),
	}

	// init the keeper
	nft.InitGenesis(ctx, nftKeeper, genState)

	// assert the keeper state

	// params

	params := nftKeeper.GetParams(ctx)
	assertT.EqualValues(types.DefaultParams(), params)

	// check that export is equal import
	exportedGenState := nft.ExportGenesis(ctx, nftKeeper)

	assertT.EqualValues(genState.Params, exportedGenState.Params)
}
