package delay_test

import (
	"testing"
	"time"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v4/x/delay"
	"github.com/CoreumFoundation/coreum/v4/x/delay/types"
)

func TestInitAndExportGenesis(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()

	ctx := testApp.BaseApp.NewContextLegacy(false, tmproto.Header{})
	keeper := testApp.DelayKeeper

	msg := &assetfttypes.DelayedTokenUpgradeV1{
		Denom: "denom",
	}

	anyMsg1, err := codectypes.NewAnyWithValue(msg)
	requireT.NoError(err)
	anyMsg2, err := codectypes.NewAnyWithValue(msg)
	requireT.NoError(err)
	anyMsg3, err := codectypes.NewAnyWithValue(msg)
	requireT.NoError(err)

	genState := types.GenesisState{
		DelayedItems: []types.DelayedItem{
			{
				ID:            "item1",
				ExecutionTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
				Data:          anyMsg1,
			},
			{
				ID:            "item2",
				ExecutionTime: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
				Data:          anyMsg2,
			},
			{
				ID:            "item3",
				ExecutionTime: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
				Data:          anyMsg3,
			},
		},
		BlockItems: []types.BlockItem{
			{
				ID:     "item4",
				Height: 1,
				Data:   anyMsg3,
			},
			{
				ID:     "item5",
				Height: 2,
				Data:   anyMsg2,
			},
			{
				ID:     "item6",
				Height: 3,
				Data:   anyMsg1,
			},
		},
	}

	require.NoError(t, genState.Validate())
	delay.InitGenesis(ctx, keeper, genState)

	gotGenState := delay.ExportGenesis(ctx, keeper)

	wantJSON, err := testApp.AppCodec().MarshalJSON(&genState)
	require.NoError(t, err)
	gotJSON, err := testApp.AppCodec().MarshalJSON(gotGenState)
	require.NoError(t, err)

	require.Equal(t, wantJSON, gotJSON)
}
