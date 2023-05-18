package keeper_test

import (
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	"github.com/CoreumFoundation/coreum/x/nft"
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
	nft.RegisterQueryServer(queryHelper, app.NFTKeeper)
	queryClient := nft.NewQueryClient(queryHelper)

	s.app = app
	s.ctx = ctx
	s.queryClient = queryClient
	s.addrs = simapp.AddTestAddrsIncremental(app, ctx, 3, sdk.NewInt(30000000))
	s.testClassID = fmt.Sprintf("%s-%s", "kitty", sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address()))
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestSaveClass() {
	except := nft.Class{
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
	s.Require().EqualValues([]*nft.Class{&except}, classes)
}

func (s *TestSuite) TestUpdateClass() {
	class := nft.Class{
		Id:          s.testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.app.NFTKeeper.SaveClass(s.ctx, class)
	s.Require().NoError(err)

	noExistClass := nft.Class{
		Id:          fmt.Sprintf("%s-%s", "kitty1", sdk.AccAddress(ed25519.GenPrivKey().PubKey().Address())),
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}

	err = s.app.NFTKeeper.UpdateClass(s.ctx, noExistClass)
	s.Require().Error(err)
	s.Require().Contains(err.Error(), "nft class does not exist")

	except := nft.Class{
		Id:          s.testClassID,
		Name:        "My crypto Kitty",
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}

	err = s.app.NFTKeeper.UpdateClass(s.ctx, except)
	s.Require().NoError(err)

	actual, has := s.app.NFTKeeper.GetClass(s.ctx, s.testClassID)
	s.Require().True(has)
	s.Require().EqualValues(except, actual)
}

func (s *TestSuite) TestMint() {
	class := nft.Class{
		Id:          s.testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.app.NFTKeeper.SaveClass(s.ctx, class)
	s.Require().NoError(err)

	expNFT := nft.NFT{
		ClassId: s.testClassID,
		Id:      testID,
		Uri:     testURI,
	}
	err = s.app.NFTKeeper.Mint(s.ctx, expNFT, s.addrs[0])
	s.Require().NoError(err)

	// test GetNFT
	actNFT, has := s.app.NFTKeeper.GetNFT(s.ctx, s.testClassID, testID)
	s.Require().True(has)
	s.Require().EqualValues(expNFT, actNFT)

	// test GetOwner
	owner := s.app.NFTKeeper.GetOwner(s.ctx, s.testClassID, testID)
	s.Require().True(s.addrs[0].Equals(owner))

	// test GetNFTsOfClass
	actNFTs := s.app.NFTKeeper.GetNFTsOfClass(s.ctx, s.testClassID)
	s.Require().EqualValues([]nft.NFT{expNFT}, actNFTs)

	// test GetNFTsOfClassByOwner
	actNFTs = s.app.NFTKeeper.GetNFTsOfClassByOwner(s.ctx, s.testClassID, s.addrs[0])
	s.Require().EqualValues([]nft.NFT{expNFT}, actNFTs)

	// test GetBalance
	balance := s.app.NFTKeeper.GetBalance(s.ctx, s.testClassID, s.addrs[0])
	s.Require().EqualValues(uint64(1), balance)

	// test GetTotalSupply
	supply := s.app.NFTKeeper.GetTotalSupply(s.ctx, s.testClassID)
	s.Require().EqualValues(uint64(1), supply)

	expNFT2 := nft.NFT{
		ClassId: s.testClassID,
		Id:      testID + "2",
		Uri:     testURI + "2",
	}
	err = s.app.NFTKeeper.Mint(s.ctx, expNFT2, s.addrs[0])
	s.Require().NoError(err)

	// test GetNFTsOfClassByOwner
	actNFTs = s.app.NFTKeeper.GetNFTsOfClassByOwner(s.ctx, s.testClassID, s.addrs[0])
	s.Require().EqualValues([]nft.NFT{expNFT, expNFT2}, actNFTs)

	// test GetBalance
	balance = s.app.NFTKeeper.GetBalance(s.ctx, s.testClassID, s.addrs[0])
	s.Require().EqualValues(uint64(2), balance)
}

func (s *TestSuite) TestBurn() {
	except := nft.Class{
		Id:          s.testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.app.NFTKeeper.SaveClass(s.ctx, except)
	s.Require().NoError(err)

	expNFT := nft.NFT{
		ClassId: s.testClassID,
		Id:      testID,
		Uri:     testURI,
	}
	err = s.app.NFTKeeper.Mint(s.ctx, expNFT, s.addrs[0])
	s.Require().NoError(err)

	err = s.app.NFTKeeper.Burn(s.ctx, s.testClassID, testID)
	s.Require().NoError(err)

	// test GetNFT
	_, has := s.app.NFTKeeper.GetNFT(s.ctx, s.testClassID, testID)
	s.Require().False(has)

	// test GetOwner
	owner := s.app.NFTKeeper.GetOwner(s.ctx, s.testClassID, testID)
	s.Require().Nil(owner)

	// test GetNFTsOfClass
	actNFTs := s.app.NFTKeeper.GetNFTsOfClass(s.ctx, s.testClassID)
	s.Require().Empty(actNFTs)

	// test GetNFTsOfClassByOwner
	actNFTs = s.app.NFTKeeper.GetNFTsOfClassByOwner(s.ctx, s.testClassID, s.addrs[0])
	s.Require().Empty(actNFTs)

	// test GetBalance
	balance := s.app.NFTKeeper.GetBalance(s.ctx, s.testClassID, s.addrs[0])
	s.Require().EqualValues(uint64(0), balance)

	// test GetTotalSupply
	supply := s.app.NFTKeeper.GetTotalSupply(s.ctx, s.testClassID)
	s.Require().EqualValues(uint64(0), supply)
}

func (s *TestSuite) TestUpdate() {
	class := nft.Class{
		Id:          s.testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.app.NFTKeeper.SaveClass(s.ctx, class)
	s.Require().NoError(err)

	myNFT := nft.NFT{
		ClassId: s.testClassID,
		Id:      testID,
		Uri:     testURI,
	}
	err = s.app.NFTKeeper.Mint(s.ctx, myNFT, s.addrs[0])
	s.Require().NoError(err)

	expNFT := nft.NFT{
		ClassId: s.testClassID,
		Id:      testID,
		Uri:     "updated",
	}

	err = s.app.NFTKeeper.Update(s.ctx, expNFT)
	s.Require().NoError(err)

	// test GetNFT
	actNFT, has := s.app.NFTKeeper.GetNFT(s.ctx, s.testClassID, testID)
	s.Require().True(has)
	s.Require().EqualValues(expNFT, actNFT)
}

func (s *TestSuite) TestTransfer() {
	class := nft.Class{
		Id:          s.testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.app.NFTKeeper.SaveClass(s.ctx, class)
	s.Require().NoError(err)

	expNFT := nft.NFT{
		ClassId: s.testClassID,
		Id:      testID,
		Uri:     testURI,
	}
	err = s.app.NFTKeeper.Mint(s.ctx, expNFT, s.addrs[0])
	s.Require().NoError(err)

	// valid owner
	err = s.app.NFTKeeper.Transfer(s.ctx, s.testClassID, testID, s.addrs[1])
	s.Require().NoError(err)

	// test GetOwner
	owner := s.app.NFTKeeper.GetOwner(s.ctx, s.testClassID, testID)
	s.Require().Equal(s.addrs[1], owner)

	balanceAddr0 := s.app.NFTKeeper.GetBalance(s.ctx, s.testClassID, s.addrs[0])
	s.Require().EqualValues(uint64(0), balanceAddr0)

	balanceAddr1 := s.app.NFTKeeper.GetBalance(s.ctx, s.testClassID, s.addrs[1])
	s.Require().EqualValues(uint64(1), balanceAddr1)

	// test GetNFTsOfClassByOwner
	actNFTs := s.app.NFTKeeper.GetNFTsOfClassByOwner(s.ctx, s.testClassID, s.addrs[1])
	s.Require().EqualValues([]nft.NFT{expNFT}, actNFTs)
}

func (s *TestSuite) TestExportGenesis() {
	class := nft.Class{
		Id:          s.testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	err := s.app.NFTKeeper.SaveClass(s.ctx, class)
	s.Require().NoError(err)

	expNFT := nft.NFT{
		ClassId: s.testClassID,
		Id:      testID,
		Uri:     testURI,
	}
	err = s.app.NFTKeeper.Mint(s.ctx, expNFT, s.addrs[0])
	s.Require().NoError(err)

	expGenesis := &nft.GenesisState{
		Classes: []*nft.Class{&class},
		Entries: []*nft.Entry{{
			Owner: s.addrs[0].String(),
			Nfts:  []*nft.NFT{&expNFT},
		}},
	}
	genesis := s.app.NFTKeeper.ExportGenesis(s.ctx)
	s.Require().Equal(expGenesis, genesis)
}

func (s *TestSuite) TestInitGenesis() {
	expClass := nft.Class{
		Id:          s.testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
		UriHash:     testClassURIHash,
	}
	expNFT := nft.NFT{
		ClassId: s.testClassID,
		Id:      testID,
		Uri:     testURI,
	}
	expGenesis := &nft.GenesisState{
		Classes: []*nft.Class{&expClass},
		Entries: []*nft.Entry{{
			Owner: s.addrs[0].String(),
			Nfts:  []*nft.NFT{&expNFT},
		}},
	}
	s.app.NFTKeeper.InitGenesis(s.ctx, expGenesis)

	actual, has := s.app.NFTKeeper.GetClass(s.ctx, s.testClassID)
	s.Require().True(has)
	s.Require().EqualValues(expClass, actual)

	// test GetNFT
	actNFT, has := s.app.NFTKeeper.GetNFT(s.ctx, s.testClassID, testID)
	s.Require().True(has)
	s.Require().EqualValues(expNFT, actNFT)
}
