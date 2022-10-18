package config_test

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"strings"
	"testing"
	"time"

	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/staking"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	feemodeltypes "github.com/CoreumFoundation/coreum/x/feemodel/types"
)

func init() {
	n := testNetwork()
	n.SetSDKConfig()
}

var feeConfig = config.FeeConfig{
	FeeModel: feemodeltypes.NewModel(feemodeltypes.ModelParams{
		InitialGasPrice:         sdk.NewDec(2),
		MaxGasPriceMultiplier:   sdk.NewDec(2),
		MaxDiscount:             sdk.MustNewDecFromStr("0.4"),
		EscalationStartFraction: sdk.MustNewDecFromStr("0.8"),
		MaxBlockGas:             20,
		ShortEmaBlockLength:     3,
		LongEmaBlockLength:      5,
	}),
	DeterministicGas: config.DeterministicGasRequirements{
		BankSend: 10,
	},
}

func testNetwork() config.Network {
	validatorPubKey, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	stakerPrivKey := *cosmossecp256k1.GenPrivKey()
	stakerPubKey := stakerPrivKey.PubKey()

	clientCtx := tx.NewClientContext(app.ModuleBasics)
	createValidatorTx, err := staking.PrepareTxStakingCreateValidator(clientCtx, validatorPubKey, stakerPrivKey, "1000core")
	if err != nil {
		panic(err)
	}
	return config.NewNetwork(config.NetworkConfig{
		ChainID:       "test-network",
		GenesisTime:   time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
		AddressPrefix: "core",
		Denom:         config.DenomDev,
		BaseDenom:     config.BaseDenomMain,
		Fee:           feeConfig,
		FundedAccounts: []config.FundedAccount{{
			PublicKey: stakerPubKey,
			Balances:  "1000some-test-token",
		}},
		GenTxs: []json.RawMessage{createValidatorTx},
		GovConfig: config.GovConfig{
			ProposalConfig: config.GovProposalConfig{
				MinDepositAmount: "10000000",
				MinDepositPeriod: "172800s",
				VotingPeriod:     "172800s",
			},
		},
		StakingConfig: config.StakingConfig{
			UnbondingTime: "1814400s",
			MaxValidators: 32,
		},
	})
}

func TestAddressPrefixIsSet(t *testing.T) {
	assertT := assert.New(t)
	n := testNetwork()
	address := sdk.AccAddress(cosmossecp256k1.GenPrivKey().PubKey().Address())
	assertT.True(strings.HasPrefix(address.String(), n.AddressPrefix()))
}

func TestGenesisValidation(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	n := testNetwork()

	genesisJSON, err := n.EncodeGenesis()
	requireT.NoError(err)
	gen, err := tmtypes.GenesisDocFromJSON(genesisJSON)
	requireT.NoError(err)
	encCfg := config.NewEncodingConfig(app.ModuleBasics)

	genDocBytes, err := n.EncodeGenesis()
	requireT.NoError(err)

	parsedGenesisDoc, err := tmtypes.GenesisDocFromJSON(genDocBytes)
	requireT.NoError(err)

	assertT.EqualValues(parsedGenesisDoc.ChainID, n.ChainID())
	assertT.EqualValues(parsedGenesisDoc.GenesisTime, n.GenesisTime())

	// In order to compare app state, we need to unmarshal it first
	// because comparing json.RawMessage may give false negatives.
	appStateMap := map[string]interface{}{}
	err = json.Unmarshal(gen.AppState, &appStateMap)
	requireT.NoError(err)
	parsedAppStateMap := map[string]interface{}{}
	err = json.Unmarshal(parsedGenesisDoc.AppState, &parsedAppStateMap)
	requireT.NoError(err)
	assertT.EqualValues(appStateMap, parsedAppStateMap)

	var appStateMapJSONRawMessage map[string]json.RawMessage
	err = json.Unmarshal(gen.AppState, &appStateMapJSONRawMessage)
	requireT.NoError(err)
	requireT.NoError(
		app.ModuleBasics.ValidateGenesis(
			encCfg.Codec,
			encCfg.TxConfig,
			appStateMapJSONRawMessage,
		))
}

