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

// TestMintNonFungibleToken tests non-fungible token minting.
func TestMintNonFungibleToken(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	sender := chain.GenAccount()
	receiver := chain.GenAccount()

	nftClient := nft.NewQueryClient(chain.ClientContext)
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, sender, testing.BalancesOptions{
			Messages: []sdk.Msg{
				&assettypes.MsgCreateNonFungibleTokenClass{},
				&assettypes.MsgMintNonFungibleToken{},
				&nft.MsgSend{},
			},
		}),
	)

	// create new NFT class
	createMsg := &assettypes.MsgCreateNonFungibleTokenClass{
		Creator: sender.String(),
		Symbol:  "nftsymbol",
	}
	_, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(createMsg)),
		createMsg,
	)
	requireT.NoError(err)

	// mint new token in that class
	classID := assettypes.BuildNonFungibleTokenClassID(createMsg.Symbol, sender)
	mintMsg := &assettypes.MsgMintNonFungibleToken{
		Sender:  sender.String(),
		Id:      "id-1",
		ClassId: classID,
		Uri:     "https://my-class-meta.int/1",
		UriHash: "35b326a2b3b605270c26185c38d2581e937b2eae0418b4964ef521efe79cdf34",
	}
	res, err := tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(mintMsg), uint64(res.GasUsed))

	nftMintedEvt, err := event.FindTypedEvent[*nft.EventMint](res.Events)
	requireT.NoError(err)
	requireT.Equal(&nft.EventMint{
		ClassId: classID,
		Id:      "id-1",
		Owner:   sender.String(),
	}, nftMintedEvt)

	// check that token is present in the nft module
	nftRes, err := nftClient.NFT(ctx, &nft.QueryNFTRequest{
		ClassId: classID,
		Id:      nftMintedEvt.Id,
	})
	requireT.NoError(err)
	requireT.Equal(&nft.NFT{
		ClassId: classID,
		Id:      "id-1",
		Uri:     mintMsg.Uri,
		UriHash: mintMsg.UriHash,
	}, nftRes.Nft)

	// check the owner
	ownerRes, err := nftClient.Owner(ctx, &nft.QueryOwnerRequest{
		ClassId: classID,
		Id:      nftMintedEvt.Id,
	})
	requireT.NoError(err)
	requireT.Equal(sender.String(), ownerRes.Owner)

	// change the owner
	sendMsg := &nft.MsgSend{
		Sender:   sender.String(),
		Receiver: receiver.String(),
		Id:       "id-1",
		ClassId:  classID,
	}
	res, err = tx.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(sender),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(sendMsg), uint64(res.GasUsed))
	nftSentEvt, err := event.FindTypedEvent[*nft.EventSend](res.Events)
	requireT.NoError(err)
	requireT.Equal(&nft.EventSend{
		Sender:   sendMsg.Sender,
		Receiver: sendMsg.Receiver,
		ClassId:  sendMsg.ClassId,
		Id:       sendMsg.Id,
	}, nftSentEvt)
	// check new owner
	ownerRes, err = nftClient.Owner(ctx, &nft.QueryOwnerRequest{
		ClassId: classID,
		Id:      nftMintedEvt.Id,
	})
	requireT.NoError(err)
	requireT.Equal(receiver.String(), ownerRes.Owner)
}
