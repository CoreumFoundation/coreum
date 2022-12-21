//go:build integrationtests

package modules

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assetnfttypes "github.com/CoreumFoundation/coreum/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/x/nft"
)

// TestAssetNFTIssueClass tests non-fungible token class creation.
func TestAssetNFTIssueClass(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()

	nftClient := nft.NewQueryClient(chain.ClientContext)
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assetnfttypes.MsgIssueClass{},
			},
		}),
	)

	// issue new NFT class
	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer:      issuer.String(),
		Symbol:      "symbol",
		Name:        "name",
		Description: "description",
		URI:         "https://my-class-meta.invalid/1",
		URIHash:     "content-hash",
	}
	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(issueMsg), uint64(res.GasUsed))
	tokenIssuedEvents, err := event.FindTypedEvents[*assetnfttypes.EventClassIssued](res.Events)
	requireT.NoError(err)
	tokenIssuedEvent := tokenIssuedEvents[0]
	requireT.Equal(&assetnfttypes.EventClassIssued{
		ID:          assetnfttypes.BuildClassID(issueMsg.Symbol, issuer),
		Issuer:      issuer.String(),
		Symbol:      issueMsg.Symbol,
		Name:        issueMsg.Name,
		Description: issueMsg.Description,
		URI:         issueMsg.URI,
		URIHash:     issueMsg.URIHash,
	}, tokenIssuedEvent)

	// check that class is present in the nft module
	nftClassRes, err := nftClient.Class(ctx, &nft.QueryClassRequest{
		ClassId: tokenIssuedEvent.ID,
	})
	requireT.NoError(err)

	requireT.Equal(&nft.Class{
		Id:          assetnfttypes.BuildClassID(issueMsg.Symbol, issuer),
		Symbol:      issueMsg.Symbol,
		Name:        issueMsg.Name,
		Description: issueMsg.Description,
		Uri:         issueMsg.URI,
		UriHash:     issueMsg.URIHash,
	}, nftClassRes.Class)
}

// TestAssetNFTMint tests non-fungible token minting.
func TestAssetNFTMint(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()
	receiver := chain.GenAccount()

	nftClient := nft.NewQueryClient(chain.ClientContext)
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assetnfttypes.MsgIssueClass{},
				&assetnfttypes.MsgMint{},
				&nft.MsgSend{},
			},
		}),
	)

	// issue new NFT class
	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer: issuer.String(),
		Symbol: "NFTClassSymbol",
	}
	_, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	// mint new token in that class
	classID := assetnfttypes.BuildClassID(issueMsg.Symbol, issuer)
	mintMsg := &assetnfttypes.MsgMint{
		Sender:  issuer.String(),
		ID:      "id-1",
		ClassID: classID,
		URI:     "https://my-class-meta.invalid/1",
		URIHash: "content-hash",
	}
	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(mintMsg), uint64(res.GasUsed))

	nftMintedEvents, err := event.FindTypedEvents[*nft.EventMint](res.Events)
	requireT.NoError(err)
	nftMintedEvent := nftMintedEvents[0]
	requireT.Equal(&nft.EventMint{
		ClassId: classID,
		Id:      mintMsg.ID,
		Owner:   issuer.String(),
	}, nftMintedEvent)

	// check that token is present in the nft module
	nftRes, err := nftClient.NFT(ctx, &nft.QueryNFTRequest{
		ClassId: classID,
		Id:      nftMintedEvent.Id,
	})
	requireT.NoError(err)
	requireT.Equal(&nft.NFT{
		ClassId: classID,
		Id:      mintMsg.ID,
		Uri:     mintMsg.URI,
		UriHash: mintMsg.URIHash,
	}, nftRes.Nft)

	// check the owner
	ownerRes, err := nftClient.Owner(ctx, &nft.QueryOwnerRequest{
		ClassId: classID,
		Id:      nftMintedEvent.Id,
	})
	requireT.NoError(err)
	requireT.Equal(issuer.String(), ownerRes.Owner)

	// change the owner
	sendMsg := &nft.MsgSend{
		Sender:   issuer.String(),
		Receiver: receiver.String(),
		Id:       mintMsg.ID,
		ClassId:  classID,
	}
	res, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(sendMsg), uint64(res.GasUsed))
	nftSentEvents, err := event.FindTypedEvents[*nft.EventSend](res.Events)
	requireT.NoError(err)
	nftSentEvent := nftSentEvents[0]
	requireT.Equal(&nft.EventSend{
		Sender:   sendMsg.Sender,
		Receiver: sendMsg.Receiver,
		ClassId:  sendMsg.ClassId,
		Id:       sendMsg.Id,
	}, nftSentEvent)
	// check new owner
	ownerRes, err = nftClient.Owner(ctx, &nft.QueryOwnerRequest{
		ClassId: classID,
		Id:      nftMintedEvent.Id,
	})
	requireT.NoError(err)
	requireT.Equal(receiver.String(), ownerRes.Owner)
}