func TestAddFundsToGenesis(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	n := testNetwork()

	pubKey := cosmossecp256k1.GenPrivKey().PubKey()
	requireT.NoError(n.FundAccount(pubKey, "1000someTestToken"))
	accountAddress := sdk.AccAddress(pubKey.Address())

	pubKey2 := cosmossecp256k1.GenPrivKey().PubKey()
	requireT.NoError(n.FundAccount(pubKey2, "2000someTestToken"))
	accountAddress2 := sdk.AccAddress(pubKey2.Address())

	requireT.Len(n.FundedAccounts(), 3)

	genDocBytes, err := n.EncodeGenesis()
	requireT.NoError(err)

	parsedGenesisDoc, err := tmtypes.GenesisDocFromJSON(genDocBytes)
	requireT.NoError(err)

	type coin struct {
		Denom  string `json:"denom"`
		Amount string `json:"amount"`
	}
	type balance struct {
		Address string `json:"address"`
		Coins   []coin `json:"coins"`
	}
	type account struct {
		Address string `json:"address"`
	}
	var state struct {
		Bank struct {
			Balances []balance `json:"balances"`
			Supply   []coin    `json:"supply"`
		} `json:"bank"`
		Auth struct {
			Accounts []account `json:"accounts"`
		} `json:"auth"`
	}

	err = json.Unmarshal(parsedGenesisDoc.AppState, &state)
	requireT.NoError(err)

	assertT.Subset(state.Bank.Balances, []balance{
		{
			Address: accountAddress.String(),
			Coins: []coin{
				{Denom: "someTestToken", Amount: "1000"},
			},
		},
		{
			Address: accountAddress2.String(),
			Coins: []coin{
				{Denom: "someTestToken", Amount: "2000"},
			},
		},
	})

	assertT.Contains(
		state.Bank.Supply,
		coin{Denom: "someTestToken", Amount: "3000"},
	)
	requireT.Len(state.Auth.Accounts, 3)
	assertT.Subset(state.Auth.Accounts, []account{
		{Address: accountAddress.String()},
		{Address: accountAddress2.String()},
	})
}

func TestAddGenTx(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	n := testNetwork()
	validatorPubKey, _, err := ed25519.GenerateKey(rand.Reader)
	requireT.NoError(err)

	stakerPrivKey := *cosmossecp256k1.GenPrivKey()
	clientCtx := tx.NewClientContext(app.ModuleBasics)
	createValidatorTx, err := staking.PrepareTxStakingCreateValidator(clientCtx, validatorPubKey, stakerPrivKey, "1000core")
	requireT.NoError(err)
	n.AddGenesisTx(createValidatorTx)

	genDocBytes, err := n.EncodeGenesis()
	requireT.NoError(err)

	parsedGenesisDoc, err := tmtypes.GenesisDocFromJSON(genDocBytes)
	requireT.NoError(err)

	var state struct {
		GenUtil struct {
			GenTxs []json.RawMessage `json:"gen_txs"` //nolint:tagliatelle
		} `json:"genutil"`
	}

	err = json.Unmarshal(parsedGenesisDoc.AppState, &state)
	requireT.NoError(err)
	assertT.Len(state.GenUtil.GenTxs, 2)
}

func TestDeterministicGas(t *testing.T) {
	assert.Equal(t, config.DeterministicGasRequirements{
		BankSend: 10,
	}, testNetwork().DeterministicGas())
}

func TestNetworkSlicesNotMutable(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	n, err := config.NetworkByChainID(config.ChainIDDev)
	requireT.NoError(err)

	pubKey := cosmossecp256k1.GenPrivKey().PubKey()
	requireT.NoError(n.FundAccount(pubKey, "1000someTestToken"))
	n.AddGenesisTx([]byte("test string"))

	assertT.Len(n.FundedAccounts(), 6)
	assertT.Len(n.GenTxs(), 5)

	n, err = config.NetworkByChainID(config.ChainIDDev)
	requireT.NoError(err)
	assertT.Len(n.FundedAccounts(), 5)
	assertT.Len(n.GenTxs(), 4)
}

