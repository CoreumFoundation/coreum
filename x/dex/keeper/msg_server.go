package keeper

import (
	"context"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

var _ types.MsgServer = MsgServer{}

// MsgServer serves grpc tx requests for assets module.
type MsgServer struct {
}

// NewMsgServer returns a new instance of the MsgServer.
func NewMsgServer() MsgServer {
	return MsgServer{}
}

func (ms MsgServer) CreateLimitOrder(ctx context.Context, order *types.MsgCreateLimitOrder) (*types.EmptyResponse, error) {
	//TODO implement me
	panic("implement me")
}
