package config_test

import (
	"crypto/sha256"
	_ "embed"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"unsafe"

	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/CoreumFoundation/coreum/v2/pkg/config"
	"github.com/CoreumFoundation/coreum/v2/pkg/config/constant"
)

func TestAddFundsToTheNetwork(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	n, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	unsealConfig()
	n.SetSDKConfig()

	requireT.NoError(err)

	pubKey := cosmossecp256k1.GenPrivKey().PubKey()
	accountAddress := sdk.AccAddress(pubKey.Address())

	pubKey2 := cosmossecp256k1.GenPrivKey().PubKey()
	accountAddress2 := sdk.AccAddress(pubKey2.Address())

	provider := n.Provider.(config.DynamicConfigProvider)
	provider2 := provider.
		WithAccount(accountAddress, sdk.NewCoins(sdk.NewInt64Coin("someTestToken", 1000))).
		WithAccount(accountAddress2, sdk.NewCoins(sdk.NewInt64Coin("someTestToken", 2000)))

	// default 5 + two additional
	requireT.Len(provider2.FundedAccounts, len(provider.FundedAccounts)+2)

	n2 := n
	n2.Provider = provider2

	genDocBytes, err := n2.EncodeGenesis()
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
	requireT.Len(state.Auth.Accounts, len(provider2.FundedAccounts))
	assertT.Subset(state.Auth.Accounts, []account{
		{Address: accountAddress.String()},
		{Address: accountAddress2.String()},
	})
}

func TestNetworkNotMutable(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)
	pubKey := cosmossecp256k1.GenPrivKey().PubKey()

	// slices not mutable
	n, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	requireT.NoError(err)

	provider := n.Provider.(config.DynamicConfigProvider)
	provider2 := provider.
		WithAccount(sdk.AccAddress(pubKey.Address()), sdk.NewCoins(sdk.NewInt64Coin("someTestToken", 1000))).
		WithGenesisTx([]byte("test string"))

	assertT.Len(provider.FundedAccounts, 4)
	assertT.Len(provider.GenTxs, 3)

	assertT.Len(provider2.FundedAccounts, 5)
	assertT.Len(provider2.GenTxs, 4)

	// re-init the config and check that length remains the same
	n, err = config.NetworkConfigByChainID(constant.ChainIDDev)
	requireT.NoError(err)
	provider = n.Provider.(config.DynamicConfigProvider)

	assertT.Len(provider.FundedAccounts, 4)
	assertT.Len(provider.GenTxs, 3)
}

func TestGenesisHash(t *testing.T) {
	tests := []struct {
		name        string
		chainID     constant.ChainID
		genesisHash string
	}{
		{
			chainID:     constant.ChainIDMain,
			genesisHash: "5be3b3e0fee69842c4c73eb5f54eb64684420736473f0f5cef0ba6b81d44f253",
		},
		{
			chainID:     constant.ChainIDTest,
			genesisHash: "276d5df3856ccfba9240687f463c9464e10176e0fc355efb7162e36b09b0e3af",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.chainID), func(t *testing.T) {
			n, err := config.NetworkConfigByChainID(tt.chainID)
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
			n, err := config.NetworkConfigByChainID(tt.chainID)
			require.NoError(t, err)

			unsealConfig()
			n.SetSDKConfig()
			appState, err := n.Provider.AppState()
			require.NoError(t, err)

			bankGenesis, ok := appState[banktypes.ModuleName]
			require.True(t, ok)

			var bankGenesisState banktypes.GenesisState
			err = json.Unmarshal(bankGenesis, &bankGenesisState)
			require.NoError(t, err)
			require.Equal(t, tt.wantSupply.Amount.String(), bankGenesisState.Supply.AmountOf(tt.wantSupply.Denom).String())
		})
	}
}

func TestStaticConfigProviders(t *testing.T) {
	tests := []struct {
		name          string
		chainID       constant.ChainID
		denom         string
		addressPrefix string
	}{
		{
			name:          "devnetnet",
			chainID:       constant.ChainIDDev,
			denom:         constant.DenomDev,
			addressPrefix: constant.AddressPrefixDev,
		},
		{
			name:          "testnet",
			chainID:       constant.ChainIDTest,
			denom:         constant.DenomTest,
			addressPrefix: constant.AddressPrefixTest,
		},
		{
			name:          "mainnet",
			chainID:       constant.ChainIDMain,
			denom:         constant.DenomMain,
			addressPrefix: constant.AddressPrefixMain,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			n, err := config.NetworkConfigByChainID(tt.chainID)
			require.NoError(t, err)

			assert.Equal(t, tt.chainID, n.ChainID())
			assert.Equal(t, tt.denom, n.Denom())

			assert.Equal(t, tt.chainID, n.Provider.GetChainID())
			assert.Equal(t, tt.denom, n.Provider.GetDenom())
			assert.Equal(t, tt.addressPrefix, n.Provider.GetAddressPrefix())
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
