package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/coreumfoundation/coreum/coreum/testutil/keeper"
	"github.com/coreumfoundation/coreum/coreum/x/issuance/keeper"
	"github.com/coreumfoundation/coreum/coreum/x/issuance/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.IssuanceKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
