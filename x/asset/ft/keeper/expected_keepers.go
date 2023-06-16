package keeper

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DelayKeeper defines methods required from the delay keeper.
type DelayKeeper interface {
	DelayExecution(ctx sdk.Context, id string, data codec.ProtoMarshaler, delay time.Duration) error
}
