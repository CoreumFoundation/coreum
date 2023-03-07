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

// AssetFTMsg represents asset ft module messages integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type AssetFTMsg struct {
	Issue               *assetfttypes.MsgIssue               `json:"Issue"`
	Mint                *assetfttypes.MsgMint                `json:"Mint"`
	Burn                *assetfttypes.MsgBurn                `json:"Burn"`
	Freeze              *assetfttypes.MsgFreeze              `json:"Freeze"`
	Unfreeze            *assetfttypes.MsgUnfreeze            `json:"Unfreeze"`
	GloballyFreeze      *assetfttypes.MsgGloballyFreeze      `json:"GloballyFreeze"`
	GloballyUnfreeze    *assetfttypes.MsgGloballyUnfreeze    `json:"GloballyUnfreeze"`
	SetWhitelistedLimit *assetfttypes.MsgSetWhitelistedLimit `json:"SetWhitelistedLimit"`
}

// AssetNFTMsgIssueClass defines message for the IssueClass method with string represented data field.
//
//nolint:tagliatelle // we keep the name same as consume
type AssetNFTMsgIssueClass struct {
	Symbol      string                       `json:"symbol"`
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	URI         string                       `json:"uri"`
	URIHash     string                       `json:"uri_hash"`
	Data        string                       `json:"data"`
	Features    []assetnfttypes.ClassFeature `json:"features"`
	RoyaltyRate sdk.Dec                      `json:"royalty_rate"`
}

// AssetNFTMsgMint defines message for the Mint method with string represented data field.
//
//nolint:tagliatelle // we keep the name same as consume
type AssetNFTMsgMint struct {
	ClassID string `json:"class_id"`
	ID      string `json:"id"`
	URI     string `json:"uri"`
	URIHash string `json:"uri_hash"`
	Data    string `json:"data"`
}

// AssetNFTMsg represents asset nft module messages integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type AssetNFTMsg struct {
	IssueClass          *AssetNFTMsgIssueClass                `json:"IssueClass"`
	Mint                *AssetNFTMsgMint                      `json:"Mint"`
	Burn                *assetnfttypes.MsgBurn                `json:"Burn"`
	Freeze              *assetnfttypes.MsgFreeze              `json:"Freeze"`
	Unfreeze            *assetnfttypes.MsgUnfreeze            `json:"Unfreeze"`
	AddToWhitelist      *assetnfttypes.MsgAddToWhitelist      `json:"AddToWhitelist"`
	RemoveFromWhitelist *assetnfttypes.MsgRemoveFromWhitelist `json:"RemoveFromWhitelist"`
}

// NFTMsg represents nft module messages integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type NFTMsg struct {
	Send *nfttypes.MsgSend `json:"Send"`
}

// CoreumMsg represents all supported custom messages integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type CoreumMsg struct {
	AssetFT  *AssetFTMsg  `json:"AssetFT"`
	AssetNFT *AssetNFTMsg `json:"AssetNFT"`
	NFT      *NFTMsg      `json:"NFT"`
}

// AssetFTQuery represents asset ft module queries integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type AssetFTQuery struct {
	Token              *assetfttypes.QueryTokenRequest              `json:"Token"`
	FrozenBalance      *assetfttypes.QueryFrozenBalanceRequest      `json:"FrozenBalance"`
	WhitelistedBalance *assetfttypes.QueryWhitelistedBalanceRequest `json:"WhitelistedBalance"`
}

// AssetNFTClass is the asset NFT Class with string data.
//
//nolint:tagliatelle // we keep the name same as consume
type AssetNFTClass struct {
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

// AssetNFTClassResponse is the asset NFT Class response with string data.
type AssetNFTClassResponse struct {
	Class AssetNFTClass `json:"class"`
}

// AssetNFTQuery represents asset nft module queries integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type AssetNFTQuery struct {
	Class       *assetnfttypes.QueryClassRequest       `json:"Class"`
	Frozen      *assetnfttypes.QueryFrozenRequest      `json:"Frozen"`
	Whitelisted *assetnfttypes.QueryWhitelistedRequest `json:"Whitelisted"`
}

// NFT is the NFT with string data.
//
//nolint:tagliatelle // we keep the name same as consume
type NFT struct {
	ClassID string `json:"class_id"`
	ID      string `json:"id"`
	URI     string `json:"uri"`
	URIHash string `json:"uri_hash"`
	Data    string `json:"data"`
}

