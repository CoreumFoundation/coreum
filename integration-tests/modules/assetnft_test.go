//go:build integrationtests

package modules

import (
	"bytes"
	"context"
	"testing"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	integrationtests "github.com/CoreumFoundation/coreum/integration-tests"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/testutil/event"
	assetnfttypes "github.com/CoreumFoundation/coreum/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/x/nft"
)

// TestAssetNFTQueryParams queries parameters of asset/nft module.
func TestAssetNFTQueryParams(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	mintFee := getMintFee(ctx, t, chain.ClientContext)

	assert.Equal(t, chain.ChainSettings.Denom, mintFee.Denom)
}

// TestAssetNFTIssueClass tests non-fungible token class creation.
//
//nolint:funlen // there are many tests
func TestAssetNFTIssueClass(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()

	assetNftClient := assetnfttypes.NewQueryClient(chain.ClientContext)

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
		},
	})

	// issue new NFT class with invalid data type

	data, err := codectypes.NewAnyWithValue(&assetnfttypes.MsgMint{})
	requireT.NoError(err)

	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer:      issuer.String(),
		Symbol:      "symbol",
		Name:        "name",
		Description: "description",
		URI:         "https://my-class-meta.invalid/1",
		URIHash:     "content-hash",
		Data:        data,
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_burning,
		},
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.True(assetnfttypes.ErrInvalidInput.Is(err))

	// issue new NFT class with too long data

	data, err = codectypes.NewAnyWithValue(&assetnfttypes.DataBytes{Data: bytes.Repeat([]byte{0x01}, assetnfttypes.MaxDataSize+1)})
	requireT.NoError(err)

	issueMsg = &assetnfttypes.MsgIssueClass{
		Issuer:      issuer.String(),
		Symbol:      "symbol",
		Name:        "name",
		Description: "description",
		URI:         "https://my-class-meta.invalid/1",
		URIHash:     "content-hash",
		Data:        data,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.True(assetnfttypes.ErrInvalidInput.Is(err))

	jsonData := []byte(`{"name": "Name", "description": "Description"}`)

	// issue new NFT class
	data, err = codectypes.NewAnyWithValue(&assetnfttypes.DataBytes{Data: jsonData})
	requireT.NoError(err)

	// we need to do this, otherwise assertion fails because some private fields are set differently
	dataToCompare := &codectypes.Any{
		TypeUrl: data.TypeUrl,
		Value:   data.Value,
	}

	issueMsg = &assetnfttypes.MsgIssueClass{
		Issuer:      issuer.String(),
		Symbol:      "symbol",
		Name:        "name",
		Description: "description",
		URI:         "https://my-class-meta.invalid/1",
		URIHash:     "content-hash",
		Data:        data,
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_burning,
			assetnfttypes.ClassFeature_disable_sending,
		},
		RoyaltyRate: sdk.MustNewDecFromStr("0.1"),
	}
	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(issueMsg), uint64(res.GasUsed))
	tokenIssuedEvents, err := event.FindTypedEvents[*assetnfttypes.EventClassIssued](res.Events)
	requireT.NoError(err)
	issuedEvent := tokenIssuedEvents[0]

	classID := assetnfttypes.BuildClassID(issueMsg.Symbol, issuer)
	requireT.Equal(&assetnfttypes.EventClassIssued{
		ID:          classID,
		Issuer:      issuer.String(),
		Symbol:      issueMsg.Symbol,
		Name:        issueMsg.Name,
		Description: issueMsg.Description,
		URI:         issueMsg.URI,
		URIHash:     issueMsg.URIHash,
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_burning,
			assetnfttypes.ClassFeature_disable_sending,
		},
		RoyaltyRate: sdk.MustNewDecFromStr("0.1"),
	}, issuedEvent)

	// query nft asset with features
	assetNftClassRes, err := assetNftClient.Class(ctx, &assetnfttypes.QueryClassRequest{
		Id: classID,
	})
	requireT.NoError(err)

	expectedClass := assetnfttypes.Class{
		Id:          classID,
		Issuer:      issuer.String(),
		Symbol:      issueMsg.Symbol,
		Name:        issueMsg.Name,
		Description: issueMsg.Description,
		URI:         issueMsg.URI,
		URIHash:     issueMsg.URIHash,
		Data:        dataToCompare,
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_burning,
			assetnfttypes.ClassFeature_disable_sending,
		},
		RoyaltyRate: sdk.MustNewDecFromStr("0.1"),
	}
	requireT.Equal(expectedClass, assetNftClassRes.Class)

	var data2 assetnfttypes.DataBytes
	requireT.NoError(proto.Unmarshal(assetNftClassRes.Class.Data.Value, &data2))

	requireT.Equal(jsonData, data2.Data)

	assetNftClassesRes, err := assetNftClient.Classes(ctx, &assetnfttypes.QueryClassesRequest{
		Issuer: issuer.String(),
	})
	requireT.NoError(err)
	requireT.Equal(1, len(assetNftClassesRes.Classes))
	requireT.Equal(uint64(1), assetNftClassesRes.Pagination.Total)
	requireT.Equal(expectedClass, assetNftClassesRes.Classes[0])
}

