package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindAndRemoveNFT(t *testing.T) {
	testCases := []struct {
		name         string
		nfts         []NFTIdentifier
		classID      string
		nftID        string
		expectedNfts []NFTIdentifier
		found        bool
	}{
		{
			"nft not found",
			[]NFTIdentifier{
				{"class1", "nft1"},
				{"class2", "nft2"},
				{"class3", "nft3"},
			},
			"class", "nft",
			[]NFTIdentifier{
				{"class1", "nft1"},
				{"class2", "nft2"},
				{"class3", "nft3"},
			},
			false,
		},
		{
			"single element in list",
			[]NFTIdentifier{
				{"class1", "nft1"},
			},
			"class1", "nft1",
			[]NFTIdentifier{},
			true,
		},
		{
			"match start of the list",
			[]NFTIdentifier{
				{"class1", "nft1"},
				{"class2", "nft2"},
				{"class3", "nft3"},
			},
			"class1", "nft1",
			[]NFTIdentifier{
				{"class2", "nft2"},
				{"class3", "nft3"},
			},
			true,
		},
		{
			"match end of the list",
			[]NFTIdentifier{
				{"class1", "nft1"},
				{"class2", "nft2"},
				{"class3", "nft3"},
			},
			"class3", "nft3",
			[]NFTIdentifier{
				{"class1", "nft1"},
				{"class2", "nft2"},
			},
			true,
		},
		{
			"match middle of the list",
			[]NFTIdentifier{
				{"class1", "nft1"},
				{"class2", "nft2"},
				{"class3", "nft3"},
			},
			"class2", "nft2",
			[]NFTIdentifier{
				{"class1", "nft1"},
				{"class3", "nft3"},
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			a := NewSendAuthorization(tc.nfts)
			found := a.findAndRemoveNFT(tc.classID, tc.nftID)
			assert.EqualValues(t, tc.found, found)
			assert.EqualValues(t, tc.expectedNfts, a.Nfts)
		})
	}
}
