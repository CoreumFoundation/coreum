package keeper

import (
	"github.com/coreumfoundation/coreum/coreum/x/issuance/types"
)

var _ types.QueryServer = Keeper{}
