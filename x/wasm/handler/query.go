package handler

import (
	"context"
	"encoding/base64"
	"encoding/json"

	sdkmath "cosmossdk.io/math"
	nfttypes "cosmossdk.io/x/nft"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/pkg/errors"

	assetfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/v5/x/asset/nft/types"
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
	RoyaltyRate sdkmath.LegacyDec            `json:"royalty_rate"`
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
	ClassFrozen               *assetnfttypes.QueryClassFrozenRequest               `json:"ClassFrozen"`
	ClassFrozenAccounts       *assetnfttypes.QueryClassFrozenAccountsRequest       `json:"ClassFrozenAccounts"`
	Whitelisted               *assetnfttypes.QueryWhitelistedRequest               `json:"Whitelisted"`
	WhitelistedAccountsforNFT *assetnfttypes.QueryWhitelistedAccountsForNFTRequest `json:"WhitelistedAccountsforNft"`
	ClassWhitelistedAccounts  *assetnfttypes.QueryClassWhitelistedAccountsRequest  `json:"ClassWhitelistedAccounts"`
	BurntNFT                  *assetnfttypes.QueryBurntNFTRequest                  `json:"BurntNft"`
	BurntNFTsInClass          *assetnfttypes.QueryBurntNFTsInClassRequest          `json:"BurntNftsInClass"`
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
//
//nolint:lll // access list is long
func NewCoreumQueryHandler(assetFTQueryServer assetfttypes.QueryServer, assetNFTQueryServer assetnfttypes.QueryServer, nftQueryServer nfttypes.QueryServer, gRPCQueryRouter *baseapp.GRPCQueryRouter, codec *codec.ProtoCodec) *wasmkeeper.QueryPlugins {
	acceptList := wasmkeeper.AcceptedQueries{
		"/coreum.asset.ft.v1.Query/Token":              &assetfttypes.QueryTokenResponse{},
		"/coreum.asset.ft.v1.Query/FrozenBalance":      &assetfttypes.QueryFrozenBalanceResponse{},
		"/coreum.asset.ft.v1.Query/WhitelistedBalance": &assetfttypes.QueryWhitelistedBalanceResponse{},
		"/cosmos.bank.v1beta1.Query/Balance":           &banktypes.QueryBalanceResponse{},
	}
	/*
		TODO: Other potential queries to add to the whitelist:
		/coreum.asset.ft.v1.Query/Balance
		/coreum.asset.ft.v1.Query/DEXSettings
		/coreum.asset.ft.v1.Query/FrozenBalance
		/coreum.asset.ft.v1.Query/FrozenBalances
		/coreum.asset.ft.v1.Query/Params
		/coreum.asset.ft.v1.Query/Token
		/coreum.asset.ft.v1.Query/Tokens
		/coreum.asset.ft.v1.Query/TokenUpgradeStatuses
		/coreum.asset.ft.v1.Query/WhitelistedBalance
		/coreum.asset.ft.v1.Query/WhitelistedBalances
		/coreum.asset.nft.v1.Query/BurntNFT
		/coreum.asset.nft.v1.Query/BurntNFTsInClass
		/coreum.asset.nft.v1.Query/Class
		/coreum.asset.nft.v1.Query/Classes
		/coreum.asset.nft.v1.Query/ClassFrozen
		/coreum.asset.nft.v1.Query/ClassFrozenAccounts
		/coreum.asset.nft.v1.Query/ClassWhitelistedAccounts
		/coreum.asset.nft.v1.Query/Frozen
		/coreum.asset.nft.v1.Query/Params
		/coreum.asset.nft.v1.Query/Whitelisted
		/coreum.asset.nft.v1.Query/WhitelistedAccountsForNFT
		/coreum.customparams.v1.Query/StakingParams
		/coreum.dex.v1.Query/AccountDenomOrdersCount
		/coreum.dex.v1.Query/Order
		/coreum.dex.v1.Query/OrderBooks
		/coreum.dex.v1.Query/Orders
		/coreum.dex.v1.Query/OrdersBookOrders
		/coreum.dex.v1.Query/Params
		/coreum.feemodel.v1.Query/MinGasPrice
		/coreum.feemodel.v1.Query/Params
		/coreum.feemodel.v1.Query/RecommendedGasPrice
		/cosmos.auth.v1beta1.Query/Account
		/cosmos.auth.v1beta1.Query/AccountAddressByID
		/cosmos.auth.v1beta1.Query/AccountInfo
		/cosmos.auth.v1beta1.Query/Accounts
		/cosmos.auth.v1beta1.Query/AddressBytesToString
		/cosmos.auth.v1beta1.Query/AddressStringToBytes
		/cosmos.auth.v1beta1.Query/Bech32Prefix
		/cosmos.auth.v1beta1.Query/ModuleAccountByName
		/cosmos.auth.v1beta1.Query/ModuleAccounts
		/cosmos.auth.v1beta1.Query/Params
		/cosmos.authz.v1beta1.Query/GranteeGrants
		/cosmos.authz.v1beta1.Query/GranterGrants
		/cosmos.authz.v1beta1.Query/Grants
		/cosmos.autocli.v1.Query/AppOptions
		/cosmos.bank.v1beta1.Query/AllBalances
		/cosmos.bank.v1beta1.Query/Balance
		/cosmos.bank.v1beta1.Query/DenomMetadata
		/cosmos.bank.v1beta1.Query/DenomMetadataByQueryString
		/cosmos.bank.v1beta1.Query/DenomOwners
		/cosmos.bank.v1beta1.Query/DenomOwnersByQuery
		/cosmos.bank.v1beta1.Query/DenomsMetadata
		/cosmos.bank.v1beta1.Query/Params
		/cosmos.bank.v1beta1.Query/SendEnabled
		/cosmos.bank.v1beta1.Query/SpendableBalanceByDenom
		/cosmos.bank.v1beta1.Query/SpendableBalances
		/cosmos.bank.v1beta1.Query/SupplyOf
		/cosmos.bank.v1beta1.Query/TotalSupply
		/cosmos.base.node.v1beta1.Service/Config
		/cosmos.base.node.v1beta1.Service/Status
		/cosmos.base.reflection.v1beta1.ReflectionService/ListAllInterfaces
		/cosmos.base.reflection.v1beta1.ReflectionService/ListImplementations
		/cosmos.base.tendermint.v1beta1.Service/ABCIQuery
		/cosmos.base.tendermint.v1beta1.Service/GetBlockByHeight
		/cosmos.base.tendermint.v1beta1.Service/GetLatestBlock
		/cosmos.base.tendermint.v1beta1.Service/GetLatestValidatorSet
		/cosmos.base.tendermint.v1beta1.Service/GetNodeInfo
		/cosmos.base.tendermint.v1beta1.Service/GetSyncing
		/cosmos.base.tendermint.v1beta1.Service/GetValidatorSetByHeight
		/cosmos.consensus.v1.Query/Params
		/cosmos.distribution.v1beta1.Query/CommunityPool
		/cosmos.distribution.v1beta1.Query/DelegationRewards
		/cosmos.distribution.v1beta1.Query/DelegationTotalRewards
		/cosmos.distribution.v1beta1.Query/DelegatorValidators
		/cosmos.distribution.v1beta1.Query/DelegatorWithdrawAddress
		/cosmos.distribution.v1beta1.Query/Params
		/cosmos.distribution.v1beta1.Query/ValidatorCommission
		/cosmos.distribution.v1beta1.Query/ValidatorDistributionInfo
		/cosmos.distribution.v1beta1.Query/ValidatorOutstandingRewards
		/cosmos.distribution.v1beta1.Query/ValidatorSlashes
		/cosmos.evidence.v1beta1.Query/AllEvidence
		/cosmos.evidence.v1beta1.Query/Evidence
		/cosmos.feegrant.v1beta1.Query/Allowance
		/cosmos.feegrant.v1beta1.Query/Allowances
		/cosmos.feegrant.v1beta1.Query/AllowancesByGranter
		/cosmos.gov.v1.Query/Constitution
		/cosmos.gov.v1.Query/Deposit
		/cosmos.gov.v1.Query/Deposits
		/cosmos.gov.v1.Query/Params
		/cosmos.gov.v1.Query/Proposal
		/cosmos.gov.v1.Query/Proposals
		/cosmos.gov.v1.Query/TallyResult
		/cosmos.gov.v1.Query/Vote
		/cosmos.gov.v1.Query/Votes
		/cosmos.gov.v1beta1.Query/Deposit
		/cosmos.gov.v1beta1.Query/Deposits
		/cosmos.gov.v1beta1.Query/Params
		/cosmos.gov.v1beta1.Query/Proposal
		/cosmos.gov.v1beta1.Query/Proposals
		/cosmos.gov.v1beta1.Query/TallyResult
		/cosmos.gov.v1beta1.Query/Vote
		/cosmos.gov.v1beta1.Query/Votes
		/cosmos.group.v1.Query/GroupInfo
		/cosmos.group.v1.Query/GroupMembers
		/cosmos.group.v1.Query/GroupPoliciesByAdmin
		/cosmos.group.v1.Query/GroupPoliciesByGroup
		/cosmos.group.v1.Query/GroupPolicyInfo
		/cosmos.group.v1.Query/Groups
		/cosmos.group.v1.Query/GroupsByAdmin
		/cosmos.group.v1.Query/GroupsByMember
		/cosmos.group.v1.Query/Proposal
		/cosmos.group.v1.Query/ProposalsByGroupPolicy
		/cosmos.group.v1.Query/TallyResult
		/cosmos.group.v1.Query/VoteByProposalVoter
		/cosmos.group.v1.Query/VotesByProposal
		/cosmos.group.v1.Query/VotesByVoter
		/cosmos.mint.v1beta1.Query/AnnualProvisions
		/cosmos.mint.v1beta1.Query/Inflation
		/cosmos.mint.v1beta1.Query/Params
		/cosmos.nft.v1beta1.Query/Balance
		/cosmos.nft.v1beta1.Query/Class
		/cosmos.nft.v1beta1.Query/Classes
		/cosmos.nft.v1beta1.Query/NFT
		/cosmos.nft.v1beta1.Query/NFTs
		/cosmos.nft.v1beta1.Query/Owner
		/cosmos.nft.v1beta1.Query/Supply
		/cosmos.params.v1beta1.Query/Params
		/cosmos.params.v1beta1.Query/Subspaces
		/cosmos.reflection.v1.ReflectionService/FileDescriptors
		/cosmos.slashing.v1beta1.Query/Params
		/cosmos.slashing.v1beta1.Query/SigningInfo
		/cosmos.slashing.v1beta1.Query/SigningInfos
		/cosmos.staking.v1beta1.Query/Delegation
		/cosmos.staking.v1beta1.Query/DelegatorDelegations
		/cosmos.staking.v1beta1.Query/DelegatorUnbondingDelegations
		/cosmos.staking.v1beta1.Query/DelegatorValidator
		/cosmos.staking.v1beta1.Query/DelegatorValidators
		/cosmos.staking.v1beta1.Query/HistoricalInfo
		/cosmos.staking.v1beta1.Query/Params
		/cosmos.staking.v1beta1.Query/Pool
		/cosmos.staking.v1beta1.Query/Redelegations
		/cosmos.staking.v1beta1.Query/UnbondingDelegation
		/cosmos.staking.v1beta1.Query/Validator
		/cosmos.staking.v1beta1.Query/ValidatorDelegations
		/cosmos.staking.v1beta1.Query/Validators
		/cosmos.staking.v1beta1.Query/ValidatorUnbondingDelegations
		/cosmos.tx.v1beta1.Service/BroadcastTx
		/cosmos.tx.v1beta1.Service/GetBlockWithTxs
		/cosmos.tx.v1beta1.Service/GetTx
		/cosmos.tx.v1beta1.Service/GetTxsEvent
		/cosmos.tx.v1beta1.Service/Simulate
		/cosmos.tx.v1beta1.Service/TxDecode
		/cosmos.tx.v1beta1.Service/TxDecodeAmino
		/cosmos.tx.v1beta1.Service/TxEncode
		/cosmos.tx.v1beta1.Service/TxEncodeAmino
		/cosmos.upgrade.v1beta1.Query/AppliedPlan
		/cosmos.upgrade.v1beta1.Query/Authority
		/cosmos.upgrade.v1beta1.Query/CurrentPlan
		/cosmos.upgrade.v1beta1.Query/ModuleVersions
		/cosmos.upgrade.v1beta1.Query/UpgradedConsensusState
		/cosmwasm.wasm.v1.Query/AllContractState
		/cosmwasm.wasm.v1.Query/BuildAddress
		/cosmwasm.wasm.v1.Query/Code
		/cosmwasm.wasm.v1.Query/Codes
		/cosmwasm.wasm.v1.Query/ContractHistory
		/cosmwasm.wasm.v1.Query/ContractInfo
		/cosmwasm.wasm.v1.Query/ContractsByCode
		/cosmwasm.wasm.v1.Query/ContractsByCreator
		/cosmwasm.wasm.v1.Query/Params
		/cosmwasm.wasm.v1.Query/PinnedCodes
		/cosmwasm.wasm.v1.Query/RawContractState
		/cosmwasm.wasm.v1.Query/SmartContractState
		/ibc.applications.interchain_accounts.controller.v1.Query/InterchainAccount
		/ibc.applications.interchain_accounts.controller.v1.Query/Params
		/ibc.applications.interchain_accounts.host.v1.Query/Params
		/ibc.applications.transfer.v1.Query/DenomHash
		/ibc.applications.transfer.v1.Query/DenomTrace
		/ibc.applications.transfer.v1.Query/DenomTraces
		/ibc.applications.transfer.v1.Query/EscrowAddress
		/ibc.applications.transfer.v1.Query/Params
		/ibc.applications.transfer.v1.Query/TotalEscrowForDenom
		/ibc.core.channel.v1.Query/Channel
		/ibc.core.channel.v1.Query/ChannelClientState
		/ibc.core.channel.v1.Query/ChannelConsensusState
		/ibc.core.channel.v1.Query/ChannelParams
		/ibc.core.channel.v1.Query/Channels
		/ibc.core.channel.v1.Query/ConnectionChannels
		/ibc.core.channel.v1.Query/NextSequenceReceive
		/ibc.core.channel.v1.Query/NextSequenceSend
		/ibc.core.channel.v1.Query/PacketAcknowledgement
		/ibc.core.channel.v1.Query/PacketAcknowledgements
		/ibc.core.channel.v1.Query/PacketCommitment
		/ibc.core.channel.v1.Query/PacketCommitments
		/ibc.core.channel.v1.Query/PacketReceipt
		/ibc.core.channel.v1.Query/UnreceivedAcks
		/ibc.core.channel.v1.Query/UnreceivedPackets
		/ibc.core.channel.v1.Query/Upgrade
		/ibc.core.channel.v1.Query/UpgradeError
		/ibc.core.client.v1.Query/ClientParams
		/ibc.core.client.v1.Query/ClientState
		/ibc.core.client.v1.Query/ClientStates
		/ibc.core.client.v1.Query/ClientStatus
		/ibc.core.client.v1.Query/ConsensusState
		/ibc.core.client.v1.Query/ConsensusStateHeights
		/ibc.core.client.v1.Query/ConsensusStates
		/ibc.core.client.v1.Query/UpgradedClientState
		/ibc.core.client.v1.Query/UpgradedConsensusState
		/ibc.core.client.v1.Query/VerifyMembership
		/ibc.core.connection.v1.Query/ClientConnections
		/ibc.core.connection.v1.Query/Connection
		/ibc.core.connection.v1.Query/ConnectionClientState
		/ibc.core.connection.v1.Query/ConnectionConsensusState
		/ibc.core.connection.v1.Query/ConnectionParams
		/ibc.core.connection.v1.Query/Connections
		/packetforward.v1.Query/Params
	*/
	return &wasmkeeper.QueryPlugins{
		Grpc: wasmkeeper.AcceptListGrpcQuerier(acceptList, gRPCQueryRouter, codec),
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

func processAssetFTQuery(
	ctx sdk.Context, assetFTQuery *assetFTQuery, assetFTQueryServer assetfttypes.QueryServer,
) ([]byte, error) {
	if assetFTQuery.Params != nil {
		return executeQuery(
			ctx,
			assetFTQuery.Params,
			func(ctx context.Context, req *assetfttypes.QueryParamsRequest) (*assetfttypes.QueryParamsResponse, error) {
				return assetFTQueryServer.Params(ctx, req)
			},
		)
	}
	if assetFTQuery.Token != nil {
		return executeQuery(
			ctx,
			assetFTQuery.Token,
			func(ctx context.Context, req *assetfttypes.QueryTokenRequest) (*assetfttypes.QueryTokenResponse, error) {
				return assetFTQueryServer.Token(ctx, req)
			},
		)
	}
	if assetFTQuery.Tokens != nil {
		return executeQuery(
			ctx,
			assetFTQuery.Tokens,
			func(ctx context.Context, req *assetfttypes.QueryTokensRequest) (*assetfttypes.QueryTokensResponse, error) {
				return assetFTQueryServer.Tokens(ctx, req)
			},
		)
	}
	if assetFTQuery.Balance != nil {
		return executeQuery(
			ctx,
			assetFTQuery.Balance,
			func(ctx context.Context, req *assetfttypes.QueryBalanceRequest) (*assetfttypes.QueryBalanceResponse, error) {
				return assetFTQueryServer.Balance(ctx, req)
			},
		)
	}
	if assetFTQuery.FrozenBalance != nil {
		return executeQuery(
			ctx,
			assetFTQuery.FrozenBalance,
			func(
				ctx context.Context, req *assetfttypes.QueryFrozenBalanceRequest,
			) (*assetfttypes.QueryFrozenBalanceResponse, error) {
				return assetFTQueryServer.FrozenBalance(ctx, req)
			},
		)
	}
	if assetFTQuery.FrozenBalances != nil {
		return executeQuery(
			ctx,
			assetFTQuery.FrozenBalances,
			func(
				ctx context.Context, req *assetfttypes.QueryFrozenBalancesRequest,
			) (*assetfttypes.QueryFrozenBalancesResponse, error) {
				return assetFTQueryServer.FrozenBalances(ctx, req)
			},
		)
	}
	if assetFTQuery.WhitelistedBalance != nil {
		return executeQuery(
			ctx,
			assetFTQuery.WhitelistedBalance,
			func(
				ctx context.Context, req *assetfttypes.QueryWhitelistedBalanceRequest,
			) (*assetfttypes.QueryWhitelistedBalanceResponse, error) {
				return assetFTQueryServer.WhitelistedBalance(ctx, req)
			},
		)
	}
	if assetFTQuery.WhitelistedBalances != nil {
		return executeQuery(
			ctx,
			assetFTQuery.WhitelistedBalances,
			func(
				ctx context.Context, req *assetfttypes.QueryWhitelistedBalancesRequest,
			) (*assetfttypes.QueryWhitelistedBalancesResponse, error) {
				return assetFTQueryServer.WhitelistedBalances(ctx, req)
			},
		)
	}

	return nil, nil
}

//nolint:funlen
func processAssetNFTQuery(
	ctx sdk.Context,
	assetNFTQuery *assetNFTQuery,
	assetNFTQueryServer assetnfttypes.QueryServer,
) ([]byte, error) {
	if assetNFTQuery.Params != nil {
		return executeQuery(
			ctx,
			assetNFTQuery.Params,
			func(
				ctx context.Context, req *assetnfttypes.QueryParamsRequest,
			) (*assetnfttypes.QueryParamsResponse, error) {
				return assetNFTQueryServer.Params(ctx, req)
			},
		)
	}
	if assetNFTQuery.Class != nil {
		return executeQuery(
			ctx,
			assetNFTQuery.Class,
			func(ctx context.Context, req *assetnfttypes.QueryClassRequest) (*assetNFTClassResponse, error) {
				classRes, err := assetNFTQueryServer.Class(ctx, req)
				if err != nil {
					return nil, err
				}

				var dataString string
				if classRes.Class.Data != nil {
					dataString, err = unmarshalData(classRes.Class.Data)
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
			},
		)
	}
	if assetNFTQuery.Classes != nil {
		return executeQuery(
			ctx,
			assetNFTQuery.Classes,
			func(ctx context.Context, req *assetnfttypes.QueryClassesRequest) (*assetNFTClassesResponse, error) {
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
						dataString, err = unmarshalData(classesRes.Classes[i].Data)
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
			},
		)
	}

	if assetNFTQuery.Frozen != nil {
		return executeQuery(
			ctx,
			assetNFTQuery.Frozen,
			func(ctx context.Context, req *assetnfttypes.QueryFrozenRequest) (*assetnfttypes.QueryFrozenResponse, error) {
				return assetNFTQueryServer.Frozen(ctx, req)
			},
		)
	}
	if assetNFTQuery.ClassFrozen != nil {
		return executeQuery(
			ctx, assetNFTQuery.ClassFrozen,
			func(
				ctx context.Context, req *assetnfttypes.QueryClassFrozenRequest,
			) (*assetnfttypes.QueryClassFrozenResponse, error) {
				return assetNFTQueryServer.ClassFrozen(ctx, req)
			},
		)
	}
	if assetNFTQuery.ClassFrozenAccounts != nil {
		return executeQuery(
			ctx,
			assetNFTQuery.ClassFrozenAccounts,
			func(
				ctx context.Context, req *assetnfttypes.QueryClassFrozenAccountsRequest,
			) (*assetnfttypes.QueryClassFrozenAccountsResponse, error) {
				return assetNFTQueryServer.ClassFrozenAccounts(ctx, req)
			})
	}
	if assetNFTQuery.Whitelisted != nil {
		return executeQuery(
			ctx,
			assetNFTQuery.Whitelisted,
			func(
				ctx context.Context, req *assetnfttypes.QueryWhitelistedRequest,
			) (*assetnfttypes.QueryWhitelistedResponse, error) {
				return assetNFTQueryServer.Whitelisted(ctx, req)
			},
		)
	}
	if assetNFTQuery.WhitelistedAccountsforNFT != nil {
		return executeQuery(
			ctx,
			assetNFTQuery.WhitelistedAccountsforNFT,
			func(
				ctx context.Context, req *assetnfttypes.QueryWhitelistedAccountsForNFTRequest,
			) (*assetnfttypes.QueryWhitelistedAccountsForNFTResponse, error) {
				return assetNFTQueryServer.WhitelistedAccountsForNFT(ctx, req)
			},
		)
	}
	if assetNFTQuery.ClassWhitelistedAccounts != nil {
		return executeQuery(
			ctx,
			assetNFTQuery.ClassWhitelistedAccounts,
			func(
				ctx context.Context, req *assetnfttypes.QueryClassWhitelistedAccountsRequest,
			) (*assetnfttypes.QueryClassWhitelistedAccountsResponse, error) {
				return assetNFTQueryServer.ClassWhitelistedAccounts(ctx, req)
			},
		)
	}
	if assetNFTQuery.BurntNFT != nil {
		return executeQuery(
			ctx,
			assetNFTQuery.BurntNFT,
			func(
				ctx context.Context, req *assetnfttypes.QueryBurntNFTRequest,
			) (*assetnfttypes.QueryBurntNFTResponse, error) {
				return assetNFTQueryServer.BurntNFT(ctx, req)
			},
		)
	}

	if assetNFTQuery.BurntNFTsInClass != nil {
		return executeQuery(
			ctx,
			assetNFTQuery.BurntNFTsInClass,
			func(
				ctx context.Context, req *assetnfttypes.QueryBurntNFTsInClassRequest,
			) (*assetnfttypes.QueryBurntNFTsInClassResponse, error) {
				return assetNFTQueryServer.BurntNFTsInClass(ctx, req)
			},
		)
	}

	return nil, nil
}

//nolint:funlen
func processNFTQuery(ctx sdk.Context, nftQuery *nftQuery, nftQueryServer nfttypes.QueryServer) ([]byte, error) {
	if nftQuery.Balance != nil {
		return executeQuery(
			ctx,
			nftQuery.Balance,
			func(ctx context.Context, req *nfttypes.QueryBalanceRequest) (*nfttypes.QueryBalanceResponse, error) {
				return nftQueryServer.Balance(ctx, req)
			},
		)
	}
	if nftQuery.Owner != nil {
		return executeQuery(
			ctx,
			nftQuery.Owner,
			func(ctx context.Context, req *nfttypes.QueryOwnerRequest) (*nfttypes.QueryOwnerResponse, error) {
				return nftQueryServer.Owner(ctx, req)
			},
		)
	}
	if nftQuery.Supply != nil {
		return executeQuery(
			ctx,
			nftQuery.Supply,
			func(ctx context.Context, req *nfttypes.QuerySupplyRequest) (*nfttypes.QuerySupplyResponse, error) {
				return nftQueryServer.Supply(ctx, req)
			},
		)
	}
	if nftQuery.NFT != nil { //nolint:nestif // the ifs are for the error checks mostly
		return executeQuery(
			ctx,
			nftQuery.NFT,
			func(ctx context.Context, req *nfttypes.QueryNFTRequest) (*NFTResponse, error) {
				nftRes, err := nftQueryServer.NFT(ctx, req)
				if err != nil {
					return nil, err
				}

				if nftRes.Nft == nil {
					return &NFTResponse{}, nil
				}

				var dataString string
				if nftRes.Nft.Data != nil {
					dataString, err = unmarshalData(nftRes.Nft.Data)
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
			},
		)
	}
	if nftQuery.NFTs != nil { //nolint:nestif // the ifs are for the error checks mostly
		return executeQuery(
			ctx,
			nftQuery.NFTs,
			func(ctx context.Context, req *nfttypes.QueryNFTsRequest) (*NFTsResponse, error) {
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
						dataString, err = unmarshalData(nftsRes.Nfts[i].Data)
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
			},
		)
	}
	if nftQuery.Class != nil { //nolint:nestif // the ifs are for the error checks mostly
		return executeQuery(
			ctx,
			nftQuery.Class,
			func(ctx context.Context, req *nfttypes.QueryClassRequest) (*NFTClassResponse, error) {
				nftClassRes, err := nftQueryServer.Class(ctx, req)
				if err != nil {
					return nil, err
				}

				if nftClassRes.Class == nil {
					return &NFTClassResponse{}, nil
				}

				var dataString string
				if nftClassRes.Class.Data != nil {
					dataString, err = unmarshalData(nftClassRes.Class.Data)
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
			},
		)
	}
	if nftQuery.Classes != nil { //nolint:nestif // the ifs are for the error checks mostly
		return executeQuery(
			ctx,
			nftQuery.Classes,
			func(ctx context.Context, req *nfttypes.QueryClassesRequest) (*NFTClassesResponse, error) {
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
						dataString, err = unmarshalData(nftClassesRes.Classes[i].Data)
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
			},
		)
	}

	return nil, nil
}

func executeQuery[T, K any](
	ctx sdk.Context,
	reqStruct T,
	reqExecutor func(ctx context.Context, req T) (K, error),
) (json.RawMessage, error) {
	res, err := reqExecutor(ctx, reqStruct)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	raw, err := json.Marshal(res)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return raw, nil
}

func unmarshalData(data *codectypes.Any) (string, error) {
	switch data.TypeUrl {
	case "/" + proto.MessageName((*assetnfttypes.DataBytes)(nil)):
		var datab assetnfttypes.DataBytes
		err := proto.Unmarshal(data.Value, &datab)
		if err != nil {
			return "", errors.WithStack(err)
		}
		return base64.StdEncoding.EncodeToString(datab.Data), nil
	case "/" + proto.MessageName((*assetnfttypes.DataDynamic)(nil)):
		var datadynamic assetnfttypes.DataDynamic
		err := proto.Unmarshal(data.Value, &datadynamic)
		if err != nil {
			return "", errors.WithStack(err)
		}
		bytes, err := datadynamic.Marshal()
		if err != nil {
			return "", errors.WithStack(err)
		}
		return base64.StdEncoding.EncodeToString(bytes), nil
	default:
		return "", errors.Errorf("unsupported data type %s", data.TypeUrl)
	}
}
