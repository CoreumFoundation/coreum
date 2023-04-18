package testutil

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/testutil/network"
	"github.com/CoreumFoundation/coreum/x/nft"
)

const (
	testClassID          = "kitty"
	testClassName        = "Crypto Kitty"
	testClassSymbol      = "kitty"
	testClassDescription = "Crypto Kitty"
	testClassURI         = "class uri"
	testID               = "kitty1"
	testURI              = "kitty uri"
)

var (
	ExpClass = nft.Class{ //nolint:revive // test constants
		Id:          testClassID,
		Name:        testClassName,
		Symbol:      testClassSymbol,
		Description: testClassDescription,
		Uri:         testClassURI,
	}

	ExpNFT = nft.NFT{
		ClassId: testClassID,
		Id:      testID,
		Uri:     testURI,
	}
)

type IntegrationTestSuite struct { //nolint:revive // test helper
	suite.Suite

	cfg     network.Config
	network *network.Network
	owner   string
}

func NewIntegrationTestSuite() *IntegrationTestSuite { //nolint:revive // test helper
	return &IntegrationTestSuite{
		cfg: network.DefaultConfig(),
	}
}

func (s *IntegrationTestSuite) SetupSuite() { //nolint:revive // test helper
	s.T().Log("setting up integration test suite")

	// gen account to use as nft owner
	keyInfo, mnemonic := genAccount(s)
	address, err := keyInfo.GetAddress()
	s.Require().NoError(err)
	s.T().Logf("Created new account address:%s", address)

	// fund account to pay for the transactions
	cfg, err := network.ApplyConfigOptions(s.cfg, network.WithChainDenomFundedAccounts(
		[]network.FundedAccount{
			{
				Address: address,
				Amount:  sdkmath.NewInt(10_000_000),
			},
		}))
	s.Require().NoError(err)
	s.cfg = cfg
	s.owner = address.String()

	// set owner in the genesis
	genesisState := s.cfg.GenesisState
	nftGenesis := nft.GenesisState{
		Classes: []*nft.Class{&ExpClass},
		Entries: []*nft.Entry{{
			Owner: s.owner,
			Nfts:  []*nft.NFT{&ExpNFT},
		}},
	}
	nftDataBz, err := s.cfg.Codec.MarshalJSON(&nftGenesis)
	s.Require().NoError(err)
	genesisState[nft.ModuleName] = nftDataBz
	s.cfg.GenesisState = genesisState

	// start simapp network
	s.network = network.New(s.T(), s.cfg)
	s.Require().NoError(err)

	// import key
	s.owner = address.String()
	s.importMnemonic(s.owner, mnemonic, s.network.Validators[0].ClientCtx)

	_, err = s.network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() { //nolint:revive // test helper
	s.T().Log("tearing down integration test suite")
}

func (s *IntegrationTestSuite) TestCLITxSend() { //nolint:revive // test
	val := s.network.Validators[0]
	args := []string{
		fmt.Sprintf("--%s=%s", flags.FlagFrom, s.owner),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=%s", flags.FlagFees, sdk.NewCoins(sdk.NewCoin(s.cfg.BondDenom, sdkmath.NewInt(10000))).String()),
	}
	testCases := []struct {
		name         string
		args         []string
		expectedCode uint32
		expectErr    bool
	}{
		{
			"valid transaction",
			[]string{
				testClassID,
				testID,
				val.Address.String(),
			},
			0,
			false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.Run(tc.name, func() {
			clientCtx := val.ClientCtx
			args = append(args, tc.args...)
			out, err := ExecSend(
				val,
				args,
			)
			if tc.expectErr {
				s.Require().Error(err)
			} else {
				var txResp sdk.TxResponse
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &txResp), out.String())
				s.Require().Equal(tc.expectedCode, txResp.Code, out.String())
			}
			err = s.network.WaitForNextBlock()
			require.NoError(s.T(), err)
		})
	}
}

func genAccount(s *IntegrationTestSuite) (*keyring.Record, string) {
	// Generate and store a new mnemonic using temporary keyring
	keyRecord, mnemonic, err := keyring.NewInMemory(config.NewEncodingConfig(app.ModuleBasics).Codec).NewMnemonic(
		"tmp",
		keyring.English,
		sdk.GetConfig().GetFullBIP44Path(),
		"",
		hd.Secp256k1,
	)
	s.Require().NoError(err)

	return keyRecord, mnemonic
}

func (s *IntegrationTestSuite) importMnemonic(name, mnemonic string, clientCtx client.Context) {
	_, err := clientCtx.Keyring.NewAccount(
		name,
		mnemonic,
		"",
		sdk.GetConfig().GetFullBIP44Path(),
		hd.Secp256k1,
	)
	s.Require().NoError(err)
}
