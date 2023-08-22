package simulation_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/kv"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v2/app"
	"github.com/CoreumFoundation/coreum/v2/pkg/config"
	"github.com/CoreumFoundation/coreum/v2/x/nft"
	"github.com/CoreumFoundation/coreum/v2/x/nft/keeper"
	"github.com/CoreumFoundation/coreum/v2/x/nft/simulation"
)

var (
	ownerPk1   = ed25519.GenPrivKey().PubKey()
	ownerAddr1 = sdk.AccAddress(ownerPk1.Address())
)

func TestDecodeStore(t *testing.T) {
	cdc := config.NewEncodingConfig(app.ModuleBasics).Codec
	dec := simulation.NewDecodeStore(cdc)

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	classID := fmt.Sprintf("%s-%s", "myclass", addr.String())
	class := nft.Class{
		Id:          classID,
		Name:        "ClassName",
		Symbol:      "classsymbol",
		Description: "ClassDescription",
		Uri:         "ClassURI",
	}
	classBz, err := cdc.Marshal(&class)
	require.NoError(t, err)

	nft := nft.NFT{
		ClassId: classID,
		Id:      "NFTID",
		Uri:     "NFTURI",
	}
	nftBz, err := cdc.Marshal(&nft)
	require.NoError(t, err)

	nftOfClassByOwnerValue := []byte{0x01}

	totalSupply := 1
	totalSupplyBz := sdk.Uint64ToBigEndian(1)

	kvPairs := kv.Pairs{
		Pairs: []kv.Pair{
			{Key: keeper.ClassKey, Value: classBz},
			{Key: keeper.NFTKey, Value: nftBz},
			{Key: keeper.NFTOfClassByOwnerKey, Value: nftOfClassByOwnerValue},
			{Key: keeper.OwnerKey, Value: ownerAddr1},
			{Key: keeper.ClassTotalSupply, Value: totalSupplyBz},
			{Key: []byte{0x99}, Value: []byte{0x99}},
		},
	}

	tests := []struct {
		name        string
		expectErr   bool
		expectedLog string
	}{
		{"Class", false, fmt.Sprintf("%v\n%v", class, class)},
		{"NFT", false, fmt.Sprintf("%v\n%v", nft, nft)},
		{"NFTOfClassByOwnerKey", false, fmt.Sprintf("%v\n%v", nftOfClassByOwnerValue, nftOfClassByOwnerValue)},
		{"OwnerKey", false, fmt.Sprintf("%v\n%v", ownerAddr1, ownerAddr1)},
		{"ClassTotalSupply", false, fmt.Sprintf("%v\n%v", totalSupply, totalSupply)},
		{"other", true, ""},
	}

	for i, tt := range tests {
		i, tt := i, tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectErr {
				require.Panics(t, func() { dec(kvPairs.Pairs[i], kvPairs.Pairs[i]) }, tt.name)
			} else {
				require.Equal(t, tt.expectedLog, dec(kvPairs.Pairs[i], kvPairs.Pairs[i]), tt.name)
			}
		})
	}
}