// NFTResponse is the NFT response with string data.
type NFTResponse struct {
	NFT NFT `json:"nft"`
}

// NFTQuery represents nft module queries integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type NFTQuery struct {
	Balance *nfttypes.QueryBalanceRequest `json:"Balance"`
	Owner   *nfttypes.QueryOwnerRequest   `json:"Owner"`
	Supply  *nfttypes.QuerySupplyRequest  `json:"Supply"`
	NFT     *nfttypes.QueryNFTRequest     `json:"NFT"`
}

// CoreumQuery represents all coreum module queries integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type CoreumQuery struct {
	AssetFT  *AssetFTQuery  `json:"AssetFT"`
	AssetNFT *AssetNFTQuery `json:"AssetNFT"`
	NFT      *NFTQuery      `json:"NFT"`
}

// NewCoreumMsgHandler returns coreum handler that handles messages received from smart contracts.
// The in the input sender is the address of smart contract.
func NewCoreumMsgHandler() *wasmkeeper.MessageEncoders {
	return &wasmkeeper.MessageEncoders{
		Custom: func(sender sdk.AccAddress, msg json.RawMessage) ([]sdk.Msg, error) {
			var coreumMsg CoreumMsg
			if err := json.Unmarshal(msg, &coreumMsg); err != nil {
				return nil, errors.WithStack(err)
			}

			decodedMsg, err := decodeCoreumMessage(coreumMsg, sender)
			if err != nil {
				return nil, err
			}
			if decodedMsg == nil {
				return nil, nil
			}

			if err := decodedMsg.ValidateBasic(); err != nil {
				return nil, errors.WithStack(err)
			}

			return []sdk.Msg{decodedMsg}, nil
		},
	}
}

// NewCoreumQueryHandler returns the coreum handler which handles queries from smart contracts.
func NewCoreumQueryHandler(
	assetFTQueryServer assetfttypes.QueryServer,
	assetNFTQueryServer assetnfttypes.QueryServer,
	nftQueryServer nfttypes.QueryServer,
) *wasmkeeper.QueryPlugins {
	return &wasmkeeper.QueryPlugins{
		Custom: func(ctx sdk.Context, query json.RawMessage) ([]byte, error) {
			var coreumQuery CoreumQuery
			if err := json.Unmarshal(query, &coreumQuery); err != nil {
				return nil, errors.WithStack(err)
			}

			return processCoreumQuery(ctx, coreumQuery, assetFTQueryServer, assetNFTQueryServer, nftQueryServer)
		},
	}
}

func decodeCoreumMessage(coreumMessages CoreumMsg, sender sdk.AccAddress) (sdk.Msg, error) {
	if coreumMessages.AssetFT != nil {
		return decodeAssetFTMessage(coreumMessages.AssetFT, sender.String())
	}
	if coreumMessages.AssetNFT != nil {
		return decodeAssetNFTMessage(coreumMessages.AssetNFT, sender.String())
	}
	if coreumMessages.NFT != nil {
		return decodeNFTMessage(coreumMessages.NFT, sender.String())
	}

	return nil, nil
}

func decodeAssetFTMessage(assetFTMsg *AssetFTMsg, sender string) (sdk.Msg, error) {
	if assetFTMsg.Issue != nil {
		assetFTMsg.Issue.Issuer = sender
		return assetFTMsg.Issue, nil
	}
	if assetFTMsg.Mint != nil {
		assetFTMsg.Mint.Sender = sender
		return assetFTMsg.Mint, nil
	}
	if assetFTMsg.Burn != nil {
		assetFTMsg.Burn.Sender = sender
		return assetFTMsg.Burn, nil
	}
	if assetFTMsg.Freeze != nil {
		assetFTMsg.Freeze.Sender = sender
		return assetFTMsg.Freeze, nil
	}
	if assetFTMsg.Unfreeze != nil {
		assetFTMsg.Unfreeze.Sender = sender
		return assetFTMsg.Unfreeze, nil
	}
	if assetFTMsg.GloballyFreeze != nil {
		assetFTMsg.GloballyFreeze.Sender = sender
		return assetFTMsg.GloballyFreeze, nil
	}
	if assetFTMsg.GloballyUnfreeze != nil {
		assetFTMsg.GloballyUnfreeze.Sender = sender
		return assetFTMsg.GloballyUnfreeze, nil
	}
	if assetFTMsg.SetWhitelistedLimit != nil {
		assetFTMsg.SetWhitelistedLimit.Sender = sender
		return assetFTMsg.SetWhitelistedLimit, nil
	}

	return nil, nil
}

