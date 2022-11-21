package keeper_test

import (
	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/wstaking/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
)

func TestKeeper_InitAndExportGenesis(t *testing.T) {
	testApp := simapp.New()
	keeper := testApp.WStakingKeeper
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	getState := types.GenesisState{
		Params: types.Params{
			MinSelfDelegation: sdk.OneInt(),
		},
	}
	keeper.InitGenesis(ctx, getState)

	requireT := require.New(t)
	requireT.Equal(sdk.OneInt().String(), keeper.GetParams(ctx).MinSelfDelegation.String())

	exportedGetState := keeper.ExportGenesis(ctx)
	requireT.Equal(getState, *exportedGetState)
}
