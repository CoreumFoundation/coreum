//go:build integrationtests

package modules

import (
	"bytes"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
	tmjson "github.com/tendermint/tendermint/libs/json"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
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

	assetNftClient := assetnfttypes.NewQueryClient(chain.ClientContext)

	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assetnfttypes.MsgIssueClass{},
			},
		}),
	)

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
			assetnfttypes.ClassFeature_burning, //nolint:nosnakecase // generated variable
		},
	}
	_, err = tx.BroadcastTx(
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
	_, err = tx.BroadcastTx(
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
			assetnfttypes.ClassFeature_burning, //nolint:nosnakecase // generated variable
		},
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
			assetnfttypes.ClassFeature_burning, //nolint:nosnakecase // generated variable
		},
	}, issuedEvent)

	// query nft asset with features
	assetNftClassRes, err := assetNftClient.Class(ctx, &assetnfttypes.QueryClassRequest{
		Id: classID,
	})
	requireT.NoError(err)

	requireT.Equal(assetnfttypes.Class{
		Id:          classID,
		Issuer:      issuer.String(),
		Symbol:      issueMsg.Symbol,
		Name:        issueMsg.Name,
		Description: issueMsg.Description,
		URI:         issueMsg.URI,
		URIHash:     issueMsg.URIHash,
		Data:        dataToCompare,
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_burning, //nolint:nosnakecase // generated variable
		},
	}, assetNftClassRes.Class)

	var data2 assetnfttypes.DataBytes
	requireT.NoError(proto.Unmarshal(assetNftClassRes.Class.Data.Value, &data2))

	requireT.Equal(jsonData, data2.Data)
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
			Amount: chain.NetworkConfig.AssetNFTConfig.MintFee,
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
	_, err = tx.BroadcastTx(
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
	_, err = tx.BroadcastTx(
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

	// check that balance is 0 meaning mint fee was taken

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: issuer.String(),
		Denom:   chain.NetworkConfig.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(chain.NewCoin(sdk.ZeroInt()).String(), resp.Balance.String())
}

// TestAssetNFTMintFeeProposal tests proposal upgrading mint fee
func TestAssetNFTMintFeeProposal(t *testing.T) {
	// This test can't be run together with other tests because it affects balances due to unexpected issue fee.
	// That's why t.Parallel() is not here.

	ctx, chain := integrationtests.NewTestingContext(t)
	requireT := require.New(t)
	origMintFee := chain.NetworkConfig.AssetNFTConfig.MintFee

	requireT.NoError(chain.Governance.UpdateParams(ctx, "Propose changing MintFee in the assetnft module",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(assetnfttypes.ModuleName, string(assetnfttypes.KeyMintFee), string(must.Bytes(tmjson.Marshal(sdk.NewCoin(chain.NetworkConfig.Denom, sdk.OneInt()))))),
		}))

	issuer := chain.GenAccount()
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assetnfttypes.MsgIssueClass{},
				&assetnfttypes.MsgMint{},
			},
			Amount: sdk.OneInt(),
		}))

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

	// verify issue fee was burnt

	burntStr, err := event.FindStringEventAttribute(res.Events, banktypes.EventTypeCoinBurn, sdk.AttributeKeyAmount)
	requireT.NoError(err)
	requireT.Equal(sdk.NewCoin(chain.NetworkConfig.Denom, sdk.OneInt()).String(), burntStr)

	// check that balance is 0 meaning mint fee was taken

	bankClient := banktypes.NewQueryClient(chain.ClientContext)
	resp, err := bankClient.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: issuer.String(),
		Denom:   chain.NetworkConfig.Denom,
	})
	requireT.NoError(err)
	requireT.Equal(chain.NewCoin(sdk.ZeroInt()).String(), resp.Balance.String())

	// Revert to original mint fee
	requireT.NoError(chain.Governance.UpdateParams(ctx, "Propose changing MintFee in the assetnft module",
		[]paramproposal.ParamChange{
			paramproposal.NewParamChange(assetnfttypes.ModuleName, string(assetnfttypes.KeyMintFee), string(must.Bytes(tmjson.Marshal(sdk.NewCoin(chain.NetworkConfig.Denom, origMintFee))))),
		}))
}

// TestAssetNFTBurn tests non-fungible token burning.
func TestAssetNFTBurn(t *testing.T) {
	t.Parallel()

	ctx, chain := integrationtests.NewTestingContext(t)

	requireT := require.New(t)
	issuer := chain.GenAccount()

	nftClient := nft.NewQueryClient(chain.ClientContext)
	requireT.NoError(
		chain.Faucet.FundAccountsWithOptions(ctx, issuer, integrationtests.BalancesOptions{
			Messages: []sdk.Msg{
				&assetnfttypes.MsgIssueClass{},
				&assetnfttypes.MsgMint{},
				&assetnfttypes.MsgBurn{},
			},
		}),
	)

	// issue new NFT class
	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer: issuer.String(),
		Symbol: "NFTClassSymbol",
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_burning, //nolint:nosnakecase // generated variable
		},
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
	}
	res, err := tx.BroadcastTx(
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
	res, err = tx.BroadcastTx(
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
}