// TestAssetNFTMint tests non-fungible token minting.
//
//nolint:funlen // there are many tests
func TestAssetNFTMint(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()
	recipient := chain.GenAccount()

	nftClient := nft.NewQueryClient(chain.ClientContext)
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
			&assetnfttypes.MsgMint{},
			&nft.MsgSend{},
		},
		Amount: getMintFee(ctx, t, chain.ClientContext).Amount,
	})

	// issue new NFT class
	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer: issuer.String(),
		Symbol: "NFTClassSymbol",
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	classID := assetnfttypes.BuildClassID(issueMsg.Symbol, issuer)

	// mint with invalid data type

	data, err := codectypes.NewAnyWithValue(&assetnfttypes.MsgMint{})
	requireT.NoError(err)

	mintMsg := &assetnfttypes.MsgMint{
		Sender:  issuer.String(),
		ID:      "id-1",
		ClassID: classID,
		URI:     "https://my-class-meta.invalid/1",
		URIHash: "content-hash",
		Data:    data,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.True(assetnfttypes.ErrInvalidInput.Is(err))

	// mint with too long data

	data, err = codectypes.NewAnyWithValue(&assetnfttypes.DataBytes{Data: bytes.Repeat([]byte{0x01}, assetnfttypes.MaxDataSize+1)})
	requireT.NoError(err)

	mintMsg = &assetnfttypes.MsgMint{
		Sender:  issuer.String(),
		ID:      "id-1",
		ClassID: classID,
		URI:     "https://my-class-meta.invalid/1",
		URIHash: "content-hash",
		Data:    data,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.True(assetnfttypes.ErrInvalidInput.Is(err))

	jsonData := []byte(`{"name": "Name", "description": "Description"}`)

	// mint new token in that class
	data, err = codectypes.NewAnyWithValue(&assetnfttypes.DataBytes{Data: jsonData})
	requireT.NoError(err)

	// we need to do this, otherwise assertion fails because some private fields are set differently
	dataToCompare := &codectypes.Any{
		TypeUrl: data.TypeUrl,
		Value:   data.Value,
	}

	mintMsg = &assetnfttypes.MsgMint{
		Sender:  issuer.String(),
		ID:      "id-1",
		ClassID: classID,
		URI:     "https://my-class-meta.invalid/1",
		URIHash: "content-hash",
		Data:    data,
	}
	res, err := client.BroadcastTx(
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
		Data:    dataToCompare,
	}, nftRes.Nft)

	var data2 assetnfttypes.DataBytes
	requireT.NoError(proto.Unmarshal(nftRes.Nft.Data.Value, &data2))

	requireT.Equal(jsonData, data2.Data)

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
		Receiver: recipient.String(),
		Id:       mintMsg.ID,
		ClassId:  classID,
	}
	res, err = client.BroadcastTx(
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
	requireT.Equal(recipient.String(), ownerRes.Owner)

	// check that balance is 0 meaning mint fee was taken

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: issuer.String(),
		Denom:   chain.ChainSettings.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(chain.NewCoin(sdk.ZeroInt()).String(), resp.Balance.String())
}

