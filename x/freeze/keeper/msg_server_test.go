package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
    "github.com/CoreumFoundation/coreum/x/freeze/types"
    "github.com/CoreumFoundation/coreum/x/freeze/keeper"
    keepertest "github.com/CoreumFoundation/coreum/testutil/keeper"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.FreezeKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
