package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/x/snapshot/store"
	"github.com/CoreumFoundation/coreum/x/snapshot/types"
)

// AddListenersDecorator adds listeners to the multistore so snapshot module is notified whenever any key of interest is modified
type AddListenersDecorator struct {
	snapshotKey     sdk.StoreKey
	transformations map[sdk.StoreKey][]types.Transformation
}

// NewAddListenersDecorator creates new AddListenersDecorator
func NewAddListenersDecorator(snapshotKey sdk.StoreKey, transformations []types.Transformation) AddListenersDecorator {
	ts := map[sdk.StoreKey][]types.Transformation{}
	for _, t := range transformations {
		ts[t.StoreKey()] = append(ts[t.StoreKey()], t)
	}
	return AddListenersDecorator{
		snapshotKey:     snapshotKey,
		transformations: ts,
	}
}

// AnteHandle resets the gas limit inside GasMeter
func (ald AddListenersDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	return next(ctx.WithMultiStore(store.New(ctx.MultiStore(), ald.snapshotKey, ald.transformations)), tx, simulate)
}
