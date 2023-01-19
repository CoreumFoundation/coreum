package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

func newFreezingFeatureStore(kvstore sdk.KVStore) nftFreezingStore {
	return nftFreezingStore{
		store: prefix.NewStore(kvstore, types.NFTFreezingKeyPrefix),
	}
}

type nftFreezingStore struct {
	store prefix.Store
}

func (s nftFreezingStore) freeze(classID, nftID string) {
	s.store.Set(types.CreateFreezingKey(classID, nftID), []byte{0x1})
}

func (s nftFreezingStore) unfreeze(classID, nftID string) {
	s.store.Delete(types.CreateFreezingKey(classID, nftID))
}

func (s nftFreezingStore) isFrozen(classID, nftID string) bool {
	return s.store.Has(types.CreateFreezingKey(classID, nftID))
}

func (s nftFreezingStore) allFrozen(q *query.PageRequest) (*query.PageResponse, []types.FrozenNFT, error) {
	mp := make(map[string][]string, 0)
	pageRes, err := query.Paginate(s.store, q, func(key, value []byte) error {
		classID, nftID, err := types.ParseFreezingKey(key)
		if err != nil {
			return err
		}
		mp[classID] = append(mp[classID], nftID)
		return nil
	})
	if err != nil {
		return nil, nil, err
	}

	var frozen []types.FrozenNFT
	for classID, nfts := range mp {
		frozen = append(frozen, types.FrozenNFT{
			ClassID: classID,
			NftIDs:  nfts,
		})
	}

	return pageRes, frozen, nil
}
