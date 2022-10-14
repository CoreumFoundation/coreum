package keeper_test

import (
	"errors"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestKeeper_IssueFungibleToken(t *testing.T) {
	requireT := require.New(t)

	testApp := simapp.New()
	ctx := testApp.BaseApp.NewContext(false, tmproto.Header{})

	assetKeeper := testApp.AssetKeeper
	bankKeeper := testApp.BankKeeper

	addr := sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())

	settings := types.IssueFungibleTokenSettings{
		Issuer:        addr,
		Symbol:        "BTC",
		Description:   "BTC Desc",
		Recipient:     addr,
		InitialAmount: sdk.NewInt(777),
	}

	denom, err := assetKeeper.IssueFungibleToken(ctx, settings)
	requireT.NoError(err)
	requireT.Equal(types.BuildFungibleTokenDenom(settings.Symbol, settings.Issuer), denom)

	gotToken, err := assetKeeper.GetFungibleToken(ctx, denom)
	requireT.NoError(err)
	requireT.Equal(types.FungibleToken{
		Denom:       denom,
		Issuer:      settings.Issuer.String(),
		Symbol:      settings.Symbol,
		Description: settings.Description,
	}, gotToken)

	// check the metadata
	storedMetadata, found := bankKeeper.GetDenomMetaData(ctx, denom)
	requireT.True(found)
	requireT.Equal(banktypes.Metadata{
		Name:        denom,
		Symbol:      settings.Symbol,
		Description: settings.Description,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    denom,
				Exponent: uint32(0),
			},
		},
		Base:    denom,
		Display: denom,
	}, storedMetadata)

	// check the account state
	issuedAssetBalance := bankKeeper.GetBalance(ctx, addr, denom)
	requireT.Equal(sdk.NewCoin(denom, settings.InitialAmount).String(), issuedAssetBalance.String())

	// issue one more time check the double issue validation
	_, err = assetKeeper.IssueFungibleToken(ctx, settings)
	requireT.True(errors.Is(types.ErrInvalidFungibleToken, err))
}
