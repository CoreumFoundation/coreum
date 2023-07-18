package v1_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/v2/pkg/store"
	"github.com/CoreumFoundation/coreum/v2/testutil/simapp"
	v1 "github.com/CoreumFoundation/coreum/v2/x/asset/nft/legacy/v1"
	"github.com/CoreumFoundation/coreum/v2/x/asset/nft/types"
)

func TestMigrateStore(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.NewContext(false, tmproto.Header{})

	symbol := "mysymbol"

	address1 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	address2 := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())
	classID1 := types.BuildClassID(symbol, address1)
	classID2 := types.BuildClassID(symbol, address2)

	storeKey := testApp.GetKey(types.ModuleName)
	moduleStore := ctx.KVStore(storeKey)

	definition := types.ClassDefinition{
		Issuer: address1.String(),
		Features: []types.ClassFeature{
			types.ClassFeature_burning,
		},
		RoyaltyRate: sdk.MustNewDecFromStr("0.1"),
	}

	definition1 := definition
	definition1.ID = classID1

	definition2 := definition
	definition2.ID = classID2

	oldKey1 := createV1NFTClassKey(classID1)
	moduleStore.Set(oldKey1, testApp.AppCodec().MustMarshal(&definition1))

	oldKey2 := createV1NFTClassKey(classID2)
	moduleStore.Set(oldKey2, testApp.AppCodec().MustMarshal(&definition2))

	// check that records are not available by old keys
	_, err := testApp.AssetNFTKeeper.GetClassDefinition(ctx, classID1)
	requireT.ErrorIs(err, types.ErrClassNotFound)

	_, err = testApp.AssetNFTKeeper.GetClassDefinition(ctx, classID2)
	requireT.ErrorIs(err, types.ErrClassNotFound)

	err = v1.MigrateStore(ctx, storeKey)
	requireT.NoError(err)

	// check that records are available now
	gotDefinition1, err := testApp.AssetNFTKeeper.GetClassDefinition(ctx, classID1)
	requireT.NoError(err)
	requireT.Equal(definition1, gotDefinition1)

	gotDefinition2, err := testApp.AssetNFTKeeper.GetClassDefinition(ctx, classID2)
	requireT.NoError(err)
	requireT.Equal(definition2, gotDefinition2)
}

func createV1NFTClassKey(classID string) []byte {
	return store.JoinKeys(types.NFTClassKeyPrefix, []byte(classID))
}