// TestAssetNFTMintFeeProposal tests proposal upgrading mint fee.
func TestAssetNFTMintFeeProposal(t *testing.T) {
	// This test can't be run together with other tests because it affects balances due to unexpected issue fee.
	// That's why t.Parallel() is not here.

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)
	origMintFee := getMintFee(ctx, t, chain.ClientContext)

	chain.Governance.UpdateParams(ctx, t, "Propose changing MintFee in the assetnft module",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(assetnfttypes.ModuleName, string(assetnfttypes.KeyMintFee), string(must.Bytes(tmjson.Marshal(sdk.NewCoin(origMintFee.Denom, sdk.OneInt()))))),
		})

	issuer := chain.GenAccount()
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
			&assetnfttypes.MsgMint{},
		},
		Amount: sdk.OneInt(),
	})

	// issue new NFT class
	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer: issuer.String(),
		Symbol: "NFTClassSymbol",
	}
	_, err := client.BroadcastTx(
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
	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.NoError(err)

	// verify issue fee was burnt

	burntStr, err := event.FindStringEventAttribute(res.Events, banktypes.EventTypeCoinBurn, sdk.AttributeKeyAmount)
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(chain.ChainSettings.Denom, sdk.OneInt()).String(), burntStr)

	// check that balance is 0 meaning mint fee was taken

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: issuer.String(),
		Denom:   chain.ChainSettings.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(chain.NewCoin(sdk.ZeroInt()).String(), resp.Balance.String())

	// Revert to original mint fee
	chain.Governance.UpdateParams(ctx, t, "Propose changing MintFee in the assetnft module",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(assetnfttypes.ModuleName, string(assetnfttypes.KeyMintFee), string(must.Bytes(tmjson.Marshal(origMintFee)))),
		})
}

// TestAssetNFTBurn tests non-fungible token burning.
//
//nolint:funlen // there are many tests
func TestAssetNFTBurn(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()

	nftClient := nft.NewQueryClient(chain.ClientContext)
	assetnftClient := assetnfttypes.NewQueryClient(chain.ClientContext)
	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
			&assetnfttypes.MsgMint{},
			&assetnfttypes.MsgBurn{},
			&assetnfttypes.MsgBurn{},
			&assetnfttypes.MsgMint{},
			&assetnfttypes.MsgMint{},
		},
	})

	// issue new NFT class
	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer: issuer.String(),
		Symbol: "NFTClassSymbol",
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_burning,
		},
	}
	_, err := client.BroadcastTx(
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
	}
	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(mintMsg), uint64(res.GasUsed))

	// check that token is present in the nft module
	nftRes, err := nftClient.NFT(ctx, &nft.QueryNFTRequest{
		ClassId: classID,
		Id:      mintMsg.ID,
	})
	requireT.NoError(err)
	requireT.Equal(&nft.NFT{
		ClassId: classID,
		Id:      mintMsg.ID,
		Uri:     mintMsg.URI,
		UriHash: mintMsg.URIHash,
	}, nftRes.Nft)

	// burn the NFT
	msgBurn := &assetnfttypes.MsgBurn{
		Sender:  issuer.String(),
		ClassID: classID,
		ID:      "id-1",
	}
	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgBurn)),
		msgBurn,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(msgBurn), uint64(res.GasUsed))

	// assert the burning event
	burnEvents, err := event.FindTypedEvents[*nft.EventBurn](res.Events)
	requireT.NoError(err)
	burnEvent := burnEvents[0]
	requireT.Equal(&nft.EventBurn{
		ClassId: classID,
		Id:      msgBurn.ID,
		Owner:   issuer.String(),
	}, burnEvent)

	// check that token isn't presented in the nft module anymore
	_, err = nftClient.NFT(ctx, &nft.QueryNFTRequest{
		ClassId: classID,
		Id:      mintMsg.ID,
	})
	requireT.Error(err)
	requireT.Contains(err.Error(), nft.ErrNFTNotExists.Error()) // the nft wraps the errors with the `errors` so the client doesn't decode them as sdk errors.

	// try to mint token with the same ID, should fail
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.Error(err)
	requireT.ErrorIs(err, assetnfttypes.ErrInvalidInput)

	// mint token with different ID, should succeed
	mintMsg.ID += "-2"
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.NoError(err)

	// burn the second NFT
	msgBurn = &assetnfttypes.MsgBurn{
		Sender:  issuer.String(),
		ClassID: classID,
		ID:      "id-1-2",
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgBurn)),
		msgBurn,
	)
	requireT.NoError(err)

	// query burnt NFTs
	burntListRes, err := assetnftClient.BurntNFTsInClass(ctx, &assetnfttypes.QueryBurntNFTsInClassRequest{
		Pagination: &query.PageRequest{},
		ClassId:    mintMsg.ClassID,
	})
	requireT.NoError(err)
	requireT.Len(burntListRes.NftIds, 2)

	// test pagination works
	burntListRes, err = assetnftClient.BurntNFTsInClass(ctx, &assetnfttypes.QueryBurntNFTsInClassRequest{
		Pagination: &query.PageRequest{
			Offset: 1,
		},
		ClassId: mintMsg.ClassID,
	})
	requireT.NoError(err)
	requireT.Len(burntListRes.NftIds, 1)

	// query is nft burnt
	burntNft, err := assetnftClient.BurntNFT(ctx, &assetnfttypes.QueryBurntNFTRequest{
		ClassId: mintMsg.ClassID,
		NftId:   "id-1",
	})
	requireT.NoError(err)
	requireT.True(burntNft.Burnt)
}

