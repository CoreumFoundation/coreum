package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// MsgKeeper defines subscope of keeper methods required by msg service.
type MsgKeeper interface {
	IssueAsset(ctx sdk.Context, name string) string
}

// MsgServer serves grpc tx requests for assets module.
type MsgServer struct {
	keeper MsgKeeper
}

// NewMsgServer returns a new instance of the MsgServer.
func NewMsgServer(keeper MsgKeeper) MsgServer {
	return MsgServer{
		keeper: keeper,
	}
}

// IssueAsset defines a tx handler to issue a new asset.
func (ms MsgServer) IssueAsset(ctx context.Context, req *types.MsgIssueAsset) (*types.MsgIssueAssetResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ms.keeper.IssueAsset(sdk.UnwrapSDKContext(ctx), req.Name)

	return &types.MsgIssueAssetResponse{}, nil
}
