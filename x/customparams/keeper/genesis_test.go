package keeper_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v6/x/customparams/types"
)

func TestKeeper_InitAndExportGenesis(t *testing.T) {
	testApp := simapp.New()
	keeper := testApp.CustomParamsKeeper
	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})

	genState := types.GenesisState{
		StakingParams: types.StakingParams{
			MinSelfDelegation: sdkmath.OneInt(),
		},
	}
	keeper.InitGenesis(ctx, genState)

	requireT := require.New(t)
	params, err := keeper.GetStakingParams(ctx)
	requireT.NoError(err)
	requireT.Equal(sdkmath.OneInt().String(), params.MinSelfDelegation.String())

	exportedGetState := keeper.ExportGenesis(ctx)
	requireT.Equal(genState, *exportedGetState)
}