// TestAssetNFTBurnFrozen tests that frozen NFT cannot be burnt.
//
//nolint:funlen // there are many tests
func TestAssetNFTBurnFrozen(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()
	recipient1 := chain.GenAccount()
	assetNFTClient := assetnfttypes.NewQueryClient(chain.ClientContext)

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
			&assetnfttypes.MsgMint{},
			&nft.MsgSend{},
			&assetnfttypes.MsgFreeze{},
			&assetnfttypes.MsgUnfreeze{},
		},
		Amount: getMintFee(ctx, t, chain.ClientContext).Amount,
	})

	chain.FundAccountsWithOptions(ctx, t, recipient1, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgBurn{},
			&assetnfttypes.MsgBurn{},
		},
	})

	// issue new NFT class
	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer: issuer.String(),
		Symbol: "NFTClassSymbol",
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_freezing,
			assetnfttypes.ClassFeature_burning,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	// mint new token in that class
	classID := assetnfttypes.BuildClassID(issueMsg.Symbol, issuer)
	nftID := "id-1" //nolint:goconst // repeating values in tests are ok
	mintMsg := &assetnfttypes.MsgMint{
		Sender:  issuer.String(),
		ID:      nftID,
		ClassID: classID,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.NoError(err)

	// freeze the NFT
	msgFreeze := &assetnfttypes.MsgFreeze{
		Sender:  issuer.String(),
		ClassID: classID,
		ID:      nftID,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgFreeze)),
		msgFreeze,
	)
	requireT.NoError(err)

	queryRes, err := assetNFTClient.Frozen(ctx, &assetnfttypes.QueryFrozenRequest{
		ClassId: classID,
		Id:      nftID,
	})
	requireT.NoError(err)
	requireT.True(queryRes.Frozen)

	// send from issuer to recipient1 (send is allowed)
	sendMsg := &nft.MsgSend{
		Sender:   issuer.String(),
		ClassId:  classID,
		Id:       nftID,
		Receiver: recipient1.String(),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// burn from recipient1 (burn is not allowed since it is frozen)
	burnMsg := &assetnfttypes.MsgBurn{
		Sender:  recipient1.String(),
		ClassID: classID,
		ID:      nftID,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.ErrorIs(err, sdkerrors.ErrUnauthorized)

	// unfreeze the nft
	msgUnFreeze := &assetnfttypes.MsgUnfreeze{
		Sender:  issuer.String(),
		ClassID: classID,
		ID:      nftID,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgUnFreeze)),
		msgUnFreeze,
	)
	requireT.NoError(err)

	// burn from recipient1 (it is now allowed)
	burnMsg = &assetnfttypes.MsgBurn{
		Sender:  recipient1.String(),
		ClassID: classID,
		ID:      nftID,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.NoError(err)
}

