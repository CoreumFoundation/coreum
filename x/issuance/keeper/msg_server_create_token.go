package keeper

import (
	"context"

	"github.com/coreumfoundation/coreum/coreum/x/issuance/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CreateToken(goCtx context.Context, msg *types.MsgCreateToken) (*types.MsgCreateTokenResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	ctx.Logger().Info("Mint coints and handle the message")
	_ = ctx

	return &types.MsgCreateTokenResponse{}, nil
}
