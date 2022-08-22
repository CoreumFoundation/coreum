package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/CoreumFoundation/coreum/testutil/keeper"
	"github.com/CoreumFoundation/coreum/x/freeze/keeper"
	"github.com/CoreumFoundation/coreum/x/freeze/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.FreezeKeeper(t)
	return keeper.NewMsgServerImpl(k), sdk.WrapSDKContext(ctx)
}
