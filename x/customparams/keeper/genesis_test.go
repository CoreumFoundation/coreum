package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/v2/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v2/x/customparams/types"
)

func TestKeeper_InitAndExportGenesis(t *testing.T) {
	testApp := simapp.New()
	keeper := testApp.CustomParamsKeeper
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	genState := types.GenesisState{
		StakingParams: types.StakingParams{
			MinSelfDelegation: sdk.OneInt(),
		},
	}
	keeper.InitGenesis(ctx, genState)

	requireT := require.New(t)
	requireT.Equal(sdk.OneInt().String(), keeper.GetStakingParams(ctx).MinSelfDelegation.String())

	exportedGetState := keeper.ExportGenesis(ctx)
	requireT.Equal(genState, *exportedGetState)
}
