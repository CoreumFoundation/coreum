package modules

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	assetnfttypes "github.com/CoreumFoundation/coreum/v3/x/asset/nft/types"
)

// IssueNFTRequest is used to issue NFTs.
//
//nolint:tagliatelle
type IssueNFTRequest struct {
	Name        string                       `json:"name"`
	Symbol      string                       `json:"symbol"`
	Description string                       `json:"description"`
	URI         string                       `json:"uri"`
	URIHash     string                       `json:"uri_hash"`
	Data        string                       `json:"data"`
	Features    []assetnfttypes.ClassFeature `json:"features"`
	RoyaltyRate string                       `json:"royalty_rate"`
}

// NftMintRequest is used to mint NFTs.
//
//nolint:tagliatelle
type NftMintRequest struct {
	ID      string `json:"id"`
	URI     string `json:"uri"`
	URIHash string `json:"uri_hash"`
	Data    string `json:"data"`
}

// NftIDRequest is used to query NFT with ID.
type NftIDRequest struct {
	ID string `json:"id"`
}

// NftIssuerRequest is used to query NFT with issuer.
type NftIssuerRequest struct {
	Issuer string `json:"issuer"`
}

// NftIDWithAccountRequest is used to query NFT with id and account.
type NftIDWithAccountRequest struct {
	ID      string `json:"id"`
	Account string `json:"account"`
}

// NftIDWithReceiverRequest is used query NFT with id and receiver.
type NftIDWithReceiverRequest struct {
	ID       string `json:"id"`
	Receiver string `json:"receiver"`
}

// NftOwnerRequest is used to query the NFT with owner.
type NftOwnerRequest struct {
	Owner string `json:"owner"`
}

// NftMethod is a wrapper type for all the methods used in smart contract.
type NftMethod string

// all the methods used for smart contract.
const (
	// transactions.
	NftMethodMint                NftMethod = "mint"
	NftMethodBurn                NftMethod = "burn"
	NftMethodFreeze              NftMethod = "freeze"
	NftMethodUnfreeze            NftMethod = "unfreeze"
	NftMethodAddToWhitelist      NftMethod = "add_to_whitelist"
	NftMethodRemoveFromWhiteList NftMethod = "remove_from_whitelist"
	NftMethodSend                NftMethod = "send"
	// queries.
	NftMethodParams                    NftMethod = "params"
	NftMethodClass                     NftMethod = "class"
	NftMethodClasses                   NftMethod = "classes"
	NftMethodFrozen                    NftMethod = "frozen"
	NftMethodWhitelisted               NftMethod = "whitelisted"
	NftMethodWhitelistedAccountsForNft NftMethod = "whitelisted_accounts_for_nft"
	NftMethodBalance                   NftMethod = "balance"
	NftMethodOwner                     NftMethod = "owner"
	NftMethodSupply                    NftMethod = "supply"
	NftMethodNFT                       NftMethod = "nft"
	NftMethodNFTs                      NftMethod = "nfts"
	NftMethodClassNFT                  NftMethod = "class_nft"
	NftMethodClassesNFT                NftMethod = "classes_nft"
)

// AssetnftClass represents the Class in asset nft module.
//
//nolint:tagliatelle
type AssetnftClass struct {
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

// AssetnftClassResponse is returned when querying for class info.
type AssetnftClassResponse struct {
	Class AssetnftClass `json:"class"`
}

// NftItem is represents the NFT returned from smart contract.
//
//nolint:tagliatelle
type NftItem struct {
	ClassID string `json:"class_id"`
	ID      string `json:"id"`
	URI     string `json:"uri"`
	URIHash string `json:"uri_hash"`
	Data    string `json:"data"`
}

// NftRes is returned when querying for the NFT.
type NftRes struct {
	NFT NftItem `json:"nft"`
}

// PageResponse represents pagination response for listings.
//
//nolint:tagliatelle
type PageResponse struct {
	NextKey []byte `json:"next_key"`
	Total   uint64 `json:"total"`
}

// NftsRes is used to return a list of NFTs.
type NftsRes struct {
	NFTs       []NftItem    `json:"nfts"`
	Pagination PageResponse `json:"pagination"`
}

// NftClass returns class info.
//
//nolint:tagliatelle
type NftClass struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Description string `json:"description"`
	URI         string `json:"uri"`
	URIHash     string `json:"uri_hash"`
	Data        string `json:"data"`
}

// NftClassResponse is the response returned when querying for class info.
type NftClassResponse struct {
	Class NftClass `json:"class"`
}

// NftClassesResponse is the response returned when querying for list of class info.
type NftClassesResponse struct {
	Classes    []NftClass   `json:"classes"`
	Pagination PageResponse `json:"pagination"`
}
