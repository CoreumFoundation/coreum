package modules

import (
	assetnfttypes "github.com/CoreumFoundation/coreum/x/asset/nft/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

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

//nolint:tagliatelle
type NftMintRequest struct {
	ID      string `json:"id"`
	URI     string `json:"uri"`
	URIHash string `json:"uri_hash"`
	Data    string `json:"data"`
}

type NftIDRequest struct {
	ID string `json:"id"`
}

type NftIssuerRequest struct {
	Issuer string `json:"issuer"`
}

type NftIDWithAccountRequest struct {
	ID      string `json:"id"`
	Account string `json:"account"`
}

type NftIDWithReceiverRequest struct {
	ID       string `json:"id"`
	Receiver string `json:"receiver"`
}

type NftOwnerRequest struct {
	Owner string `json:"owner"`
}

type NftMethod string

const (
	// tx.
	NftMethodMint                NftMethod = "mint"
	NftMethodBurn                NftMethod = "burn"
	NftMethodFreeze              NftMethod = "freeze"
	NftMethodUnfreeze            NftMethod = "unfreeze"
	NftMethodAddToWhitelist      NftMethod = "add_to_whitelist"
	NftMethodRemoveFromWhiteList NftMethod = "remove_from_whitelist"
	NftMethodSend                NftMethod = "send"
	// query.
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

type AssetnftClassResponse struct {
	Class AssetnftClass `json:"class"`
}

//nolint:tagliatelle
type NftItem struct {
	ClassID string `json:"class_id"`
	ID      string `json:"id"`
	URI     string `json:"uri"`
	URIHash string `json:"uri_hash"`
	Data    string `json:"data"`
}

type NftRes struct {
	NFT NftItem `json:"nft"`
}

//nolint:tagliatelle
type PageResponse struct {
	NextKey []byte `json:"next_key"`
	Total   uint64 `json:"total"`
}

type NftsRes struct {
	NFTs       []NftItem    `json:"nfts"`
	Pagination PageResponse `json:"pagination"`
}

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

type NftClassResponse struct {
	Class NftClass `json:"class"`
}

type NftClassesResponse struct {
	Classes    []NftClass   `json:"classes"`
	Pagination PageResponse `json:"pagination"`
}
