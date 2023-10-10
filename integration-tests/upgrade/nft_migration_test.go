//go:build integrationtests

package upgrade

import (
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosnft "github.com/cosmos/cosmos-sdk/x/nft"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v3/integration-tests"
	"github.com/CoreumFoundation/coreum/v3/pkg/client"
	"github.com/CoreumFoundation/coreum/v3/testutil/integration"
	assetnfttypes "github.com/CoreumFoundation/coreum/v3/x/asset/nft/types"
	cnft "github.com/CoreumFoundation/coreum/v3/x/nft"
	cnftkeeper "github.com/CoreumFoundation/coreum/v3/x/nft/keeper"
)

type nftMigrationTest struct {
	nfts    []*cnft.NFT
	classes []*cnft.Class
	issuer  sdk.AccAddress
}

func (nut *nftMigrationTest) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	nftClient := cnft.NewQueryClient(chain.ClientContext)

	issuer := chain.GenAccount()
	nut.issuer = issuer
	chain.FundAccountWithOptions(ctx, t, issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
			&assetnfttypes.MsgIssueClass{},
			&assetnfttypes.MsgMint{},
			&assetnfttypes.MsgMint{},
		},
		Amount: chain.QueryAssetNFTParams(ctx, t).MintFee.Amount,
	})

	// issue nft class and mint nfts
	jsonData := []byte(`{"name": "Name", "description": "Description"}`)
	data, err := codectypes.NewAnyWithValue(&assetnfttypes.DataBytes{Data: jsonData})
	requireT.NoError(err)

	symbol1 := "NFTClassSymbol1"
	symbol2 := "NFTClassSymbol2"

	issueAndMint := []sdk.Msg{
		&assetnfttypes.MsgIssueClass{
			Issuer: issuer.String(),
			Symbol: symbol1,
		},
		&assetnfttypes.MsgIssueClass{
			Issuer: issuer.String(),
			Symbol: symbol2,
		},
		&assetnfttypes.MsgMint{
			Sender:  issuer.String(),
			ID:      "nft-1",
			ClassID: assetnfttypes.BuildClassID(symbol1, issuer),
			URI:     "https://my-class-meta.invalid/1",
			URIHash: "content-hash-1",
			Data:    data,
		},
		&assetnfttypes.MsgMint{
			Sender:  issuer.String(),
			ID:      "nft-2",
			ClassID: assetnfttypes.BuildClassID(symbol2, issuer),
			URI:     "https://my-class-meta.invalid/2",
			URIHash: "content-hash-2",
			Data:    data,
		},
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueAndMint...)),
		issueAndMint...,
	)
	requireT.NoError(err)

	classes, err := nftClient.Classes(ctx, &cnft.QueryClassesRequest{}) //nolint:staticcheck // we are testing deprecated handlers
	requireT.NoError(err)
	nut.classes = classes.Classes

	nfts, err := nftClient.NFTs(ctx, &cnft.QueryNFTsRequest{Owner: issuer.String()}) //nolint:staticcheck // we are testing deprecated handlers
	requireT.NoError(err)
	nut.nfts = nfts.Nfts
}

func (nut *nftMigrationTest) After(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	nftClient := cosmosnft.NewQueryClient(chain.ClientContext)

	classes, err := nftClient.Classes(ctx, &cosmosnft.QueryClassesRequest{})
	requireT.NoError(err)
	requireT.ElementsMatch(nut.classes, cnftkeeper.ConvertFromCosmosClassList(classes.Classes))

	nfts, err := nftClient.NFTs(ctx, &cosmosnft.QueryNFTsRequest{Owner: nut.issuer.String()})
	requireT.NoError(err)
	requireT.ElementsMatch(nut.nfts, cnftkeeper.ConvertFromCosmosNFTList(nfts.Nfts))

	// try sending the nft minted before the upgrade
	recipient := chain.GenAccount()
	chain.FundAccountWithOptions(ctx, t, nut.issuer, integration.BalancesOptions{
		Messages: []sdk.Msg{
			&cosmosnft.MsgSend{},
			&cosmosnft.MsgSend{},
		},
	})

	sendMsg := []sdk.Msg{
		&cosmosnft.MsgSend{
			ClassId:  nfts.Nfts[0].ClassId,
			Id:       nfts.Nfts[0].Id,
			Sender:   nut.issuer.String(),
			Receiver: recipient.String(),
		},
		&cosmosnft.MsgSend{
			ClassId:  nfts.Nfts[1].ClassId,
			Id:       nfts.Nfts[1].Id,
			Sender:   nut.issuer.String(),
			Receiver: recipient.String(),
		},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(nut.issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg...)),
		sendMsg...,
	)
	requireT.NoError(err)

	nfts, err = nftClient.NFTs(ctx, &cosmosnft.QueryNFTsRequest{Owner: recipient.String()})
	requireT.NoError(err)
	requireT.Len(nfts.Nfts, 2)
}
