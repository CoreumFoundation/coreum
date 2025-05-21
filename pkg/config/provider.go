package config

import (
	"context"
	"encoding/json"

	cmtjson "github.com/cometbft/cometbft/libs/json"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/btcutil/bech32"
	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/v6/pkg/config/constant"
)

var (
	_ NetworkConfigProvider = DynamicConfigProvider{}
	_ NetworkConfigProvider = StaticConfigProvider{}
)

// NetworkConfigProvider specifies methods required by config consumer.
type NetworkConfigProvider interface {
	GetChainID() constant.ChainID
	GetDenom() string
	GetAddressPrefix() string
	EncodeGenesis(
		ctx context.Context, cosmosClientCtx cosmosclient.Context, basicManager module.BasicManager,
	) ([]byte, error)
	AppState(
		ctx context.Context, cosmosClientCtx cosmosclient.Context, basicManager module.BasicManager,
	) (map[string]json.RawMessage, error)
}

// DynamicConfigProvider provides configuration generated from fields in this structure.
type DynamicConfigProvider struct {
	GenesisInitConfig
}

// WithAccount funds address with balances at genesis.
func (dcp DynamicConfigProvider) WithAccount(accAddress sdk.AccAddress, balances sdk.Coins) DynamicConfigProvider {
	dcp = dcp.clone()
	dcp.BankBalances = append(dcp.BankBalances, banktypes.Balance{
		Address: accAddress.String(),
		Coins:   balances,
	})
	return dcp
}

// GetChainID returns chain ID.
func (dcp DynamicConfigProvider) GetChainID() constant.ChainID {
	return dcp.ChainID
}

// GetDenom returns denom.
func (dcp DynamicConfigProvider) GetDenom() string {
	return dcp.Denom
}

// GetAddressPrefix returns address prefix.
func (dcp DynamicConfigProvider) GetAddressPrefix() string {
	return dcp.AddressPrefix
}

// EncodeGenesis returns encoded genesis doc.
func (dcp DynamicConfigProvider) EncodeGenesis(
	ctx context.Context, cosmosClientCtx cosmosclient.Context, basicManager module.BasicManager,
) ([]byte, error) {
	if dcp.Denom != "" {
		sdk.DefaultBondDenom = dcp.Denom
	}
	genDoc, err := GenDocFromInput(ctx, dcp.GenesisInitConfig, cosmosClientCtx, basicManager)
	if err != nil {
		return nil, errors.Wrap(err, "not able to get genesis doc")
	}

	genDocBytes, err := cmtjson.MarshalIndent(genDoc, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "not able to marshal genesis doc")
	}

	return append(genDocBytes, '\n'), nil
}

// AppState returns the app state from the genesis doc of the network.
func (dcp DynamicConfigProvider) AppState(
	ctx context.Context, cosmosClientCtx cosmosclient.Context, basicManager module.BasicManager,
) (map[string]json.RawMessage, error) {
	if dcp.Denom != "" {
		sdk.DefaultBondDenom = dcp.Denom
	}
	genDoc, err := GenDocFromInput(ctx, dcp.GenesisInitConfig, cosmosClientCtx, basicManager)
	if err != nil {
		return nil, errors.Wrap(err, "not able to get genesis doc")
	}

	var appState map[string]json.RawMessage
	if err := json.Unmarshal(genDoc.AppState, &appState); err != nil {
		return nil, errors.Wrap(err, "not able to parse genesis app state")
	}

	return appState, nil
}

func (dcp DynamicConfigProvider) clone() DynamicConfigProvider {
	dcp.BankBalances = append([]banktypes.Balance{}, dcp.BankBalances...)

	return dcp
}

// StaticConfigProvider provides configuration based on genesis JSON.
type StaticConfigProvider struct {
	content       []byte
	genesisDoc    *tmtypes.GenesisDoc
	denom         string
	addressPrefix string
}

// NewStaticConfigProvider creates new StaticConfigProvider.
func NewStaticConfigProvider(content []byte) StaticConfigProvider {
	genesisDoc, err := tmtypes.GenesisDocFromJSON(content)
	if err != nil {
		panic(err)
	}

	var appStateMapJSONRawMessage map[string]json.RawMessage
	if err := json.Unmarshal(genesisDoc.AppState, &appStateMapJSONRawMessage); err != nil {
		panic(err)
	}

	codec := NewEncodingConfig(
		staking.AppModuleBasic{},
	).Codec
	stakingGenesisState := stakingtypes.GetGenesisStateFromAppState(codec, appStateMapJSONRawMessage)
	bankGenesisState := banktypes.GetGenesisStateFromAppState(codec, appStateMapJSONRawMessage)
	addressPrefix, _, err := bech32.Decode(bankGenesisState.Balances[0].Address, 1023)
	if err != nil {
		panic(err)
	}

	provider := StaticConfigProvider{
		content:       content,
		genesisDoc:    genesisDoc,
		denom:         stakingGenesisState.Params.BondDenom,
		addressPrefix: addressPrefix,
	}

	return provider
}

// GetChainID returns chain ID.
func (scp StaticConfigProvider) GetChainID() constant.ChainID {
	return constant.ChainID(scp.genesisDoc.ChainID)
}

// GetDenom returns denom.
func (scp StaticConfigProvider) GetDenom() string {
	return scp.denom
}

// GetAddressPrefix returns address prefix.
func (scp StaticConfigProvider) GetAddressPrefix() string {
	return scp.addressPrefix
}

// EncodeGenesis returns encoded genesis doc.
func (scp StaticConfigProvider) EncodeGenesis(
	_ context.Context, _ cosmosclient.Context, _ module.BasicManager,
) ([]byte, error) {
	return scp.content, nil
}

// AppState returns the app state from the genesis doc of the network.
func (scp StaticConfigProvider) AppState(
	_ context.Context, _ cosmosclient.Context, _ module.BasicManager,
) (map[string]json.RawMessage, error) {
	var appState map[string]json.RawMessage
	if err := json.Unmarshal(scp.genesisDoc.AppState, &appState); err != nil {
		return nil, errors.Wrap(err, "not able to parse genesis app state")
	}

	return appState, nil
}
