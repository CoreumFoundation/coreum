package config_test

import (
	_ "embed"
	"testing"

	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v5/pkg/config"
	"github.com/CoreumFoundation/coreum/v5/pkg/config/constant"
)

func TestNetworkNotMutable(t *testing.T) {
	requireT := require.New(t)
	assertT := assert.New(t)
	pubKey := cosmossecp256k1.GenPrivKey().PubKey()

	// slices not mutable
	n, err := config.NetworkConfigByChainID(constant.ChainIDDev)
	requireT.NoError(err)

	provider := n.Provider.(config.DynamicConfigProvider)
	provider2 := provider.
		WithAccount(sdk.AccAddress(pubKey.Address()), sdk.NewCoins(sdk.NewInt64Coin("someTestToken", 1000)))

	assertT.Len(provider.BankBalances, 1)
	assertT.Len(provider2.BankBalances, 2)

	// re-init the config and check that length remains the same
	n, err = config.NetworkConfigByChainID(constant.ChainIDDev)
	requireT.NoError(err)
	provider = n.Provider.(config.DynamicConfigProvider)

	assertT.Len(provider.BankBalances, 1)
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
