package app

import (
	"crypto/ed25519"
	"encoding/json"
	"strings"
	"testing"
	"time"

	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/CoreumFoundation/coreum/pkg/staking"
	"github.com/CoreumFoundation/coreum/pkg/types"
	"github.com/CoreumFoundation/coreum/x/auth/ante"
	"github.com/CoreumFoundation/coreum/x/feemodel"
)

func init() {
	n := testNetwork()
	n.SetupPrefixes()
}

var feeConfig = FeeConfig{
	FeeModel: feemodel.Model{
		InitialGasPrice:         sdk.NewInt(2),
		MaxGasPrice:             sdk.NewInt(4),
		MaxDiscount:             sdk.MustNewDecFromStr("0.4"),
		EscalationStartBlockGas: 10,
		MaxBlockGas:             20,
		ShortAverageBlockLength: 3,
		LongAverageBlockLength:  5,
	},
	DeterministicGas: ante.DeterministicGasRequirements{
		BankSend: 10,
	},
}

func testNetwork() Network {
	pubKey, privKey := types.GenerateSecp256k1Key()
	clientCtx := NewDefaultClientContext()
	tx, err := staking.PrepareTxStakingCreateValidator(clientCtx, ed25519.PublicKey(pubKey), privKey, "1000core")
	if err != nil {
		panic(err)
	}
	return NewNetwork(NetworkConfig{
		ChainID:       ChainID("test-network"),
		GenesisTime:   time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
		AddressPrefix: "core",
		TokenSymbol:   TokenSymbolMain,
		Fee:           feeConfig,
		FundedAccounts: []FundedAccount{{
			PublicKey: pubKey,
			Balances:  "1000some-test-token",
		}},
		GenTxs: []json.RawMessage{tx},
	})
}

func TestAddressPrefixIsSet(t *testing.T) {
	assertT := assert.New(t)
	n := testNetwork()
	pubKey, _ := types.GenerateSecp256k1Key()
	secp256k1 := cosmossecp256k1.PubKey{Key: pubKey}
	accountAddress := sdk.AccAddress(secp256k1.Address())
	assertT.True(strings.HasPrefix(accountAddress.String(), n.addressPrefix))
}

func TestGenesisValidation(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	n := testNetwork()

	genesisJSON, err := n.EncodeGenesis()
	requireT.NoError(err)
	gen, err := tmtypes.GenesisDocFromJSON(genesisJSON)
	requireT.NoError(err)
	encCfg := NewEncodingConfig()

	genDocBytes, err := n.EncodeGenesis()
	requireT.NoError(err)

	parsedGenesisDoc, err := tmtypes.GenesisDocFromJSON(genDocBytes)
	requireT.NoError(err)

	assertT.EqualValues(parsedGenesisDoc.ChainID, n.chainID)
	assertT.EqualValues(parsedGenesisDoc.GenesisTime, n.genesisTime)

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
		ModuleBasics.ValidateGenesis(
			encCfg.Marshaler,
			encCfg.TxConfig,
			appStateMapJSONRawMessage,
		))
}

func TestAddFundsToGenesis(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	n := testNetwork()

	pubKey, _ := types.GenerateSecp256k1Key()
	requireT.NoError(n.FundAccount(pubKey, "1000someTestToken"))
	key1 := cosmossecp256k1.PubKey{Key: pubKey}
	accountAddress := sdk.AccAddress(key1.Address())

	pubKey2, _ := types.GenerateSecp256k1Key()
	requireT.NoError(n.FundAccount(pubKey2, "2000someTestToken"))
	key2 := cosmossecp256k1.PubKey{Key: pubKey2}
	accountAddress2 := sdk.AccAddress(key2.Address())

	requireT.Len(n.fundedAccounts, 3)

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
	pubKey, privKey := types.GenerateSecp256k1Key()
	clientCtx := NewDefaultClientContext()
	tx, err := staking.PrepareTxStakingCreateValidator(clientCtx, ed25519.PublicKey(pubKey), privKey, "1000core")
	requireT.NoError(err)
	n.AddGenesisTx(tx)

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
	assert.Equal(t, ante.DeterministicGasRequirements{
		BankSend: 10,
	}, testNetwork().DeterministicGas())
}

func TestNetworkSlicesNotMutable(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	n, err := NetworkByChainID(Devnet)
	requireT.NoError(err)

	pubKey, _ := types.GenerateSecp256k1Key()
	requireT.NoError(n.FundAccount(pubKey, "1000someTestToken"))
	n.AddGenesisTx([]byte("test string"))

	assertT.Len(n.fundedAccounts, 6)
	assertT.Len(n.genTxs, 5)

	n, err = NetworkByChainID(Devnet)
	requireT.NoError(err)
	assertT.Len(n.fundedAccounts, 5)
	assertT.Len(n.genTxs, 4)
}

