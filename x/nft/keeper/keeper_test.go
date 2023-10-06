package keeper_test

import (
	"fmt"
	"testing"

	sdkmath "cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmosnft "github.com/cosmos/cosmos-sdk/x/nft"
	"github.com/stretchr/testify/suite"

	"github.com/CoreumFoundation/coreum/v3/testutil/simapp"
	"github.com/CoreumFoundation/coreum/v3/x/nft"
	nftkeeper "github.com/CoreumFoundation/coreum/v3/x/nft/keeper"
)

const (
	testClassName        = "Crypto Kitty"
	testClassSymbol      = "kitty"
	testClassDescription = "Crypto Kitty"
	testClassURI         = "class uri"
	testClassURIHash     = "ae702cefd6b6a65fe2f991ad6d9969ed"
	testID               = "kitty1"
	testURI              = "kitty uri"
)

type TestSuite struct {
	suite.Suite

	testClassID string
	app         *simapp.App
	ctx         sdk.Context
	addrs       []sdk.AccAddress
	queryClient nft.QueryClient
}

func (s *TestSuite) SetupTest() {
	app := simapp.New()
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: tmtime.Now()})
	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	keeper := nftkeeper.NewKeeper(app.NFTKeeper)
	nft.RegisterQueryServer(queryHelper, keeper)
	queryClient := nft.NewQueryClient(queryHelper)

	s.app = app
	s.ctx = ctx
	s.queryClient = queryClient
	s.addrs = simapp.AddTestAddrsIncremental(app, ctx, 3, sdkmath.NewInt(30000000))
	s.testClassID = fmt.Sprintf("%s-%s", "kitty", sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address()))
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestSaveClass() {
	except := cosmosnft.Class{
		Id:          s.testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.app.NFTKeeper.SaveClass(s.ctx, except)
	s.Require().NoError(err)

	actual, has := s.app.NFTKeeper.GetClass(s.ctx, s.testClassID)
	s.Require().True(has)
	s.Require().EqualValues(except, actual)

	classes := s.app.NFTKeeper.GetClasses(s.ctx)
	s.Require().EqualValues([]*cosmosnft.Class{&except}, classes)
}

func (s *TestSuite) TestMint() {
	class := cosmosnft.Class{
		Id:          s.testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.app.NFTKeeper.SaveClass(s.ctx, class)
	s.Require().NoError(err)

	expNFT := cosmosnft.NFT{
		ClassId: s.testClassID,
		Id:      testID,
		Uri:     testURI,
	}
	err = s.app.NFTKeeper.Mint(s.ctx, expNFT, s.addrs[0])
	s.Require().NoError(err)

	expNFT2 := cosmosnft.NFT{
		ClassId: s.testClassID,
		Id:      testID + "2",
		Uri:     testURI + "2",
	}
	err = s.app.NFTKeeper.Mint(s.ctx, expNFT2, s.addrs[0])
	s.Require().NoError(err)
}
