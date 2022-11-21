package wstaking

import (
	"github.com/CoreumFoundation/coreum/testutil/simapp"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

func TestAppModule_InitExportGenesis(t *testing.T) {
	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	module := NewAppModule(testApp.WStakingKeeper)
	module.InitGenesis(ctx)

}