func TestNetworkConfigNotMutable(t *testing.T) {
	assertT := assert.New(t)

	pubKey := cosmossecp256k1.GenPrivKey().PubKey()
	cfg := config.NetworkConfig{
		ChainID:        "test-network",
		GenesisTime:    time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
		AddressPrefix:  "core",
		BaseDenom:      config.BaseDenomMain,
		Fee:            feeConfig,
		FundedAccounts: []config.FundedAccount{{PublicKey: pubKey, Balances: "100test-token"}},
		GenTxs:         []json.RawMessage{[]byte("tx1")},
	}

	n1 := config.NewNetwork(cfg)

	params := cfg.Fee.FeeModel.Params()
	params.InitialGasPrice.Add(sdk.NewDec(10))
	params.MaxGasPriceMultiplier.Add(sdk.NewDec(10))
	cfg.FundedAccounts[0] = config.FundedAccount{PublicKey: pubKey, Balances: "100test-token2"}
	cfg.GenTxs[0] = []byte("tx2")

	nParams := n1.FeeModel().Params()
	assertT.True(nParams.InitialGasPrice.Equal(sdk.NewDec(2)))
	assertT.True(nParams.MaxGasPriceMultiplier.Equal(sdk.NewDec(2)))
	assertT.True(nParams.MaxDiscount.Equal(sdk.MustNewDecFromStr("0.4")))
	assertT.True(nParams.EscalationStartFraction.Equal(sdk.MustNewDecFromStr("0.8")))
	assertT.EqualValues(20, nParams.MaxBlockGas)
	assertT.EqualValues(3, nParams.ShortEmaBlockLength)
	assertT.EqualValues(5, nParams.LongEmaBlockLength)
	assertT.EqualValues(n1.FundedAccounts()[0], config.FundedAccount{PublicKey: pubKey, Balances: "100test-token"})
	assertT.EqualValues(n1.GenTxs()[0], []byte("tx1"))
}

func TestNetworkFeesNotMutable(t *testing.T) {
	assertT := assert.New(t)

	cfg := config.NetworkConfig{
		ChainID:       "test-network",
		GenesisTime:   time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
		AddressPrefix: "core",
		BaseDenom:     config.BaseDenomMain,
		Fee:           feeConfig,
	}

	n1 := config.NewNetwork(cfg)

	nParams := n1.FeeModel().Params()
	nParams.InitialGasPrice.Add(sdk.NewDec(10))
	nParams.MaxGasPriceMultiplier.Add(sdk.NewDec(10))

	assertT.True(nParams.InitialGasPrice.Equal(sdk.NewDec(2)))
	assertT.True(nParams.MaxGasPriceMultiplier.Equal(sdk.NewDec(2)))
}

func TestValidateAllGenesis(t *testing.T) {
	assertT := assert.New(t)
	encCfg := config.NewEncodingConfig(app.ModuleBasics)

	for _, n := range config.EnabledNetworks() {
		genesisJSON, err := n.EncodeGenesis()
		if !assertT.NoError(err) {
			continue
		}

		gen, err := tmtypes.GenesisDocFromJSON(genesisJSON)
		if !assertT.NoError(err) {
			continue
		}

		var appStateMapJSONRawMessage map[string]json.RawMessage
		err = json.Unmarshal(gen.AppState, &appStateMapJSONRawMessage)
		if !assertT.NoError(err) {
			continue
		}

		assertT.NoErrorf(
			app.ModuleBasics.ValidateGenesis(
				encCfg.Codec,
				encCfg.TxConfig,
				appStateMapJSONRawMessage,
			), "genesis for network '%s' is invalid", n.ChainID())
	}
}

func TestNetworkConfigConditions(t *testing.T) {
	assertT := assert.New(t)
	for _, n := range config.EnabledNetworks() {
		assert.NoError(t, n.FeeModel().Params().ValidateBasic())
		assertT.Greater(n.DeterministicGas().BankSend, uint64(0))
	}
}