// TestAssetNFTBurnFrozen_Issuer tests that frozen NFT can be burnt by the issuer.
func TestAssetNFTBurnFrozen_Issuer(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()
	assetNFTClient := assetnfttypes.NewQueryClient(chain.ClientContext)
	nftClient := nft.NewQueryClient(chain.ClientContext)

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
			&assetnfttypes.MsgMint{},
			&assetnfttypes.MsgFreeze{},
			&assetnfttypes.MsgBurn{},
		},
		Amount: getMintFee(ctx, t, chain.ClientContext).Amount,
	})

	// issue new NFT class
	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer: issuer.String(),
		Symbol: "NFTClassSymbol",
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_freezing,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	// mint new token in that class
	classID := assetnfttypes.BuildClassID(issueMsg.Symbol, issuer)
	nftID := "id-1"
	mintMsg := &assetnfttypes.MsgMint{
		Sender:  issuer.String(),
		ID:      nftID,
		ClassID: classID,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.NoError(err)

	// freeze the NFT
	msgFreeze := &assetnfttypes.MsgFreeze{
		Sender:  issuer.String(),
		ClassID: classID,
		ID:      nftID,
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgFreeze)),
		msgFreeze,
	)
	requireT.NoError(err)

	queryRes, err := assetNFTClient.Frozen(ctx, &assetnfttypes.QueryFrozenRequest{
		ClassId: classID,
		Id:      nftID,
	})
	requireT.NoError(err)
	requireT.True(queryRes.Frozen)

	_, err = nftClient.NFT(ctx, &nft.QueryNFTRequest{
		ClassId: classID,
		Id:      nftID,
	})
	requireT.NoError(err)

	// burn from issuer (burn is allowed even though nft is frozen)
	burnMsg := &assetnfttypes.MsgBurn{
		Sender:  issuer.String(),
		ClassID: classID,
		ID:      nftID,
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(burnMsg)),
		burnMsg,
	)
	requireT.NoError(err)

	_, err = nftClient.NFT(ctx, &nft.QueryNFTRequest{
		ClassId: classID,
		Id:      nftID,
	})
	requireT.Error(err)
	requireT.Contains(err.Error(), "not found nft")
}

