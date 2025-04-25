//go:build !objstore
// +build !objstore

package rootmulti

import (
	"fmt"

	"cosmossdk.io/store/types"
	"github.com/CoreumFaundation/coreum/memiavl"
)

func (rs *Store) loadExtraStore(_ *memiavl.DB, _ types.StoreKey, params storeParams) (types.CommitStore, error) {
	panic(fmt.Sprintf("unrecognized store type %v", params.typ))
}
