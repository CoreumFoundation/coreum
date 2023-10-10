package keeper

import (
	"context"

	cosmosnft "github.com/cosmos/cosmos-sdk/x/nft"

	"github.com/CoreumFoundation/coreum/v3/x/nft"
)

var _ nft.QueryServer = Keeper{}

// Balance return the number of NFTs of a given class owned by the owner, same as balanceOf in ERC721.
func (k Keeper) Balance(goCtx context.Context, r *nft.QueryBalanceRequest) (*nft.QueryBalanceResponse, error) {
	resp, err := k.wkeeper.Balance(goCtx, &cosmosnft.QueryBalanceRequest{
		ClassId: r.ClassId,
		Owner:   r.Owner,
	})
	if err != nil {
		return nil, err
	}

	return &nft.QueryBalanceResponse{
		Amount: resp.Amount,
	}, nil
}

// Owner return the owner of the NFT based on its class and id, same as ownerOf in ERC721.
func (k Keeper) Owner(goCtx context.Context, r *nft.QueryOwnerRequest) (*nft.QueryOwnerResponse, error) {
	resp, err := k.wkeeper.Owner(goCtx, &cosmosnft.QueryOwnerRequest{
		ClassId: r.ClassId,
		Id:      r.Id,
	})
	if err != nil {
		return nil, err
	}

	return &nft.QueryOwnerResponse{
		Owner: resp.Owner,
	}, nil
}

// Supply return the number of NFTs from the given class, same as totalSupply of ERC721.
func (k Keeper) Supply(goCtx context.Context, r *nft.QuerySupplyRequest) (*nft.QuerySupplyResponse, error) {
	resp, err := k.wkeeper.Supply(goCtx, &cosmosnft.QuerySupplyRequest{
		ClassId: r.ClassId,
	})
	if err != nil {
		return nil, err
	}
	return &nft.QuerySupplyResponse{
		Amount: resp.Amount,
	}, nil
}

// NFTs queries all NFTs of a given class or owner (at least one must be provided), similar to tokenByIndex in ERC721Enumerable.
func (k Keeper) NFTs(goCtx context.Context, r *nft.QueryNFTsRequest) (*nft.QueryNFTsResponse, error) {
	resp, err := k.wkeeper.NFTs(goCtx, &cosmosnft.QueryNFTsRequest{
		ClassId:    r.ClassId,
		Owner:      r.Owner,
		Pagination: r.Pagination,
	})
	if err != nil {
		return nil, err
	}
	return &nft.QueryNFTsResponse{
		Nfts:       ConvertFromCosmosNFTList(resp.Nfts),
		Pagination: resp.Pagination,
	}, nil
}

// NFT return an NFT based on its class and id.
func (k Keeper) NFT(goCtx context.Context, r *nft.QueryNFTRequest) (*nft.QueryNFTResponse, error) {
	resp, err := k.wkeeper.NFT(goCtx, &cosmosnft.QueryNFTRequest{
		ClassId: r.ClassId,
		Id:      r.Id,
	})
	if err != nil {
		return nil, err
	}
	return &nft.QueryNFTResponse{
		Nft: ConvertFromCosmosNFT(resp.Nft),
	}, nil
}

// Class return an NFT class based on its id.
func (k Keeper) Class(goCtx context.Context, r *nft.QueryClassRequest) (*nft.QueryClassResponse, error) {
	resp, err := k.wkeeper.Class(goCtx, &cosmosnft.QueryClassRequest{
		ClassId: r.ClassId,
	})
	if err != nil {
		return nil, err
	}
	return &nft.QueryClassResponse{
		Class: ConvertFromCosmosClass(resp.Class),
	}, nil
}

// Classes return all NFT classes.
func (k Keeper) Classes(goCtx context.Context, r *nft.QueryClassesRequest) (*nft.QueryClassesResponse, error) {
	resp, err := k.wkeeper.Classes(goCtx, &cosmosnft.QueryClassesRequest{
		Pagination: r.Pagination,
	})
	if err != nil {
		return nil, err
	}
	return &nft.QueryClassesResponse{
		Classes:    ConvertFromCosmosClassList(resp.Classes),
		Pagination: resp.Pagination,
	}, nil
}

// ConvertFromCosmosNFT converts cosmos nft type to cnft type.
func ConvertFromCosmosNFT(n *cosmosnft.NFT) *nft.NFT {
	return &nft.NFT{
		ClassId: n.ClassId,
		Id:      n.Id,
		Uri:     n.Uri,
		UriHash: n.UriHash,
		Data:    n.Data,
	}
}

// ConvertFromCosmosNFTList converts cosmos nft type to cnft type.
func ConvertFromCosmosNFTList(ns []*cosmosnft.NFT) []*nft.NFT {
	var nfts []*nft.NFT
	for _, n := range ns {
		n := n
		nfts = append(nfts, ConvertFromCosmosNFT(n))
	}
	return nfts
}

// ConvertFromCosmosClass converts cosmos nft type to cnft type.
func ConvertFromCosmosClass(c *cosmosnft.Class) *nft.Class {
	return &nft.Class{
		Id:          c.Id,
		Name:        c.Name,
		Symbol:      c.Symbol,
		Description: c.Description,
		Uri:         c.Uri,
		UriHash:     c.UriHash,
		Data:        c.Data,
	}
}

// ConvertFromCosmosClassList converts cosmos nft type to cnft type.
func ConvertFromCosmosClassList(cs []*cosmosnft.Class) []*nft.Class {
	var classes []*nft.Class
	for _, cosmosClass := range cs {
		cosmosClass := cosmosClass
		classes = append(classes, ConvertFromCosmosClass(cosmosClass))
	}
	return classes
}
