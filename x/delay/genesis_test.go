package delay_test

import (
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/v2/testutil/simapp"
	assetfttypes "github.com/CoreumFoundation/coreum/v2/x/asset/ft/types"
	"github.com/CoreumFoundation/coreum/v2/x/delay/types"
)

func TestInitAndExportGenesis(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	testApp := simapp.New()

	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})
	keeper := testApp.DelayKeeper

	// prepare the genesis data

	msg1 := &banktypes.MsgSend{
		FromAddress: "senderAddress1",
		ToAddress:   "recipientAddress1",
		Amount:      sdk.NewCoins(sdk.NewCoin("denom", sdk.OneInt())),
	}
	msg2 := &assetfttypes.DelayedTokenUpgradeV1{
		Denom: "denom",
	}
	msg3 := &banktypes.MsgSend{
		FromAddress: "senderAddress2",
		ToAddress:   "recipientAddress2",
		Amount:      sdk.NewCoins(sdk.NewCoin("denom", sdk.OneInt())),
	}
	anyMsg1, err := codectypes.NewAnyWithValue(msg1)
	requireT.NoError(err)
	anyMsg2, err := codectypes.NewAnyWithValue(msg2)
	requireT.NoError(err)
	anyMsg3, err := codectypes.NewAnyWithValue(msg3)
	requireT.NoError(err)

	// time sequence of items is intentionally reverted
	items := []types.DelayedItem{
		{
			Id:            "item3",
			ExecutionTime: time.Date(2023, 1, 3, 0, 0, 0, 0, time.UTC),
			Data:          anyMsg3,
		},
		{
			Id:            "item2",
			ExecutionTime: time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC),
			Data:          anyMsg2,
		},
		{
			Id:            "item1",
			ExecutionTime: time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC),
			Data:          anyMsg1,
		},
	}

	requireT.NoError(keeper.ImportDelayedItems(ctx, items))
	itemsExported, err := keeper.ExportDelayedItems(ctx)
	requireT.NoError(err)
	requireT.Len(itemsExported, len(items))

	assertT.Equal(newDelayedItemWithoutCache(items[2]), newDelayedItemWithoutCache(itemsExported[0]))
	assertT.Equal(newDelayedItemWithoutCache(items[1]), newDelayedItemWithoutCache(itemsExported[1]))
	assertT.Equal(newDelayedItemWithoutCache(items[0]), newDelayedItemWithoutCache(itemsExported[2]))
}

func newDelayedItemWithoutCache(item types.DelayedItem) types.DelayedItem {
	return types.DelayedItem{
		Id:            item.Id,
		ExecutionTime: item.ExecutionTime,
		Data: &codectypes.Any{
			TypeUrl: item.Data.TypeUrl,
			Value:   item.Data.Value,
		},
	}
}
