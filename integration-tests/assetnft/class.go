package assetnft

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assetnfttypes "github.com/CoreumFoundation/coreum/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/x/nft"
)

// TestIssueClass tests non-fungible token class creation.
func TestIssueClass(ctx context.Context, t testing.T, chain testing.Chain) {
	requireT := require.New(t)
	issuer := chain.GenAccount()

	nftClient := nft.NewQueryClient(chain.ClientContext)
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, testing.BalancesOptions{
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
