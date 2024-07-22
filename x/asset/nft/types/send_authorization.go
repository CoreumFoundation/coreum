package types

import (
	context "context"

	"cosmossdk.io/x/nft"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

var _ authz.Authorization = &SendAuthorization{}

// NewSendAuthorization returns a new SendAuthorization object.
func NewSendAuthorization(nfts []NFTIdentifier) *SendAuthorization {
	return &SendAuthorization{
		Nfts: nfts,
	}
}

// MsgTypeURL implements Authorization.MsgTypeURL.
func (a SendAuthorization) MsgTypeURL() string {
	return sdk.MsgTypeURL(&nft.MsgSend{})
}

// Accept implements Authorization.Accept.
func (a SendAuthorization) Accept(ctx context.Context, msg sdk.Msg) (authz.AcceptResponse, error) {
	mSend, ok := msg.(*nft.MsgSend)
	if !ok {
		return authz.AcceptResponse{}, sdkerrors.ErrInvalidType.Wrap("type mismatch")
	}

	exists := a.findAndRemoveNFT(mSend.ClassId, mSend.Id)
	if !exists {
		return authz.AcceptResponse{}, sdkerrors.ErrUnauthorized.Wrapf("requested NFT does not have transfer grant")
	}

	return authz.AcceptResponse{
		Accept:  true,
		Delete:  len(a.Nfts) == 0,
		Updated: &a,
	}, nil
}

// ValidateBasic implements Authorization.ValidateBasic.
func (a SendAuthorization) ValidateBasic() error {
	if len(a.Nfts) == 0 {
		return ErrInvalidInput.Wrap("empty NFT list")
	}

	for _, nft := range a.Nfts {
		if err := ValidateTokenID(nft.Id); err != nil {
			return ErrInvalidInput.Wrap(err.Error())
		}

		if _, _, err := DeconstructClassID(nft.ClassId); err != nil {
			return ErrInvalidInput.Wrap(err.Error())
		}
	}

	return nil
}

func (a *SendAuthorization) findAndRemoveNFT(classID, nftID string) bool {
	for index, nft := range a.Nfts {
		if nft.ClassId == classID && nft.Id == nftID {
			a.Nfts = append(a.Nfts[:index], a.Nfts[index+1:]...)
			return true
		}
	}
	return false
}
