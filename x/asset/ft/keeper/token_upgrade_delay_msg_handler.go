package keeper

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
)

// TokenUpgradeV1Keeper defines methods required to update tokens to V1.
type TokenUpgradeV1Keeper interface {
	UpgradeTokenToV1(ctx sdk.Context, data *types.DelayedTokenUpgradeV1) error
}

// NewDelayTokenUpgradeV1Handler handles token V1 upgrade.
func NewDelayTokenUpgradeV1Handler(keeper TokenUpgradeV1Keeper) func(ctx sdk.Context, data proto.Message) error {
	return func(ctx sdk.Context, data proto.Message) error {
		msg, ok := data.(*types.DelayedTokenUpgradeV1)
		if !ok {
			return sdkerrors.Wrapf(types.ErrInvalidState, "unrecognized %s message type: %T", types.ModuleName, data)
		}

		return keeper.UpgradeTokenToV1(ctx, msg)
	}
}
