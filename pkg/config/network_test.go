package config_test

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"
	"unsafe"

	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	feemodeltypes "github.com/CoreumFoundation/coreum/x/feemodel/types"
)

func TestAddFundsToGenesis(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	n, err := config.NetworkByChainID(constant.ChainIDDev)
	unsealConfig()
	n.SetSDKConfig()

	requireT.NoError(err)

	pubKey := cosmossecp256k1.GenPrivKey().PubKey()
	accountAddress := sdk.AccAddress(pubKey.Address())

	initiallyFundedAccounts := len(n.FundedAccounts())

	requireT.NoError(n.FundAccount(accountAddress, sdk.NewCoins(sdk.NewInt64Coin("someTestToken", 1000))))

	pubKey2 := cosmossecp256k1.GenPrivKey().PubKey()
	accountAddress2 := sdk.AccAddress(pubKey2.Address())
	requireT.NoError(n.FundAccount(accountAddress2, sdk.NewCoins(sdk.NewInt64Coin("someTestToken", 2000))))

	// default 5 + two additional
	requireT.Len(n.FundedAccounts(), initiallyFundedAccounts+2)

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
	requireT.Len(state.Auth.Accounts, len(n.FundedAccounts()))
	assertT.Subset(state.Auth.Accounts, []account{
		{Address: accountAddress.String()},
		{Address: accountAddress2.String()},
	})
}

func TestConfigNotMutable(t *testing.T) {
	requireT := require.New(t)
	pubKey := cosmossecp256k1.GenPrivKey().PubKey()
	cfg := config.NetworkConfig{
		ChainID:        "test-network",
		GenesisTime:    time.Date(2022, 6, 27, 12, 0, 0, 0, time.UTC),
		AddressPrefix:  "core",
		Denom:          "ucore",
		FundedAccounts: []config.FundedAccount{{Address: sdk.AccAddress(pubKey.Address()).String(), Balances: sdk.NewCoins(sdk.NewInt64Coin("test-token", 100))}},
		GenTxs:         []json.RawMessage{[]byte("tx1")},
		Fee: config.FeeConfig{
			FeeModel: feemodeltypes.NewModel(feemodeltypes.ModelParams{
				InitialGasPrice:       sdk.NewDec(2),
				MaxGasPriceMultiplier: sdk.NewDec(2),
			}),
		},
	}

	n := config.NewNetwork(cfg)

	// update fee settings
	params := cfg.Fee.FeeModel.Params()
	params.InitialGasPrice.Add(sdk.NewDec(10))
	params.MaxGasPriceMultiplier.Add(sdk.NewDec(10))
	// update the account
	cfg.FundedAccounts[0] = config.FundedAccount{Address: sdk.AccAddress(pubKey.Address()).String(), Balances: sdk.NewCoins(sdk.NewInt64Coin("test-token2", 100))}
	// update the gen tx
	cfg.GenTxs[0] = []byte("tx2")

	nParams := n.FeeModel().Params()
	// assert fee settings
	requireT.True(nParams.InitialGasPrice.Equal(sdk.NewDec(2)))
	requireT.True(nParams.MaxGasPriceMultiplier.Equal(sdk.NewDec(2)))
	// assert account
	requireT.EqualValues(n.FundedAccounts()[0], config.FundedAccount{Address: sdk.AccAddress(pubKey.Address()).String(), Balances: sdk.NewCoins(sdk.NewInt64Coin("test-token", 100))})
	// assert gen tx
	requireT.EqualValues(n.GenTxs()[0], []byte("tx1"))
}

func TestChainNotMutable(t *testing.T) {
	requireT := require.New(t)
	pubKey := cosmossecp256k1.GenPrivKey().PubKey()

	// slices not mutable
	n, err := config.NetworkByChainID(constant.ChainIDDev)
	requireT.NoError(err)
	requireT.NoError(n.FundAccount(sdk.AccAddress(pubKey.Address()), sdk.NewCoins(sdk.NewInt64Coin("someTestToken", 1000))))
	n.AddGenesisTx([]byte("test string"))

	requireT.Len(n.FundedAccounts(), 5)
	requireT.Len(n.GenTxs(), 4)

	// re-init the config and check that length remains the same
	n, err = config.NetworkByChainID(constant.ChainIDDev)
	requireT.NoError(err)
	requireT.Len(n.FundedAccounts(), 4)
	requireT.Len(n.GenTxs(), 3)
}

func TestGenesisHash(t *testing.T) {
	tests := []struct {
		name        string
		chainID     constant.ChainID
		genesisHash string
	}{
		{
			chainID:     constant.ChainIDMain,
			genesisHash: "b7a9fa3445d6233372e72534c37e947d939e32a18f12928b23d407fc2b8ecc4d",
		},
		{
			chainID:     constant.ChainIDTest,
			genesisHash: "8ece5edef851738ef3d8435a58bfbfe91b163ff199785ae1470da84466f0f1c1",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.chainID), func(t *testing.T) {
			n, err := config.NetworkByChainID(tt.chainID)
			require.NoError(t, err)

			unsealConfig()
			n.SetSDKConfig()
			genesisDoc, err := n.EncodeGenesis()
			require.NoError(t, err)

			require.NoError(t, err)
			require.Equal(t, tt.genesisHash, fmt.Sprintf("%x", sha256.Sum256(genesisDoc)))
		})
	}
}

func TestGenesisCoreTotalSupply(t *testing.T) {
	tests := []struct {
		name       string
		chainID    constant.ChainID
		wantSupply sdk.Coin
	}{
		{
			name:       "testnet",
			chainID:    constant.ChainIDTest,
			wantSupply: sdk.NewCoin(constant.DenomTest, sdk.NewInt(500_000_000_000_000)),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n, err := config.NetworkByChainID(tt.chainID)
			require.NoError(t, err)

			unsealConfig()
			n.SetSDKConfig()
			genesisDoc, err := n.GenesisDoc()
			require.NoError(t, err)

			var appStateMapJSONRawMessage map[string]json.RawMessage
			err = json.Unmarshal(genesisDoc.AppState, &appStateMapJSONRawMessage)
			require.NoError(t, err)

			bankGenesis, ok := appStateMapJSONRawMessage[banktypes.ModuleName]
			require.True(t, ok)

			var bankGenesisState banktypes.GenesisState
			err = json.Unmarshal(bankGenesis, &bankGenesisState)
			require.NoError(t, err)
			require.Equal(t, tt.wantSupply.Amount.String(), bankGenesisState.Supply.AmountOf(tt.wantSupply.Denom).String())
		})
	}
}

func unsealConfig() {
	sdkConfig := sdk.GetConfig()
	unsafeSetField(sdkConfig, "sealed", false)
	unsafeSetField(sdkConfig, "sealedch", make(chan struct{}))
}

func unsafeSetField(object interface{}, fieldName string, value interface{}) {
	rs := reflect.ValueOf(object).Elem()
	field := rs.FieldByName(fieldName)
	// rf can't be read or set.
	reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).
		Elem().
		Set(reflect.ValueOf(value))
}