func decodeAssetNFTMessage(assetNFTMsg *AssetNFTMsg, sender string) (sdk.Msg, error) {
	if assetNFTMsg.IssueClass != nil {
		var (
			data *codectypes.Any
			err  error
		)
		if assetNFTMsg.IssueClass.Data != "" {
			data, err = convertStringToDataBytes(assetNFTMsg.IssueClass.Data)
			if err != nil {
				return nil, err
			}
		}
		return &assetnfttypes.MsgIssueClass{
			Issuer:      sender,
			Symbol:      assetNFTMsg.IssueClass.Symbol,
			Name:        assetNFTMsg.IssueClass.Name,
			Description: assetNFTMsg.IssueClass.Description,
			URI:         assetNFTMsg.IssueClass.URI,
			URIHash:     assetNFTMsg.IssueClass.URIHash,
			Data:        data,
			Features:    assetNFTMsg.IssueClass.Features,
			RoyaltyRate: assetNFTMsg.IssueClass.RoyaltyRate,
		}, nil
	}
	if assetNFTMsg.Mint != nil {
		var (
			data *codectypes.Any
			err  error
		)
		if assetNFTMsg.Mint.Data != "" {
			data, err = convertStringToDataBytes(assetNFTMsg.Mint.Data)
			if err != nil {
				return nil, err
			}
		}
		return &assetnfttypes.MsgMint{
			Sender:  sender,
			ClassID: assetNFTMsg.Mint.ClassID,
			ID:      assetNFTMsg.Mint.ID,
			URI:     assetNFTMsg.Mint.URI,
			URIHash: assetNFTMsg.Mint.URIHash,
			Data:    data,
		}, nil
	}
	if assetNFTMsg.Burn != nil {
		assetNFTMsg.Burn.Sender = sender
		return assetNFTMsg.Burn, nil
	}
	if assetNFTMsg.Freeze != nil {
		assetNFTMsg.Freeze.Sender = sender
		return assetNFTMsg.Freeze, nil
	}
	if assetNFTMsg.Unfreeze != nil {
		assetNFTMsg.Unfreeze.Sender = sender
		return assetNFTMsg.Unfreeze, nil
	}
	if assetNFTMsg.AddToWhitelist != nil {
		assetNFTMsg.AddToWhitelist.Sender = sender
		return assetNFTMsg.AddToWhitelist, nil
	}
	if assetNFTMsg.RemoveFromWhitelist != nil {
		assetNFTMsg.RemoveFromWhitelist.Sender = sender
		return assetNFTMsg.RemoveFromWhitelist, nil
	}

	return nil, nil
}

func decodeNFTMessage(nftMsg *NFTMsg, sender string) (sdk.Msg, error) {
	if nftMsg.Send != nil {
		nftMsg.Send.Sender = sender
		return nftMsg.Send, nil
	}

	return nil, nil
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
	queries CoreumQuery,
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

func processAssetFTQuery(ctx sdk.Context, assetFTQuery *AssetFTQuery, assetFTQueryServer assetfttypes.QueryServer) ([]byte, error) {
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

func processAssetNFTQuery(ctx sdk.Context, assetNFTQuery *AssetNFTQuery, assetNFTQueryServer assetnfttypes.QueryServer) ([]byte, error) {
	if assetNFTQuery.Class != nil {
		return executeQuery(ctx, assetNFTQuery.Class, func(ctx context.Context, req *assetnfttypes.QueryClassRequest) (*AssetNFTClassResponse, error) {
			classRes, err := assetNFTQueryServer.Class(ctx, req)
			if err != nil {
				return nil, err
			}

			var dataString string
			if classRes.Class.Data != nil {
				dataString = string(classRes.Class.Data.Value)
			}
			return &AssetNFTClassResponse{
				Class: AssetNFTClass{
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
			}, err
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

func processNFTQuery(ctx sdk.Context, nftQuery *NFTQuery, nftQueryServer nfttypes.QueryServer) ([]byte, error) {
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
				NFT: NFT{
					ClassID: nftRes.Nft.ClassId,
					ID:      nftRes.Nft.Id,
					URI:     nftRes.Nft.Uri,
					URIHash: nftRes.Nft.UriHash,
					Data:    dataString,
				},
			}, err
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
