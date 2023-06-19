package upgrade

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assetnfttypes "github.com/CoreumFoundation/coreum/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/x/nft"
)

type nftTest struct {
	issuer        sdk.AccAddress
	issuedEvent   *assetnfttypes.EventClassIssued
	expectedClass assetnfttypes.Class
	mintMsg       *assetnfttypes.MsgMint
	expectedNFT   nft.NFT
}

func (n *nftTest) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	// create NFT class and mint NFT to check the keys migration
	n.issuer = chain.GenAccount()
	assetNftClient := assetnfttypes.NewQueryClient(chain.ClientContext)
	nfqQueryClient := nft.NewQueryClient(chain.ClientContext)
	chain.FundAccountsWithOptions(ctx, t, n.issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
			&assetnfttypes.MsgMint{},
		},
	})

	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer:      n.issuer.String(),
		Symbol:      "symbol",
		Name:        "name",
		Description: "description",
		URI:         "https://my-class-meta.invalid/1",
		URIHash:     "content-hash",
		RoyaltyRate: sdk.ZeroDec(),
	}
	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(n.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)
	tokenIssuedEvents, err := event.FindTypedEvents[*assetnfttypes.EventClassIssued](res.Events)
	requireT.NoError(err)
	n.issuedEvent = tokenIssuedEvents[0]

	// query nft class
	assetNftClassRes, err := assetNftClient.Class(ctx, &assetnfttypes.QueryClassRequest{
		Id: n.issuedEvent.ID,
	})
	requireT.NoError(err)

	n.expectedClass = assetnfttypes.Class{
		Id:          n.issuedEvent.ID,
		Issuer:      n.issuer.String(),
		Symbol:      issueMsg.Symbol,
		Name:        issueMsg.Name,
		Description: issueMsg.Description,
		URI:         issueMsg.URI,
		URIHash:     issueMsg.URIHash,
		RoyaltyRate: issueMsg.RoyaltyRate,
	}
	requireT.Equal(n.expectedClass, assetNftClassRes.Class)

	n.mintMsg = &assetnfttypes.MsgMint{
		Sender:  n.issuer.String(),
		ID:      "id-1",
		ClassID: n.issuedEvent.ID,
		URI:     "https://my-class-meta.invalid/1",
		URIHash: "content-hash",
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(n.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(n.mintMsg)),
		n.mintMsg,
	)
	requireT.NoError(err)

	n.expectedNFT = nft.NFT{
		ClassId: n.issuedEvent.ID,
		Id:      n.mintMsg.ID,
		Uri:     n.mintMsg.URI,
		UriHash: n.mintMsg.URIHash,
	}

	nftRes, err := nfqQueryClient.NFT(ctx, &nft.QueryNFTRequest{
		ClassId: n.mintMsg.ClassID,
		Id:      n.mintMsg.ID,
	})
	requireT.NoError(err)
	requireT.Equal(n.expectedNFT, *nftRes.Nft)
}

func (n *nftTest) After(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	assetNftClient := assetnfttypes.NewQueryClient(chain.ClientContext)
	nfqQueryClient := nft.NewQueryClient(chain.ClientContext)

	// query same nft class after the upgrade
	assetNftClassRes, err := assetNftClient.Class(ctx, &assetnfttypes.QueryClassRequest{
		Id: n.issuedEvent.ID,
	})
	requireT.NoError(err)
	requireT.Equal(n.expectedClass, assetNftClassRes.Class)

	//  query same nft after the upgrade
	nftRes, err := nfqQueryClient.NFT(ctx, &nft.QueryNFTRequest{
		ClassId: n.mintMsg.ClassID,
		Id:      n.mintMsg.ID,
	})
	requireT.NoError(err)
	requireT.Equal(n.expectedNFT, *nftRes.Nft)

	// check that we can query the same NFT class now with the classes query
	assetNftClassesRes, err := assetNftClient.Classes(ctx, &assetnfttypes.QueryClassesRequest{
		Issuer: n.issuer.String(),
	})
	requireT.NoError(err)
	requireT.Equal(1, len(assetNftClassesRes.Classes))
	requireT.Equal(uint64(1), assetNftClassesRes.Pagination.Total)
	requireT.Equal(n.expectedClass, assetNftClassesRes.Classes[0])
}