// TestAssetNFTFreeze tests non-fungible token freezing.
//
//nolint:funlen // there are many tests
func TestAssetNFTFreeze(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()
	recipient1 := chain.GenAccount()
	nftClient := assetnfttypes.NewQueryClient(chain.ClientContext)

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
			&assetnfttypes.MsgMint{},
			&nft.MsgSend{},
			&assetnfttypes.MsgFreeze{},
			&assetnfttypes.MsgUnfreeze{},
		},
		Amount: getMintFee(ctx, t, chain.ClientContext).Amount,
	})

	chain.FundAccountsWithOptions(ctx, t, recipient1, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&nft.MsgSend{},
			&nft.MsgSend{},
			&nft.MsgSend{},
		},
	})

	// issue new NFT class
	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer: issuer.String(),
		Symbol: "NFTClassSymbol",
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_freezing,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	// mint new token in that class
	classID := assetnfttypes.BuildClassID(issueMsg.Symbol, issuer)
	nftID := "id-1"
	mintMsg := &assetnfttypes.MsgMint{
		Sender:  issuer.String(),
		ID:      nftID,
		ClassID: classID,
	}
	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(mintMsg), uint64(res.GasUsed))

	// freeze the NFT
	msgFreeze := &assetnfttypes.MsgFreeze{
		Sender:  issuer.String(),
		ClassID: classID,
		ID:      nftID,
	}
	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgFreeze)),
		msgFreeze,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(msgFreeze), uint64(res.GasUsed))

	queryRes, err := nftClient.Frozen(ctx, &assetnfttypes.QueryFrozenRequest{
		ClassId: classID,
		Id:      nftID,
	})
	requireT.NoError(err)
	requireT.True(queryRes.Frozen)

	// assert the freezing event
	frozenEvents, err := event.FindTypedEvents[*assetnfttypes.EventFrozen](res.Events)
	requireT.NoError(err)
	frozenEvent := frozenEvents[0]
	requireT.Equal(&assetnfttypes.EventFrozen{
		ClassId: classID,
		Id:      msgFreeze.ID,
		Owner:   issuer.String(),
	}, frozenEvent)

	// send from issuer to recipient1 (send is allowed)
	sendMsg := &nft.MsgSend{
		Sender:   issuer.String(),
		ClassId:  classID,
		Id:       nftID,
		Receiver: recipient1.String(),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// send from recipient1 to recipient2 (send is not allowed since it is frozen)
	recipient2 := chain.GenAccount()
	sendMsg = &nft.MsgSend{
		Sender:   recipient1.String(),
		ClassId:  classID,
		Id:       nftID,
		Receiver: recipient2.String(),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.Error(err)
	requireT.True(sdkerrors.ErrUnauthorized.Is(err))

	// send from recipient1 to issuer (send is not allowed since it is frozen)
	sendMsg = &nft.MsgSend{
		Sender:   recipient1.String(),
		ClassId:  classID,
		Id:       nftID,
		Receiver: issuer.String(),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.Error(err)
	requireT.True(sdkerrors.ErrUnauthorized.Is(err))

	// unfreeze the NFT
	msgUnfreeze := &assetnfttypes.MsgUnfreeze{
		Sender:  issuer.String(),
		ClassID: classID,
		ID:      nftID,
	}
	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgUnfreeze)),
		msgUnfreeze,
	)
	requireT.NoError(err)
	requireT.EqualValues(chain.GasLimitByMsgs(msgUnfreeze), res.GasUsed)

	queryRes, err = nftClient.Frozen(ctx, &assetnfttypes.QueryFrozenRequest{
		ClassId: classID,
		Id:      nftID,
	})
	requireT.NoError(err)
	requireT.False(queryRes.Frozen)

	// assert the unfreezing event
	unFrozenEvents, err := event.FindTypedEvents[*assetnfttypes.EventUnfrozen](res.Events)
	requireT.NoError(err)
	unfrozenEvent := unFrozenEvents[0]
	requireT.Equal(&assetnfttypes.EventUnfrozen{
		ClassId: classID,
		Id:      msgFreeze.ID,
		Owner:   recipient1.String(),
	}, unfrozenEvent)

	// send from recipient1 to recipient2 (send is allowed since it is not frozen)
	sendMsg = &nft.MsgSend{
		Sender:   recipient1.String(),
		ClassId:  classID,
		Id:       nftID,
		Receiver: recipient2.String(),
	}

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient1),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
}

