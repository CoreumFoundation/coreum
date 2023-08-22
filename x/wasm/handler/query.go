package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/pkg/errors"

	assetfttypes "github.com/CoreumFoundation/coreum/v2/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/v2/x/asset/nft/types"
	nfttypes "github.com/CoreumFoundation/coreum/v2/x/nft"
)

// assetFTQuery represents asset ft module queries integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type assetFTQuery struct {
	Params              *assetfttypes.QueryParamsRequest              `json:"Params"`
	Token               *assetfttypes.QueryTokenRequest               `json:"Token"`
	Tokens              *assetfttypes.QueryTokensRequest              `json:"Tokens"`
	Balance             *assetfttypes.QueryBalanceRequest             `json:"Balance"`
	FrozenBalance       *assetfttypes.QueryFrozenBalanceRequest       `json:"FrozenBalance"`
	FrozenBalances      *assetfttypes.QueryFrozenBalancesRequest      `json:"FrozenBalances"`
	WhitelistedBalance  *assetfttypes.QueryWhitelistedBalanceRequest  `json:"WhitelistedBalance"`
	WhitelistedBalances *assetfttypes.QueryWhitelistedBalancesRequest `json:"WhitelistedBalances"`
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

// PageResponse is the structure used for pagination.
//
//nolint:tagliatelle // we keep the name same as consume
type pageResponse struct {
	NextKey []byte `json:"next_key"`
	Total   uint64 `json:"total"`
}

// assetNFTClassesResponse is the asset nft Classes response with string data.
type assetNFTClassesResponse struct {
	Pagination pageResponse    `json:"pagination"`
	Classes    []assetNFTClass `json:"classes"`
}

// assetNFTQuery represents asset nft module queries integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type assetNFTQuery struct {
	Params                    *assetnfttypes.QueryParamsRequest                    `json:"Params"`
	Class                     *assetnfttypes.QueryClassRequest                     `json:"Class"`
	Classes                   *assetnfttypes.QueryClassesRequest                   `json:"Classes"`
	Frozen                    *assetnfttypes.QueryFrozenRequest                    `json:"Frozen"`
	Whitelisted               *assetnfttypes.QueryWhitelistedRequest               `json:"Whitelisted"`
	WhitelistedAccountsforNFT *assetnfttypes.QueryWhitelistedAccountsForNFTRequest `json:"WhitelistedAccountsforNft"`
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

// NFTsResponse is the nfts response with string data.
type NFTsResponse struct {
	NFTs       []nft        `json:"nfts"`
	Pagination pageResponse `json:"pagination"`
}

// NFTClass is the NFTClass with string data.
//
//nolint:tagliatelle // we keep the name same as consume
type NFTClass struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	URI         string `json:"uri"`
	URIHash     string `json:"uri_hash"`
	Data        string `json:"data"`
}

// NFTClassResponse is the NFTClassResponse response with string data.
type NFTClassResponse struct {
	Class NFTClass `json:"class"`
}

// NFTClassesResponse is the NFTClassesResponse with string data.
type NFTClassesResponse struct {
	Classes    []NFTClass   `json:"classes"`
	Pagination pageResponse `json:"pagination"`
}

