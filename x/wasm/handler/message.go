package handler

import (
	"encoding/json"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	assetfttypes "github.com/CoreumFoundation/coreum/v2/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/v2/x/asset/nft/types"
	nfttypes "github.com/CoreumFoundation/coreum/v2/x/nft"
)

// assetFTMsg represents asset ft module messages integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type assetFTMsg struct {
	Issue               *assetfttypes.MsgIssue               `json:"Issue"`
	Mint                *assetfttypes.MsgMint                `json:"Mint"`
	Burn                *assetfttypes.MsgBurn                `json:"Burn"`
	Freeze              *assetfttypes.MsgFreeze              `json:"Freeze"`
	Unfreeze            *assetfttypes.MsgUnfreeze            `json:"Unfreeze"`
	GloballyFreeze      *assetfttypes.MsgGloballyFreeze      `json:"GloballyFreeze"`
	GloballyUnfreeze    *assetfttypes.MsgGloballyUnfreeze    `json:"GloballyUnfreeze"`
	SetWhitelistedLimit *assetfttypes.MsgSetWhitelistedLimit `json:"SetWhitelistedLimit"`
	UpgradeTokenV1      *assetfttypes.MsgUpgradeTokenV1      `json:"UpgradeTokenV1"`
}

// assetNFTMsgIssueClass defines message for the IssueClass method with string represented data field.
//
//nolint:tagliatelle // we keep the name same as consume
type assetNFTMsgIssueClass struct {
	Symbol      string                       `json:"symbol"`
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	URI         string                       `json:"uri"`
	URIHash     string                       `json:"uri_hash"`
	Data        string                       `json:"data"`
	Features    []assetnfttypes.ClassFeature `json:"features"`
	RoyaltyRate sdk.Dec                      `json:"royalty_rate"`
}

// assetNFTMsgMint defines message for the Mint method with string represented data field.
//
//nolint:tagliatelle // we keep the name same as consume
type assetNFTMsgMint struct {
	ClassID string `json:"class_id"`
	ID      string `json:"id"`
	URI     string `json:"uri"`
	URIHash string `json:"uri_hash"`
	Data    string `json:"data"`
}

// assetNFTMsg represents asset nft module messages integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type assetNFTMsg struct {
	IssueClass          *assetNFTMsgIssueClass                `json:"IssueClass"`
	Mint                *assetNFTMsgMint                      `json:"Mint"`
	Burn                *assetnfttypes.MsgBurn                `json:"Burn"`
	Freeze              *assetnfttypes.MsgFreeze              `json:"Freeze"`
	Unfreeze            *assetnfttypes.MsgUnfreeze            `json:"Unfreeze"`
	AddToWhitelist      *assetnfttypes.MsgAddToWhitelist      `json:"AddToWhitelist"`
	RemoveFromWhitelist *assetnfttypes.MsgRemoveFromWhitelist `json:"RemoveFromWhitelist"`
}

// nftMsg represents nft module messages integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type nftMsg struct {
	Send *nfttypes.MsgSend `json:"Send"`
}

// coreumMsg represents all supported custom messages integrated with the wasm handler.
//
//nolint:tagliatelle // we keep the name same as consume
type coreumMsg struct {
	AssetFT  *assetFTMsg  `json:"AssetFT"`
	AssetNFT *assetNFTMsg `json:"AssetNFT"`
	NFT      *nftMsg      `json:"nft"`
}

// NewCoreumMsgHandler returns coreum handler that handles messages received from smart contracts.
// The in the input sender is the address of smart contract.
func NewCoreumMsgHandler() *wasmkeeper.MessageEncoders {
	return &wasmkeeper.MessageEncoders{
		Custom: func(sender sdk.AccAddress, msg json.RawMessage) ([]sdk.Msg, error) {
			var coreumMsg coreumMsg
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

func decodeCoreumMessage(coreumMessages coreumMsg, sender sdk.AccAddress) (sdk.Msg, error) {
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

func decodeAssetFTMessage(assetFTMsg *assetFTMsg, sender string) (sdk.Msg, error) {
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
	if assetFTMsg.UpgradeTokenV1 != nil {
		assetFTMsg.UpgradeTokenV1.Sender = sender
		return assetFTMsg.UpgradeTokenV1, nil
	}

	return nil, nil
}

func decodeAssetNFTMessage(assetNFTMsg *assetNFTMsg, sender string) (sdk.Msg, error) {
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

func decodeNFTMessage(nftMsg *nftMsg, sender string) (sdk.Msg, error) {
	if nftMsg.Send != nil {
		nftMsg.Send.Sender = sender
		return nftMsg.Send, nil
	}

	return nil, nil
}
