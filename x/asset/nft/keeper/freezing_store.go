package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"

	"github.com/CoreumFoundation/coreum/pkg/store"
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
	s.store.Set(s.genCompositeKey(classID, nftID), []byte{0x1})
}

func (s nftFreezingStore) unfreeze(classID, nftID string) {
	s.store.Delete(s.genCompositeKey(classID, nftID))
}

func (s nftFreezingStore) isFrozen(classID, nftID string) bool {
	return s.store.Has(s.genCompositeKey(classID, nftID))
}

func (s nftFreezingStore) allFrozen() ([]types.FrozenNFT, error) {
	mp := make(map[string][]string, 0)
	_, err := query.Paginate(s.store, &query.PageRequest{Limit: query.MaxLimit}, func(key, value []byte) error {
		classID, nftID, err := s.parseCompositeKey(key)
		if err != nil {
			return err
		}
		mp[classID] = append(mp[classID], nftID)
		return nil
	})
	if err != nil {
		return nil, err
	}

	var frozen []types.FrozenNFT
	for classID, nfts := range mp {
		frozen = append(frozen, types.FrozenNFT{
			ClassID: classID,
			NftIDs:  nfts,
		})
	}

	return frozen, nil
}

func (s nftFreezingStore) genCompositeKey(classID, nftID string) []byte {
	return store.JoinKeysWithLengthMany([]byte(classID), []byte(nftID))
}

func (s nftFreezingStore) parseCompositeKey(key []byte) (classID, nftID string, err error) {
	parsedKeys := store.ParseJoinKeysWithLengthMany(key)
	if len(parsedKeys) != 2 {
		err = types.ErrInvalidKey
		return
	}
	classID = string(parsedKeys[0])
	nftID = string(parsedKeys[1])
	return classID, nftID, nil
}