// nftQuery represents nft module queries integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type nftQuery struct {
	Balance *nfttypes.QueryBalanceRequest `json:"Balance"`
	Owner   *nfttypes.QueryOwnerRequest   `json:"Owner"`
	Supply  *nfttypes.QuerySupplyRequest  `json:"Supply"`
	NFT     *nfttypes.QueryNFTRequest     `json:"nft"`
	NFTs    *nfttypes.QueryNFTsRequest    `json:"nfts"`
	Class   *nfttypes.QueryClassRequest   `json:"Class"`
	Classes *nfttypes.QueryClassesRequest `json:"Classes"`
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
	databytes, err := base64.StdEncoding.DecodeString(dataString)
	if err != nil {
		return nil, err
	}
	dataValue, err := codectypes.NewAnyWithValue(&assetnfttypes.DataBytes{Data: databytes})
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
	if assetFTQuery.Params != nil {
		return executeQuery(ctx, assetFTQuery.Params, func(ctx context.Context, req *assetfttypes.QueryParamsRequest) (*assetfttypes.QueryParamsResponse, error) {
			return assetFTQueryServer.Params(ctx, req)
		})
	}
	if assetFTQuery.Token != nil {
		return executeQuery(ctx, assetFTQuery.Token, func(ctx context.Context, req *assetfttypes.QueryTokenRequest) (*assetfttypes.QueryTokenResponse, error) {
			return assetFTQueryServer.Token(ctx, req)
		})
	}
	if assetFTQuery.Tokens != nil {
		return executeQuery(ctx, assetFTQuery.Tokens, func(ctx context.Context, req *assetfttypes.QueryTokensRequest) (*assetfttypes.QueryTokensResponse, error) {
			return assetFTQueryServer.Tokens(ctx, req)
		})
	}
	if assetFTQuery.Balance != nil {
		return executeQuery(ctx, assetFTQuery.Balance, func(ctx context.Context, req *assetfttypes.QueryBalanceRequest) (*assetfttypes.QueryBalanceResponse, error) {
			return assetFTQueryServer.Balance(ctx, req)
		})
	}
	if assetFTQuery.FrozenBalance != nil {
		return executeQuery(ctx, assetFTQuery.FrozenBalance, func(ctx context.Context, req *assetfttypes.QueryFrozenBalanceRequest) (*assetfttypes.QueryFrozenBalanceResponse, error) {
			return assetFTQueryServer.FrozenBalance(ctx, req)
		})
	}
	if assetFTQuery.FrozenBalances != nil {
		return executeQuery(ctx, assetFTQuery.FrozenBalances, func(ctx context.Context, req *assetfttypes.QueryFrozenBalancesRequest) (*assetfttypes.QueryFrozenBalancesResponse, error) {
			return assetFTQueryServer.FrozenBalances(ctx, req)
		})
	}
	if assetFTQuery.WhitelistedBalance != nil {
		return executeQuery(ctx, assetFTQuery.WhitelistedBalance, func(ctx context.Context, req *assetfttypes.QueryWhitelistedBalanceRequest) (*assetfttypes.QueryWhitelistedBalanceResponse, error) {
			return assetFTQueryServer.WhitelistedBalance(ctx, req)
		})
	}
	if assetFTQuery.WhitelistedBalances != nil {
		return executeQuery(ctx, assetFTQuery.WhitelistedBalances, func(ctx context.Context, req *assetfttypes.QueryWhitelistedBalancesRequest) (*assetfttypes.QueryWhitelistedBalancesResponse, error) {
			return assetFTQueryServer.WhitelistedBalances(ctx, req)
		})
	}

	return nil, nil
}

