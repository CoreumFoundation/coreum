package keeper_test

import (
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v6/x/delay/types"
)

var _ proto.Message = &dummyExecutionMessage{}

type dummyExecutionMessage struct {
	Value string
}

//nolint:revive,staticcheck // underscore is required
func (di *dummyExecutionMessage) XXX_MessageName() string {
	return "test.DummyExecutionMessage"
}

func (di *dummyExecutionMessage) Marshal() ([]byte, error) {
	return []byte(di.Value), nil
}

func (di *dummyExecutionMessage) MarshalTo(data []byte) (n int, err error) {
	copy(data, di.Value)
	return len(di.Value), nil
}

func (di *dummyExecutionMessage) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	copy(dAtA, di.Value)
	return len(di.Value), nil
}

func (di *dummyExecutionMessage) Size() int {
	return len(di.Value)
}

func (di *dummyExecutionMessage) Unmarshal(data []byte) error {
	di.Value = string(data)
	return nil
}

func (di *dummyExecutionMessage) Reset() {
	di.Value = ""
}

func (di *dummyExecutionMessage) String() string {
	return di.Value
}

func (di *dummyExecutionMessage) ProtoMessage() {}

func TestTimeExecution(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	testApp.InterfaceRegistry().RegisterImplementations((*proto.Message)(nil), &dummyExecutionMessage{})

	blockTime := time.Date(2023, 4, 3, 2, 3, 4, 0, time.UTC)
	ctx, _, err := testApp.BeginNextBlockAtTime(blockTime)
	requireT.NoError(err)

	delayed1 := &dummyExecutionMessage{
		Value: "value1",
	}
	delayed2 := &dummyExecutionMessage{
		Value: "value2",
	}
	delayed3 := &dummyExecutionMessage{
		Value: "value3",
	}
	delayed4 := &dummyExecutionMessage{
		Value: "value4",
	}

	// register the handler
	executedItems := make([]*dummyExecutionMessage, 0)
	requireT.NoError(testApp.DelayKeeper.Router().RegisterHandler(
		&dummyExecutionMessage{}, func(ctx sdk.Context, data proto.Message) error {
			executedItems = append(executedItems, data.(*dummyExecutionMessage))
			return nil
		}))

	requireT.NoError(testApp.DelayKeeper.DelayExecution(ctx, "delayed-id-1", delayed1, time.Second))
	// same id and time fails
	requireT.Error(testApp.DelayKeeper.DelayExecution(ctx, "delayed-id-1", delayed1, time.Second))
	// same id but different time succeeds
	requireT.NoError(testApp.DelayKeeper.DelayExecution(ctx, "delayed-id-1", delayed2, 2*time.Second))

	// two items intentionally executed at the same time
	requireT.NoError(testApp.DelayKeeper.ExecuteAfter(ctx, "delayed-id-3", delayed3, ctx.BlockTime().Add(3*time.Second)))
	requireT.NoError(testApp.DelayKeeper.ExecuteAfter(ctx, "delayed-id-4", delayed4, ctx.BlockTime().Add(3*time.Second)))
	// save and remove
	requireT.NoError(testApp.DelayKeeper.ExecuteAfter(ctx, "delayed-id-5", delayed4, ctx.BlockTime().Add(3*time.Second)))
	requireT.NoError(testApp.DelayKeeper.RemoveExecuteAfter(ctx, "delayed-id-5", ctx.BlockTime().Add(3*time.Second)))

	requireT.Error(testApp.DelayKeeper.StoreDelayedExecution(
		ctx, "delayed-id-4", delayed1, time.Date(1969, 12, 31, 23, 59, 59, 0, time.UTC),
	))

	delayedItems, err := testApp.DelayKeeper.ExportDelayedItems(ctx)
	requireT.NoError(err)
	requireT.Len(delayedItems, 4)

	expectedDelayedItems := []types.DelayedItem{
		{
			ID:            "delayed-id-1",
			ExecutionTime: blockTime.Add(time.Second),
			Data:          newAny(requireT, delayed1),
		},
		{
			ID:            "delayed-id-1",
			ExecutionTime: blockTime.Add(2 * time.Second),
			Data:          newAny(requireT, delayed2),
		},
		{
			ID:            "delayed-id-3",
			ExecutionTime: blockTime.Add(3 * time.Second),
			Data:          newAny(requireT, delayed3),
		},
		{
			ID:            "delayed-id-4",
			ExecutionTime: blockTime.Add(3 * time.Second),
			Data:          newAny(requireT, delayed4),
		},
	}

	requireT.Equal(expectedDelayedItems, delayedItems)

	blockTime = blockTime.Add(time.Second)
	_, _, err = testApp.BeginNextBlockAtTime(blockTime)
	requireT.NoError(err)

	// first item should be executed
	requireT.NoError(testApp.FinalizeBlockAtTime(blockTime))
	_, _, err = testApp.BeginNextBlockAtTime(blockTime.Add(time.Second))
	requireT.NoError(err)

	requireT.Len(executedItems, 2)
	requireT.Equal(
		[]*dummyExecutionMessage{
			delayed1,
			delayed2,
		},
		executedItems,
	)

	blockTime = blockTime.Add(3 * time.Second)
	ctx, _, err = testApp.BeginNextBlockAtTime(blockTime)
	requireT.NoError(err)

	requireT.Len(executedItems, 4)
	requireT.Equal(
		[]*dummyExecutionMessage{
			// left from prev execution
			delayed1,
			delayed2,
			// new
			delayed3,
			delayed4,
		},
		executedItems,
	)

	// no delayed items should be stored now
	delayedItems, err = testApp.DelayKeeper.ExportDelayedItems(ctx)
	requireT.NoError(err)
	requireT.Empty(delayedItems)
}

