//go:build integrationtests

package upgrade

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	integrationtests "github.com/CoreumFoundation/coreum/v2/integration-tests"
	moduleswasm "github.com/CoreumFoundation/coreum/v2/integration-tests/contracts/modules"
	"github.com/CoreumFoundation/coreum/v2/pkg/client"
	"github.com/CoreumFoundation/coreum/v2/testutil/event"
	assetnfttypes "github.com/CoreumFoundation/coreum/v2/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/v2/x/nft"
)

type nftStoreTest struct {
	issuer        sdk.AccAddress
	issuedEvent   *assetnfttypes.EventClassIssued
	expectedClass assetnfttypes.Class
	mintMsg       *assetnfttypes.MsgMint
	expectedNFT   nft.NFT
}

func (n *nftStoreTest) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	// create NFT class and mint NFT to check the keys migration
	n.issuer = chain.GenAccount()
	assetNftClient := assetnfttypes.NewQueryClient(chain.ClientContext)
	nfqQueryClient := nft.NewQueryClient(chain.ClientContext)
	chain.FundAccountWithOptions(ctx, t, n.issuer, integrationtests.BalancesOptions{
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

func (n *nftStoreTest) After(t *testing.T) {
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

type nftFeaturesTest struct {
	classID string
}

func (nt *nftFeaturesTest) Before(t *testing.T) {
	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	issuer := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
		},
	})

	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer:      issuer.String(),
		Symbol:      "symbol",
		Name:        "name",
		Description: "description",
		URI:         "https://my-class-meta.invalid/1",
		URIHash:     "content-hash",
		RoyaltyRate: sdk.ZeroDec(),
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_burning,
			assetnfttypes.ClassFeature_freezing,
			2000, // must be removed as a result of migration
			assetnfttypes.ClassFeature_whitelisting,
			3000, // must be removed as a result of migration
			assetnfttypes.ClassFeature_disable_sending,
			assetnfttypes.ClassFeature_burning,         // must be removed as a result of migration
			assetnfttypes.ClassFeature_freezing,        // must be removed as a result of migration
			2000,                                       // must be removed as a result of migration
			assetnfttypes.ClassFeature_whitelisting,    // must be removed as a result of migration
			3000,                                       // must be removed as a result of migration
			assetnfttypes.ClassFeature_disable_sending, // must be removed as a result of migration
		},
	}
	res, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)

	tokenIssuedEvents, err := event.FindTypedEvents[*assetnfttypes.EventClassIssued](res.Events)
	requireT.NoError(err)
	nt.classID = tokenIssuedEvents[0].ID
}

func (nt *nftFeaturesTest) After(t *testing.T) {
	nt.verifyClassIsFixed(t)
	nt.tryCreatingClassWithInvalidFeature(t)
	nt.tryCreatingClassWithDuplicatedFeature(t)
	nt.createValidClass(t)
}

func (nt *nftFeaturesTest) verifyClassIsFixed(t *testing.T) {
	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)

	nftClient := assetnfttypes.NewQueryClient(chain.ClientContext)
	resp, err := nftClient.Class(ctx, &assetnfttypes.QueryClassRequest{
		Id: nt.classID,
	})
	requireT.NoError(err)

	requireT.Equal([]assetnfttypes.ClassFeature{
		assetnfttypes.ClassFeature_burning,
		assetnfttypes.ClassFeature_freezing,
		assetnfttypes.ClassFeature_whitelisting,
		assetnfttypes.ClassFeature_disable_sending,
	}, resp.Class.Features)
}

func (nt *nftFeaturesTest) tryCreatingClassWithInvalidFeature(t *testing.T) {
	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	issuer := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
		},
	})

	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer:      issuer.String(),
		Symbol:      "symbol",
		Name:        "name",
		Description: "description",
		URI:         "https://my-class-meta.invalid/1",
		URIHash:     "content-hash",
		RoyaltyRate: sdk.ZeroDec(),
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_burning,
			assetnfttypes.ClassFeature_freezing,
			2000,
			assetnfttypes.ClassFeature_whitelisting,
			assetnfttypes.ClassFeature_disable_sending,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.ErrorContains(err, "non-existing class feature provided")
}

func (nt *nftFeaturesTest) tryCreatingClassWithDuplicatedFeature(t *testing.T) {
	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	issuer := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
		},
	})

	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer:      issuer.String(),
		Symbol:      "symbol",
		Name:        "name",
		Description: "description",
		URI:         "https://my-class-meta.invalid/1",
		URIHash:     "content-hash",
		RoyaltyRate: sdk.ZeroDec(),
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_burning,
			assetnfttypes.ClassFeature_freezing,
			assetnfttypes.ClassFeature_whitelisting,
			assetnfttypes.ClassFeature_disable_sending,
			assetnfttypes.ClassFeature_burning,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.ErrorContains(err, "duplicated features in the class features list")
}

func (nt *nftFeaturesTest) createValidClass(t *testing.T) {
	requireT := require.New(t)

	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	issuer := chain.GenAccount()

	chain.FundAccountWithOptions(ctx, t, issuer, integrationtests.BalancesOptions{
		Messages: []sdk.Msg{
			&assetnfttypes.MsgIssueClass{},
		},
	})

	issueMsg := &assetnfttypes.MsgIssueClass{
		Issuer:      issuer.String(),
		Symbol:      "symbol",
		Name:        "name",
		Description: "description",
		URI:         "https://my-class-meta.invalid/1",
		URIHash:     "content-hash",
		RoyaltyRate: sdk.ZeroDec(),
		Features: []assetnfttypes.ClassFeature{
			assetnfttypes.ClassFeature_burning,
			assetnfttypes.ClassFeature_freezing,
			assetnfttypes.ClassFeature_whitelisting,
			assetnfttypes.ClassFeature_disable_sending,
		},
	}
	_, err := client.BroadcastTx(
		ctx,
		chain.ClientContext.WithFromAddress(issuer),
		chain.TxFactory().WithGas(chain.GasLimitByMsgs(issueMsg)),
		issueMsg,
	)
	requireT.NoError(err)
}

