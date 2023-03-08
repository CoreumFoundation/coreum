package handler

import (
	"context"
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/x/asset/nft/types"
	nfttypes "github.com/CoreumFoundation/coreum/x/nft"
)

// assetFTQuery represents asset ft module queries integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type assetFTQuery struct {
	Token              *assetfttypes.QueryTokenRequest              `json:"Token"`
	FrozenBalance      *assetfttypes.QueryFrozenBalanceRequest      `json:"FrozenBalance"`
	WhitelistedBalance *assetfttypes.QueryWhitelistedBalanceRequest `json:"WhitelistedBalance"`
}

// assetNFTClass is the asset nft Class with string data.
//
//nolint:tagliatelle // we keep the name same as consume
type assetNFTClass struct {
	ID          string                       `json:"id"`
	Issuer      string                       `json:"issuer"`
	Name        string                       `json:"name"`
	Symbol      string                       `json:"symbol"`
	Description string                       `json:"description"`
	URI         string                       `json:"uri"`
	URIHash     string                       `json:"uri_hash"`
	Data        string                       `json:"data"`
	Features    []assetnfttypes.ClassFeature `json:"features"`
	RoyaltyRate sdk.Dec                      `json:"royalty_rate"`
}

// assetNFTClassResponse is the asset nft Class response with string data.
type assetNFTClassResponse struct {
	Class assetNFTClass `json:"class"`
}

// assetNFTQuery represents asset nft module queries integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type assetNFTQuery struct {
	Class       *assetnfttypes.QueryClassRequest       `json:"Class"`
	Frozen      *assetnfttypes.QueryFrozenRequest      `json:"Frozen"`
	Whitelisted *assetnfttypes.QueryWhitelistedRequest `json:"Whitelisted"`
}

// nft is the nft with string data.
//
//nolint:tagliatelle // we keep the name same as consume
type nft struct {
	ClassID string `json:"class_id"`
	ID      string `json:"id"`
	URI     string `json:"uri"`
	URIHash string `json:"uri_hash"`
	Data    string `json:"data"`
}

// NFTResponse is the nft response with string data.
type NFTResponse struct {
	NFT nft `json:"nft"`
}

// nftQuery represents nft module queries integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type nftQuery struct {
	Balance *nfttypes.QueryBalanceRequest `json:"Balance"`
	Owner   *nfttypes.QueryOwnerRequest   `json:"Owner"`
	Supply  *nfttypes.QuerySupplyRequest  `json:"Supply"`
	NFT     *nfttypes.QueryNFTRequest     `json:"nft"`
}

// coreumQuery represents all coreum module queries integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type coreumQuery struct {
	AssetFT  *assetFTQuery  `json:"AssetFT"`
	AssetNFT *assetNFTQuery `json:"AssetNFT"`
	NFT      *nftQuery      `json:"nft"`
}

// NewCoreumQueryHandler returns the coreum handler which handles queries from smart contracts.
func NewCoreumQueryHandler(
	assetFTQueryServer assetfttypes.QueryServer,
	assetNFTQueryServer assetnfttypes.QueryServer,
	nftQueryServer nfttypes.QueryServer,
) *wasmkeeper.QueryPlugins {
	return &wasmkeeper.QueryPlugins{
		Custom: func(ctx sdk.Context, query json.RawMessage) ([]byte, error) {
			var coreumQuery coreumQuery
			if err := json.Unmarshal(query, &coreumQuery); err != nil {
				return nil, errors.WithStack(err)
			}

			return processCoreumQuery(ctx, coreumQuery, assetFTQueryServer, assetNFTQueryServer, nftQueryServer)
		},
	}
}

func convertStringToDataBytes(dataString string) (*codectypes.Any, error) {
	dataValue, err := codectypes.NewAnyWithValue(&assetnfttypes.DataBytes{Data: []byte(dataString)})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return dataValue, nil
}

func processCoreumQuery(
	ctx sdk.Context,
	queries coreumQuery,
	assetFTQueryServer assetfttypes.QueryServer,
	assetNFTQueryServer assetnfttypes.QueryServer,
	nftQueryServer nfttypes.QueryServer,
) ([]byte, error) {
	if queries.AssetFT != nil {
		return processAssetFTQuery(ctx, queries.AssetFT, assetFTQueryServer)
	}
	if queries.AssetNFT != nil {
		return processAssetNFTQuery(ctx, queries.AssetNFT, assetNFTQueryServer)
	}
	if queries.NFT != nil {
		return processNFTQuery(ctx, queries.NFT, nftQueryServer)
	}

	return nil, nil
}