func TestBlockExecution(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	testApp.InterfaceRegistry().RegisterImplementations((*proto.Message)(nil), &dummyExecutionMessage{})

	ctx, _, err := testApp.BeginNextBlockAtHeight(20)
	require.NoError(t, err)

	blockExec1 := &dummyExecutionMessage{
		Value: "value1",
	}
	blockExec2 := &dummyExecutionMessage{
		Value: "value2",
	}
	blockExec3 := &dummyExecutionMessage{
		Value: "value3",
	}

	// register the handler
	executedItems := make([]*dummyExecutionMessage, 0)
	requireT.NoError(testApp.DelayKeeper.Router().RegisterHandler(
		&dummyExecutionMessage{}, func(ctx sdk.Context, data proto.Message) error {
			executedItems = append(executedItems, data.(*dummyExecutionMessage))
			return nil
		}))

	requireT.NoError(testApp.DelayKeeper.ExecuteAfterBlock(ctx, "id-1", blockExec1, 30))
	// same id and height fails
	requireT.Error(testApp.DelayKeeper.ExecuteAfterBlock(ctx, "id-1", blockExec1, 30))
	// same id but different height succeeds
	requireT.NoError(testApp.DelayKeeper.ExecuteAfterBlock(ctx, "id-2", blockExec2, 32))

	// same height different ID the `blockExec3` should not be executed
	requireT.NoError(testApp.DelayKeeper.ExecuteAfterBlock(ctx, "id-3", blockExec3, 31))
	requireT.NoError(testApp.DelayKeeper.RemoveExecuteAtBlock(ctx, "id-3", 31))

	blockItems, err := testApp.DelayKeeper.ExportBlockItems(ctx)
	requireT.NoError(err)
	requireT.Len(blockItems, 2)

	expectedBlockItems := []types.BlockItem{
		{
			ID:     "id-1",
			Height: 30,
			Data:   newAny(requireT, blockExec1),
		},
		{
			ID:     "id-2",
			Height: 32,
			Data:   newAny(requireT, blockExec2),
		},
	}
	requireT.Equal(expectedBlockItems, blockItems)

	_, _, err = testApp.BeginNextBlockAtHeight(31)
	requireT.NoError(err)

	requireT.Len(executedItems, 1)
	requireT.Equal(
		[]*dummyExecutionMessage{
			blockExec1,
		},
		executedItems,
	)

	ctx, _, err = testApp.BeginNextBlockAtHeight(33)
	requireT.NoError(err)

	requireT.Len(executedItems, 2)
	requireT.Equal(
		[]*dummyExecutionMessage{
			// left from prev execution
			blockExec1,
			// new
			blockExec2,
		},
		executedItems,
	)

	// no block items should be stored now
	blockItems, err = testApp.DelayKeeper.ExportBlockItems(ctx)
	requireT.NoError(err)
	requireT.Empty(blockItems)
}

func newAny(requireT *require.Assertions, data proto.Message) *codectypes.Any {
	v, err := codectypes.NewAnyWithValue(data)
	requireT.NoError(err)
	return &codectypes.Any{
		TypeUrl: v.TypeUrl,
		Value:   v.Value,
	}
}
