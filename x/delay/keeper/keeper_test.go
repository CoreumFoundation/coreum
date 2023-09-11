package keeper_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v3/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v3/x/delay/types"
)

var (
	_ codec.ProtoMarshaler = &delayedItem{}
	_ proto.Message        = &delayedItem{}
)

type delayedItem struct {
	Value string
}

//nolint:revive,stylecheck // underscore is required
func (di *delayedItem) XXX_MessageName() string {
	return "test.DummyDelayedItem"
}

func (di *delayedItem) Marshal() ([]byte, error) {
	return []byte(di.Value), nil
}

func (di *delayedItem) MarshalTo(data []byte) (n int, err error) {
	copy(data, di.Value)
	return len(di.Value), nil
}

func (di *delayedItem) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	copy(dAtA, di.Value)
	return len(di.Value), nil
}

func (di *delayedItem) Size() int {
	return len(di.Value)
}

func (di *delayedItem) Unmarshal(data []byte) error {
	di.Value = string(data)
	return nil
}

func (di *delayedItem) Reset() {
	di.Value = ""
}

func (di *delayedItem) String() string {
	return di.Value
}

func (di *delayedItem) ProtoMessage() {}

func TestDelayedExecution(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	testApp.InterfaceRegistry().RegisterImplementations((*codec.ProtoMarshaler)(nil), &delayedItem{})

	blockTime := time.Date(2023, 4, 3, 2, 3, 4, 0, time.UTC)
	ctx := testApp.BeginNextBlock(blockTime)

	delayed1 := &delayedItem{
		Value: "value1",
	}
	delayed2 := &delayedItem{
		Value: "value2",
	}
	delayed3 := &delayedItem{
		Value: "value3",
	}
	delayed4 := &delayedItem{
		Value: "value4",
	}

	delayKeeper := testApp.DelayKeeper

	requireT.NoError(delayKeeper.DelayExecution(ctx, "delayed-id-1", delayed1, time.Second))
	// same id and time fails
	requireT.Error(delayKeeper.DelayExecution(ctx, "delayed-id-1", delayed1, time.Second))
	// same id but different time succeeds
	requireT.NoError(delayKeeper.DelayExecution(ctx, "delayed-id-1", delayed2, 2*time.Second))

	// two items intentionally executed at the same time
	requireT.NoError(delayKeeper.DelayExecution(ctx, "delayed-id-3", delayed3, 3*time.Second))
	requireT.NoError(delayKeeper.DelayExecution(ctx, "delayed-id-4", delayed4, 3*time.Second))

	requireT.Error(delayKeeper.StoreDelayedExecution(ctx, "delayed-id-4", delayed1, time.Date(1969, 12, 31, 23, 59, 59, 0, time.UTC)))

	delayedItems, err := delayKeeper.ExportDelayedItems(ctx)
	requireT.NoError(err)
	requireT.Len(delayedItems, 4)

	expectedDelayedItems := []types.DelayedItem{
		{
			Id:            "delayed-id-1",
			ExecutionTime: blockTime.Add(time.Second),
			Data:          newAny(requireT, delayed1),
		},
		{
			Id:            "delayed-id-1",
			ExecutionTime: blockTime.Add(2 * time.Second),
			Data:          newAny(requireT, delayed2),
		},
		{
			Id:            "delayed-id-3",
			ExecutionTime: blockTime.Add(3 * time.Second),
			Data:          newAny(requireT, delayed3),
		},
		{
			Id:            "delayed-id-4",
			ExecutionTime: blockTime.Add(3 * time.Second),
			Data:          newAny(requireT, delayed4),
		},
	}

	requireT.Equal(expectedDelayedItems, delayedItems)

	// should panic because handler is not registered
	requireT.Panics(func() {
		testApp.BeginNextBlock(blockTime.Add(time.Second))
	})

	executedItems := []*delayedItem{}
	requireT.NoError(delayKeeper.Router().RegisterHandler(&delayedItem{}, func(ctx sdk.Context, data proto.Message) error {
		executedItems = append(executedItems, data.(*delayedItem))
		return nil
	}))

	// first item should be executed
	testApp.BeginNextBlock(blockTime.Add(time.Second))
	requireT.Len(executedItems, 1)
	requireT.Equal(delayed1, executedItems[0])

	// three items should be executed
	executedItems = []*delayedItem{}
	testApp.BeginNextBlock(blockTime.Add(3 * time.Second))
	requireT.Len(executedItems, 3)
	requireT.Equal(delayed2, executedItems[0])
	requireT.Equal(delayed3, executedItems[1])
	requireT.Equal(delayed4, executedItems[2])

	// no delayed items should be stored now
	delayedItems, err = delayKeeper.ExportDelayedItems(ctx)
	requireT.NoError(err)
	requireT.Empty(delayedItems)
}

func newAny(requireT *require.Assertions, data codec.ProtoMarshaler) *codectypes.Any {
	v, err := codectypes.NewAnyWithValue(data)
	requireT.NoError(err)
	return &codectypes.Any{
		TypeUrl: v.TypeUrl,
		Value:   v.Value,
	}
}