func processAssetFTQuery(ctx sdk.Context, assetFTQuery *assetFTQuery, assetFTQueryServer assetfttypes.QueryServer) ([]byte, error) {
	if assetFTQuery.Token != nil {
		return executeQuery(ctx, assetFTQuery.Token, func(ctx context.Context, req *assetfttypes.QueryTokenRequest) (*assetfttypes.QueryTokenResponse, error) {
			return assetFTQueryServer.Token(ctx, req)
		})
	}
	if assetFTQuery.FrozenBalance != nil {
		return executeQuery(ctx, assetFTQuery.FrozenBalance, func(ctx context.Context, req *assetfttypes.QueryFrozenBalanceRequest) (*assetfttypes.QueryFrozenBalanceResponse, error) {
			return assetFTQueryServer.FrozenBalance(ctx, req)
		})
	}
	if assetFTQuery.WhitelistedBalance != nil {
		return executeQuery(ctx, assetFTQuery.WhitelistedBalance, func(ctx context.Context, req *assetfttypes.QueryWhitelistedBalanceRequest) (*assetfttypes.QueryWhitelistedBalanceResponse, error) {
			return assetFTQueryServer.WhitelistedBalance(ctx, req)
		})
	}

	return nil, nil
}

func processAssetNFTQuery(ctx sdk.Context, assetNFTQuery *assetNFTQuery, assetNFTQueryServer assetnfttypes.QueryServer) ([]byte, error) {
	if assetNFTQuery.Class != nil {
		return executeQuery(ctx, assetNFTQuery.Class, func(ctx context.Context, req *assetnfttypes.QueryClassRequest) (*assetNFTClassResponse, error) {
			classRes, err := assetNFTQueryServer.Class(ctx, req)
			if err != nil {
				return nil, err
			}

			var dataString string
			if classRes.Class.Data != nil {
				dataString = string(classRes.Class.Data.Value)
			}
			return &assetNFTClassResponse{
				Class: assetNFTClass{
					ID:          classRes.Class.Id,
					Issuer:      classRes.Class.Issuer,
					Name:        classRes.Class.Name,
					Symbol:      classRes.Class.Symbol,
					Description: classRes.Class.Description,
					URI:         classRes.Class.URI,
					URIHash:     classRes.Class.URIHash,
					Data:        dataString,
					Features:    classRes.Class.Features,
					RoyaltyRate: classRes.Class.RoyaltyRate,
				},
			}, nil
		})
	}
	if assetNFTQuery.Frozen != nil {
		return executeQuery(ctx, assetNFTQuery.Frozen, func(ctx context.Context, req *assetnfttypes.QueryFrozenRequest) (*assetnfttypes.QueryFrozenResponse, error) {
			return assetNFTQueryServer.Frozen(ctx, req)
		})
	}
	if assetNFTQuery.Whitelisted != nil {
		return executeQuery(ctx, assetNFTQuery.Whitelisted, func(ctx context.Context, req *assetnfttypes.QueryWhitelistedRequest) (*assetnfttypes.QueryWhitelistedResponse, error) {
			return assetNFTQueryServer.Whitelisted(ctx, req)
		})
	}

	return nil, nil
}

func processNFTQuery(ctx sdk.Context, nftQuery *nftQuery, nftQueryServer nfttypes.QueryServer) ([]byte, error) {
	if nftQuery.Balance != nil {
		return executeQuery(ctx, nftQuery.Balance, func(ctx context.Context, req *nfttypes.QueryBalanceRequest) (*nfttypes.QueryBalanceResponse, error) {
			return nftQueryServer.Balance(ctx, req)
		})
	}
	if nftQuery.Owner != nil {
		return executeQuery(ctx, nftQuery.Owner, func(ctx context.Context, req *nfttypes.QueryOwnerRequest) (*nfttypes.QueryOwnerResponse, error) {
			return nftQueryServer.Owner(ctx, req)
		})
	}
	if nftQuery.Supply != nil {
		return executeQuery(ctx, nftQuery.Supply, func(ctx context.Context, req *nfttypes.QuerySupplyRequest) (*nfttypes.QuerySupplyResponse, error) {
			return nftQueryServer.Supply(ctx, req)
		})
	}
	if nftQuery.NFT != nil {
		return executeQuery(ctx, nftQuery.NFT, func(ctx context.Context, req *nfttypes.QueryNFTRequest) (*NFTResponse, error) {
			nftRes, err := nftQueryServer.NFT(ctx, req)
			if err != nil {
				return nil, err
			}

			if nftRes.Nft == nil {
				return &NFTResponse{}, nil
			}

			var dataString string
			if nftRes.Nft.Data != nil {
				dataString = string(nftRes.Nft.Data.Value)
			}
			return &NFTResponse{
				NFT: nft{
					ClassID: nftRes.Nft.ClassId,
					ID:      nftRes.Nft.Id,
					URI:     nftRes.Nft.Uri,
					URIHash: nftRes.Nft.UriHash,
					Data:    dataString,
				},
			}, nil
		})
	}

	return nil, nil
}

func executeQuery[T, K any](
	ctx sdk.Context,
	reqStruct T,
	reqExecutor func(ctx context.Context, req T) (K, error),
) (json.RawMessage, error) {
	res, err := reqExecutor(sdk.WrapSDKContext(ctx), reqStruct)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	raw, err := json.Marshal(res)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return raw, nil
}