type nftWasmDataTest struct {
	data    []byte
	classID string
	nftID   string
}

func (n *nftWasmDataTest) Before(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	admin := chain.GenAccount()
	chain.Faucet.FundAccounts(ctx, t,
		integrationtests.NewFundedAccount(admin, chain.NewCoin(sdk.NewInt(5000000000))),
	)

	clientCtx := chain.ClientContext
	txf := chain.TxFactory().
		WithSimulateAndExecute(true)
	assetNFTClient := assetnfttypes.NewQueryClient(clientCtx)
	nftClient := nft.NewQueryClient(clientCtx)

	// ********** Issuance **********

	data := make([]byte, 256)
	for i := 0; i < 256; i++ {
		data[i] = uint8(i)
	}
	n.data = data

	encodedBytesString := base64.StdEncoding.EncodeToString(data)

	issueClassReq := moduleswasm.IssueNFTRequest{
		Name:        "name",
		Symbol:      "symbol",
		Description: "description",
		URI:         "https://my-nft-class-meta.invalid/1",
		URIHash:     "hash",
		Data:        encodedBytesString,
		RoyaltyRate: sdk.ZeroDec().String(),
	}
	issuerNFTInstantiatePayload, err := json.Marshal(issueClassReq)
	requireT.NoError(err)

	// instantiate new contract
	contractAddr, _, err := chain.Wasm.DeployAndInstantiateWASMContract(
		ctx,
		txf,
		admin,
		moduleswasm.NftWASM,
		integrationtests.InstantiateConfig{
			Payload: issuerNFTInstantiatePayload,
			Label:   "non_fungible_token",
		},
	)
	requireT.NoError(err)

	classID := assetnfttypes.BuildClassID(issueClassReq.Symbol, sdk.MustAccAddressFromBech32(contractAddr))
	classRes, err := assetNFTClient.Class(ctx, &assetnfttypes.QueryClassRequest{Id: classID})
	requireT.NoError(err)
	n.classID = classID

	dataBytes, err := codectypes.NewAnyWithValue(&assetnfttypes.DataBytes{Data: []byte(encodedBytesString)})
	dataToCompare := &codectypes.Any{
		TypeUrl: dataBytes.TypeUrl,
		Value:   dataBytes.Value,
	}
	requireT.NoError(err)

	expectedClass := assetnfttypes.Class{
		Id:          classID,
		Issuer:      contractAddr,
		Name:        issueClassReq.Name,
		Symbol:      issueClassReq.Symbol,
		Description: issueClassReq.Description,
		URI:         issueClassReq.URI,
		URIHash:     issueClassReq.URIHash,
		Data:        dataToCompare,
		Features:    issueClassReq.Features,
		RoyaltyRate: sdk.ZeroDec(),
	}
	requireT.Equal(
		expectedClass, classRes.Class,
	)

	// ********** Mint **********

	mintNFTReq := moduleswasm.NftMintRequest{
		ID:      "id-1",
		URI:     "https://my-nft-meta.invalid/1",
		URIHash: "hash",
		Data:    encodedBytesString,
	}
	mintPayload, err := json.Marshal(map[moduleswasm.NftMethod]moduleswasm.NftMintRequest{
		moduleswasm.NftMethodMint: mintNFTReq,
	})
	requireT.NoError(err)

	_, err = chain.Wasm.ExecuteWASMContract(ctx, txf, admin, contractAddr, mintPayload, sdk.Coin{})
	requireT.NoError(err)

	nftResp, err := nftClient.NFT(ctx, &nft.QueryNFTRequest{
		ClassId: classID,
		Id:      mintNFTReq.ID,
	})
	requireT.NoError(err)
	n.nftID = mintNFTReq.ID

	expectedNFT1 := &nft.NFT{
		ClassId: classID,
		Id:      mintNFTReq.ID,
		Uri:     mintNFTReq.URI,
		UriHash: mintNFTReq.URIHash,
		Data:    dataToCompare,
	}
	requireT.Equal(
		expectedNFT1, nftResp.Nft,
	)
}

func (n *nftWasmDataTest) After(t *testing.T) {
	ctx, chain := integrationtests.NewCoreumTestingContext(t)
	requireT := require.New(t)

	assetNFTClient := assetnfttypes.NewQueryClient(chain.ClientContext)
	nftQueryClient := nft.NewQueryClient(chain.ClientContext)

	dataBytes, err := codectypes.NewAnyWithValue(&assetnfttypes.DataBytes{Data: n.data})
	requireT.NoError(err)
	dataToCompare := &codectypes.Any{
		TypeUrl: dataBytes.TypeUrl,
		Value:   dataBytes.Value,
	}

	// query same nft class after the upgrade
	assetNftClassRes, err := assetNFTClient.Class(ctx, &assetnfttypes.QueryClassRequest{
		Id: n.classID,
	})
	requireT.NoError(err)
	requireT.Equal(dataToCompare, assetNftClassRes.Class.Data)

	//  query the same nft after the upgrade
	nftRes, err := nftQueryClient.NFT(ctx, &nft.QueryNFTRequest{
		ClassId: n.classID,
		Id:      n.nftID,
	})
	requireT.NoError(err)
	requireT.Equal(dataToCompare, nftRes.Nft.Data)
}