func TestNetworkConfigNotMutable(t *testing.T) {
	assertT := assert.New(t)

	pubKey, _ := types.GenerateSecp256k1Key()
	cfg := NetworkConfig{
		ChainID:        ChainID("test-network"),
		GenesisTime:    time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
		AddressPrefix:  "core",
		TokenSymbol:    TokenSymbolMain,
		Fee:            feeConfig,
		FundedAccounts: []FundedAccount{{PublicKey: pubKey, Balances: "100test-token"}},
		GenTxs:         []json.RawMessage{[]byte("tx1")},
	}

	n1 := NewNetwork(cfg)

	cfg.Fee.FeeModel.InitialGasPrice.AddRaw(10)
	cfg.Fee.FeeModel.MaxGasPrice.AddRaw(10)
	cfg.FundedAccounts[0] = FundedAccount{PublicKey: pubKey, Balances: "100test-token2"}
	cfg.GenTxs[0] = []byte("tx2")

	assertT.True(n1.FeeModel().InitialGasPrice.Equal(sdk.NewInt(2)))
	assertT.True(n1.FeeModel().MaxGasPrice.Equal(sdk.NewInt(4)))
	assertT.True(n1.FeeModel().MaxDiscount.Equal(sdk.MustNewDecFromStr("0.4")))
	assertT.EqualValues(10, n1.FeeModel().EscalationStartBlockGas)
	assertT.EqualValues(20, n1.FeeModel().MaxBlockGas)
	assertT.EqualValues(3, n1.FeeModel().ShortAverageBlockLength)
	assertT.EqualValues(5, n1.FeeModel().LongAverageBlockLength)
	assertT.EqualValues(n1.fundedAccounts[0], FundedAccount{PublicKey: pubKey, Balances: "100test-token"})
	assertT.EqualValues(n1.genTxs[0], []byte("tx1"))
}

func TestValidateAllGenesis(t *testing.T) {
	assertT := assert.New(t)
	encCfg := NewEncodingConfig()

	for chainID, cfg := range networks {
		n := NewNetwork(cfg)
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
			ModuleBasics.ValidateGenesis(
				encCfg.Marshaler,
				encCfg.TxConfig,
				appStateMapJSONRawMessage,
			), "genesis for network '%s' is invalid", chainID)
	}
}

func TestNetworkFeesNotMutable(t *testing.T) {
	assertT := assert.New(t)

	cfg := NetworkConfig{
		ChainID:       ChainID("test-network"),
		GenesisTime:   time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
		AddressPrefix: "core",
		TokenSymbol:   TokenSymbolMain,
		Fee:           feeConfig,
	}

	n1 := NewNetwork(cfg)

	n1.FeeModel().InitialGasPrice.AddRaw(10)
	n1.FeeModel().MaxGasPrice.AddRaw(10)

	assertT.True(n1.FeeModel().InitialGasPrice.Equal(sdk.NewInt(2)))
	assertT.True(n1.FeeModel().MaxGasPrice.Equal(sdk.NewInt(4)))
}

func TestNetworkConfigConditions(t *testing.T) {
	assertT := assert.New(t)
	for _, cfg := range networks {
		assertT.True(cfg.Fee.FeeModel.InitialGasPrice.Sign() == 1)
		assertT.True(cfg.Fee.FeeModel.MaxGasPrice.Sign() == 1)
		assertT.True(cfg.Fee.FeeModel.MaxGasPrice.GT(cfg.Fee.FeeModel.InitialGasPrice))

		assertT.True(cfg.Fee.FeeModel.MaxDiscount.GT(sdk.ZeroDec()))
		assertT.True(cfg.Fee.FeeModel.MaxDiscount.LT(sdk.OneDec()))

		assertT.Greater(cfg.Fee.FeeModel.EscalationStartBlockGas, int64(0))
		assertT.Greater(cfg.Fee.FeeModel.MaxBlockGas, cfg.Fee.FeeModel.EscalationStartBlockGas)

		assertT.Greater(cfg.Fee.FeeModel.ShortAverageBlockLength, uint(0))
		assertT.Greater(cfg.Fee.FeeModel.LongAverageBlockLength, cfg.Fee.FeeModel.ShortAverageBlockLength)

		assertT.Greater(cfg.Fee.DeterministicGas.BankSend, uint64(0))
	}
}