// TestAssetNFTWhitelist tests non-fungible token whitelisting.
//
//nolint:funlen // there are many tests
func TestAssetNFTWhitelist(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()
	recipient := chain.GenAccount()
	nftClient := assetnfttypes.NewQueryClient(chain.ClientContext)

	chain.FundAccountsWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
			&assetnfttypes.MsgMint{},
			&nft.MsgSend{},
			&assetnfttypes.MsgAddToWhitelist{},
			&nft.MsgSend{},
			&assetnfttypes.MsgAddToWhitelist{},
			&assetnfttypes.MsgRemoveFromWhitelist{},
			&assetnfttypes.MsgAddToWhitelist{},
		},
		Amount: getMintFee(ctx, t, chain.ClientContext).Amount,
	})

	// issue new NFT class
	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer: issuer.String(),
		Symbol: "NFTClassSymbol",
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_whitelisting,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	// mint new token in that class
	classID := assetnfttypes.BuildClassID(issueMsg.Symbol, issuer)
	nftID := "id-1"
	mintMsg := &assetnfttypes.MsgMint{
		Sender:  issuer.String(),
		ID:      nftID,
		ClassID: classID,
	}
	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(mintMsg)),
		mintMsg,
	)
	requireT.NoError(err)
	requireT.Equal(chain.GasLimitByMsgs(mintMsg), uint64(res.GasUsed))

	// send to non-whitelisted recipient (send must fail)
	sendMsg := &nft.MsgSend{
		Sender:   issuer.String(),
		ClassId:  classID,
		Id:       nftID,
		Receiver: recipient.String(),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.Error(err)
	requireT.ErrorIs(err, sdkerrors.ErrUnauthorized)

	// whitelist recipient for the NFT
	msgAddToWhitelist := &assetnfttypes.MsgAddToWhitelist{
		Sender:  issuer.String(),
		ClassID: classID,
		ID:      nftID,
		Account: recipient.String(),
	}
	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgAddToWhitelist)),
		msgAddToWhitelist,
	)
	requireT.NoError(err)
	requireT.EqualValues(chain.GasLimitByMsgs(msgAddToWhitelist), res.GasUsed)

	// assert the query
	queryRes, err := nftClient.Whitelisted(ctx, &assetnfttypes.QueryWhitelistedRequest{
		ClassId: classID,
		Id:      nftID,
		Account: recipient.String(),
	})
	requireT.NoError(err)
	requireT.True(queryRes.Whitelisted)

	// assert the whitelisting event
	whitelistEvents, err := event.FindTypedEvents[*assetnfttypes.EventAddedToWhitelist](res.Events)
	requireT.NoError(err)
	whitelistEvent := whitelistEvents[0]
	requireT.Equal(&assetnfttypes.EventAddedToWhitelist{
		ClassId: classID,
		Id:      msgAddToWhitelist.ID,
		Account: recipient.String(),
	}, whitelistEvent)

	// try to send again and it should succeed now.
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// send from whitelisted recipient to non-whitelisted recipient2 (send must fail)
	chain.FundAccountsWithOptions(ctx, t, recipient, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&nft.MsgSend{},
			&nft.MsgSend{},
		},
	})
	recipient2 := chain.GenAccount()
	chain.FundAccountsWithOptions(ctx, t, recipient2, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&nft.MsgSend{},
			&nft.MsgSend{},
		},
	})

	sendMsg = &nft.MsgSend{
		Sender:   recipient.String(),
		ClassId:  classID,
		Id:       nftID,
		Receiver: recipient2.String(),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.Error(err)
	requireT.ErrorIs(err, sdkerrors.ErrUnauthorized)

	// whitelist recipient2 for the NFT
	msgAddToWhitelist = &assetnfttypes.MsgAddToWhitelist{
		Sender:  issuer.String(),
		ClassID: classID,
		ID:      nftID,
		Account: recipient2.String(),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgAddToWhitelist)),
		msgAddToWhitelist,
	)
	requireT.NoError(err)

	// try to send again from recipient to recipient2 and it should succeed now.
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)

	// unwhitelist the account
	msgRemoveFromWhitelist := &assetnfttypes.MsgRemoveFromWhitelist{
		Sender:  issuer.String(),
		ClassID: classID,
		ID:      nftID,
		Account: recipient.String(),
	}
	res, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgRemoveFromWhitelist)),
		msgRemoveFromWhitelist,
	)
	requireT.NoError(err)
	requireT.EqualValues(chain.GasLimitByMsgs(msgRemoveFromWhitelist), res.GasUsed)

	queryRes, err = nftClient.Whitelisted(ctx, &assetnfttypes.QueryWhitelistedRequest{
		ClassId: classID,
		Id:      nftID,
		Account: recipient.String(),
	})
	requireT.NoError(err)
	requireT.False(queryRes.Whitelisted)

	// assert the unwhitelisting event
	unWhitelistedEvents, err := event.FindTypedEvents[*assetnfttypes.EventRemovedFromWhitelist](res.Events)
	requireT.NoError(err)
	unWhitelistedEvent := unWhitelistedEvents[0]
	requireT.Equal(&assetnfttypes.EventRemovedFromWhitelist{
		ClassId: classID,
		Id:      msgAddToWhitelist.ID,
		Account: recipient.String(),
	}, unWhitelistedEvent)

	// try to send back from recipient2 to non-whitelisted recipient (send should fail)
	sendMsg = &nft.MsgSend{
		Sender:   recipient2.String(),
		ClassId:  classID,
		Id:       nftID,
		Receiver: recipient.String(),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient2),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.Error(err)
	requireT.ErrorIs(err, sdkerrors.ErrUnauthorized)

	// whitelisting issuer should fail
	msgAddToWhitelist = &assetnfttypes.MsgAddToWhitelist{
		Sender:  issuer.String(),
		ClassID: classID,
		ID:      nftID,
		Account: issuer.String(),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgAddToWhitelist)),
		msgAddToWhitelist,
	)
	requireT.Error(err)
	requireT.ErrorIs(err, sdkerrors.ErrUnauthorized)

	// sending to issuer should succeed
	sendMsg = &nft.MsgSend{
		Sender:   recipient2.String(),
		ClassId:  classID,
		Id:       nftID,
		Receiver: issuer.String(),
	}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(recipient2),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(sendMsg)),
		sendMsg,
	)
	requireT.NoError(err)
}