func processAssetNFTQuery(ctx sdk.Context, assetNFTQuery *assetNFTQuery, assetNFTQueryServer assetnfttypes.QueryServer) ([]byte, error) {
	if assetNFTQuery.Params != nil {
		return executeQuery(ctx, assetNFTQuery.Params, func(ctx context.Context, req *assetnfttypes.QueryParamsRequest) (*assetnfttypes.QueryParamsResponse, error) {
			return assetNFTQueryServer.Params(ctx, req)
		})
	}
	if assetNFTQuery.Class != nil {
		return executeQuery(ctx, assetNFTQuery.Class, func(ctx context.Context, req *assetnfttypes.QueryClassRequest) (*assetNFTClassResponse, error) {
			classRes, err := assetNFTQueryServer.Class(ctx, req)
			if err != nil {
				return nil, err
			}

			var dataString string
			if classRes.Class.Data != nil {
				dataString, err = unmarshalDataBytes(classRes.Class.Data)
				if err != nil {
					return nil, err
				}
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
	if assetNFTQuery.Classes != nil {
		return executeQuery(ctx, assetNFTQuery.Classes, func(ctx context.Context, req *assetnfttypes.QueryClassesRequest) (*assetNFTClassesResponse, error) {
			classesRes, err := assetNFTQueryServer.Classes(ctx, req)
			if err != nil {
				return nil, err
			}

			var classesResponse assetNFTClassesResponse

			classesResponse.Pagination.NextKey = classesRes.Pagination.NextKey
			classesResponse.Pagination.Total = classesRes.Pagination.Total
			for i := 0; i < len(classesRes.Classes); i++ {
				var dataString string
				if classesRes.Classes[i].Data != nil {
					dataString, err = unmarshalDataBytes(classesRes.Classes[i].Data)
					if err != nil {
						return nil, err
					}
				}
				classesResponse.Classes = append(classesResponse.Classes, assetNFTClass{
					ID:          classesRes.Classes[i].Id,
					Issuer:      classesRes.Classes[i].Issuer,
					Name:        classesRes.Classes[i].Name,
					Symbol:      classesRes.Classes[i].Symbol,
					Description: classesRes.Classes[i].Description,
					URI:         classesRes.Classes[i].URI,
					URIHash:     classesRes.Classes[i].URIHash,
					Data:        dataString,
					Features:    classesRes.Classes[i].Features,
					RoyaltyRate: classesRes.Classes[i].RoyaltyRate,
				})
			}
			return &classesResponse, nil
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
	if assetNFTQuery.WhitelistedAccountsforNFT != nil {
		return executeQuery(ctx, assetNFTQuery.WhitelistedAccountsforNFT, func(ctx context.Context, req *assetnfttypes.QueryWhitelistedAccountsForNFTRequest) (*assetnfttypes.QueryWhitelistedAccountsForNFTResponse, error) {
			return assetNFTQueryServer.WhitelistedAccountsForNFT(ctx, req)
		})
	}

	return nil, nil
}

//nolint:funlen
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
	if nftQuery.NFT != nil { //nolint:nestif // the ifs are for the error checks mostly
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
				dataString, err = unmarshalDataBytes(nftRes.Nft.Data)
				if err != nil {
					return nil, err
				}
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
	if nftQuery.NFTs != nil { //nolint:nestif // the ifs are for the error checks mostly
		return executeQuery(ctx, nftQuery.NFTs, func(ctx context.Context, req *nfttypes.QueryNFTsRequest) (*NFTsResponse, error) {
			nftsRes, err := nftQueryServer.NFTs(ctx, req)
			if err != nil {
				return nil, err
			}

			if nftsRes.Nfts == nil {
				return &NFTsResponse{}, nil
			}

			var nftsResponse NFTsResponse

			nftsResponse.Pagination.NextKey = nftsRes.Pagination.NextKey
			nftsResponse.Pagination.Total = nftsRes.Pagination.Total
			for i := 0; i < len(nftsRes.Nfts); i++ {
				var dataString string
				if nftsRes.Nfts[i].Data != nil {
					dataString, err = unmarshalDataBytes(nftsRes.Nfts[i].Data)
					if err != nil {
						return nil, err
					}
				}
				nftsResponse.NFTs = append(nftsResponse.NFTs, nft{
					ClassID: nftsRes.Nfts[i].ClassId,
					ID:      nftsRes.Nfts[i].Id,
					URI:     nftsRes.Nfts[i].Uri,
					URIHash: nftsRes.Nfts[i].UriHash,
					Data:    dataString,
				})
			}
			return &nftsResponse, nil
		})
	}
	if nftQuery.Class != nil { //nolint:nestif // the ifs are for the error checks mostly
		return executeQuery(ctx, nftQuery.Class, func(ctx context.Context, req *nfttypes.QueryClassRequest) (*NFTClassResponse, error) {
			nftClassRes, err := nftQueryServer.Class(ctx, req)
			if err != nil {
				return nil, err
			}

			if nftClassRes.Class == nil {
				return &NFTClassResponse{}, nil
			}

			var dataString string
			if nftClassRes.Class.Data != nil {
				dataString, err = unmarshalDataBytes(nftClassRes.Class.Data)
				if err != nil {
					return nil, err
				}
			}
			return &NFTClassResponse{
				Class: NFTClass{
					ID:          nftClassRes.Class.Id,
					Name:        nftClassRes.Class.Name,
					Symbol:      nftClassRes.Class.Symbol,
					Description: nftClassRes.Class.Description,
					URI:         nftClassRes.Class.Uri,
					URIHash:     nftClassRes.Class.UriHash,
					Data:        dataString,
				},
			}, nil
		})
	}
	if nftQuery.Classes != nil { //nolint:nestif // the ifs are for the error checks mostly
		return executeQuery(ctx, nftQuery.Classes, func(ctx context.Context, req *nfttypes.QueryClassesRequest) (*NFTClassesResponse, error) {
			nftClassesRes, err := nftQueryServer.Classes(ctx, req)
			if err != nil {
				return nil, err
			}

			if nftClassesRes.Classes == nil {
				return &NFTClassesResponse{}, nil
			}

			var nftClassesResponse NFTClassesResponse

			nftClassesResponse.Pagination.NextKey = nftClassesRes.Pagination.NextKey
			nftClassesResponse.Pagination.Total = nftClassesRes.Pagination.Total

			for i := 0; i < len(nftClassesRes.Classes); i++ {
				var dataString string
				if nftClassesRes.Classes[i].Data != nil {
					dataString, err = unmarshalDataBytes(nftClassesRes.Classes[i].Data)
					if err != nil {
						return nil, err
					}
				}

				nftClassesResponse.Classes = append(nftClassesResponse.Classes, NFTClass{
					ID:          nftClassesRes.Classes[i].Id,
					Name:        nftClassesRes.Classes[i].Name,
					Symbol:      nftClassesRes.Classes[i].Symbol,
					Description: nftClassesRes.Classes[i].Description,
					URI:         nftClassesRes.Classes[i].Uri,
					URIHash:     nftClassesRes.Classes[i].UriHash,
					Data:        dataString,
				})
			}
			return &nftClassesResponse, nil
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

func unmarshalDataBytes(data *codectypes.Any) (string, error) {
	var dataBytes assetnfttypes.DataBytes
	err := proto.Unmarshal(data.Value, &dataBytes)
	if err != nil {
		return "", errors.WithStack(err)
	}

	return base64.StdEncoding.EncodeToString(dataBytes.Data), nil
}
