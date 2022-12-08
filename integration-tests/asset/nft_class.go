package asset

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
	"github.com/CoreumFoundation/coreum/x/nft"
)

// TestCreateNonFungibleTokenClass tests non-fungible token class creation.
func TestCreateNonFungibleTokenClass(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	creator := chain.GenAccount()

	nftClient := nft.NewQueryClient(chain.ClientContext)
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, creator, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgCreateNonFungibleTokenClass{},
			},
		}),
	)

	// create new NFT class
	createMsg := &assettypes.MsgCreateNonFungibleTokenClass{
		Creator:     creator.String(),
		Symbol:      "symbol",
		Name:        "name",
		Description: "description",
		URI:         "https://my-class-meta.int/1",
		URIHash:     "35b326a2b3b605270c26185c38d2581e937b2eae0418b4964ef521efe79cdf34",
	}
	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(creator),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(createMsg)),
		createMsg,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(createMsg), uint64(res.GasUsed))
	nonFungibleTokenCreatedEvt, err := event.FindTypedEvent[*assettypes.EventNonFungibleTokenClassCreated](res.Events)
	requireT.NoError(err)
	requireT.Equal(&assettypes.EventNonFungibleTokenClassCreated{
		ID:          assettypes.BuildNonFungibleTokenClassID(createMsg.Symbol, creator),
		Creator:     creator.String(),
		Symbol:      createMsg.Symbol,
		Name:        createMsg.Name,
		Description: createMsg.Description,
		URI:         createMsg.URI,
		URIHash:     createMsg.URIHash,
	}, nonFungibleTokenCreatedEvt)

	// check that class is present in the nft module
	nftClassRes, err := nftClient.Class(ctx, &nft.QueryClassRequest{
		ClassId: nonFungibleTokenCreatedEvt.ID,
	})
	requireT.NoError(err)

	requireT.Equal(&nft.Class{
		Id:          assettypes.BuildNonFungibleTokenClassID(createMsg.Symbol, creator),
		Symbol:      createMsg.Symbol,
		Name:        createMsg.Name,
		Description: createMsg.Description,
		Uri:         createMsg.URI,
		UriHash:     createMsg.URIHash,
	}, nftClassRes.Class)
}
