package keeper

import (
	"github.com/CoreumFoundation/coreum/x/freeze/types"
)

var _ types.QueryServer = BaseKeeper{}