// TestAssetNFTAuthZ tests that assetft module works seamlessly with authz module.
func TestAssetNFTAuthZ(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	requireT := require.New(t)
	granter := chain.GenAccount()
	grantee := chain.GenAccount()
	nftClient := assetnfttypes.NewQueryClient(chain.ClientContext)

	chain.FundAccountsWithOptions(ctx, t, granter, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
			&assetnfttypes.MsgMint{},
			&authztypes.MsgGrant{},
		},
		Amount: getMintFee(ctx, t, chain.ClientContext).Amount,
	})

	// issue new NFT class
	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer: granter.String(),
		Symbol: "NFTClassSymbol",
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_freezing, //nolint:nosnakecase // generated variable
		},
	}

	// mint new token in that class
	classID := assetnfttypes.BuildClassID(issueMsg.Symbol, granter)
	nftID := "id-1"
	mintMsg := &assetnfttypes.MsgMint{
		Sender:  granter.String(),
		ID:      nftID,
		ClassID: classID,
	}

	// grant authorization to freeze nft
	grantMsg, err := authztypes.NewMsgGrant(
		granter,
		grantee,
		authztypes.NewGenericAuthorization(sdk.MsgTypeURL(&assetnfttypes.MsgFreeze{})),
		time.Now().Add(time.Minute),
	)
	requireT.NoError(err)

	msgList := []sdk.Msg{issueMsg, mintMsg, grantMsg}
	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(granter),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(msgList...)),
		msgList...,
	)
	requireT.NoError(err)

	// whitelist recipient for the NFT
	freezeMsg := &assetnfttypes.MsgFreeze{
		Sender:  granter.String(),
		ClassID: classID,
		ID:      nftID,
	}
	execMsg := authztypes.NewMsgExec(grantee, []sdk.Msg{freezeMsg})

	chain.FundAccountsWithOptions(ctx, t, grantee, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&execMsg,
		},
		Amount: getMintFee(ctx, t, chain.ClientContext).Amount,
	})

	_, err = client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(grantee),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(&execMsg)),
		&execMsg,
	)
	requireT.NoError(err)

	// assert the query
	queryRes, err := nftClient.Frozen(ctx, &assetnfttypes.QueryFrozenRequest{
		ClassId: classID,
		Id:      nftID,
	})
	requireT.NoError(err)
	requireT.True(queryRes.Frozen)
}

func getMintFee(ctx context.Context, t *testing.T, clientCtx client.Context) sdk.Coin {
	queryClient := assetnfttypes.NewQueryClient(clientCtx)
	resp, err := queryClient.Params(ctx, &assetnfttypes.QueryParamsRequest{})
	require.NoError(t, err)

	return resp.Params.MintFee
}
