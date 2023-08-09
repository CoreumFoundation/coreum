package config

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"
	"time"

	"github.com/cosmos/btcutil/bech32"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authcosmostypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/CoreumFoundation/coreum/v2/pkg/config/constant"
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
	EncodeGenesis() ([]byte, error)
	AppState() (map[string]json.RawMessage, error)
}

// DynamicConfigProvider provides configuration generated from fields in this structure.
type DynamicConfigProvider struct {
	GenesisTemplate    string
	ChainID            constant.ChainID
	Denom              string
	AddressPrefix      string
	GenesisTime        time.Time
	BlockTimeIota      time.Duration
	GovConfig          GovConfig
	CustomParamsConfig CustomParamsConfig
	FundedAccounts     []FundedAccount
	GenTxs             []json.RawMessage
}

// WithAccount funds address with balances at genesis.
func (dcp DynamicConfigProvider) WithAccount(accAddress sdk.AccAddress, balances sdk.Coins) DynamicConfigProvider {
	dcp = dcp.clone()
	dcp.FundedAccounts = append(dcp.FundedAccounts, FundedAccount{
		Address:  accAddress.String(),
		Balances: balances,
	})
	return dcp
}

// WithGenesisTx adds transaction to the genesis file.
func (dcp DynamicConfigProvider) WithGenesisTx(signedTx json.RawMessage) DynamicConfigProvider {
	dcp = dcp.clone()
	dcp.GenTxs = append(dcp.GenTxs, signedTx)
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
func (dcp DynamicConfigProvider) EncodeGenesis() ([]byte, error) {
	genesisDoc, err := dcp.genesisDoc()
	if err != nil {
		return nil, errors.Wrap(err, "not able to get genesis doc")
	}

	bs, err := tmjson.MarshalIndent(genesisDoc, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "not able to marshal genesis doc")
	}

	return append(bs, '\n'), nil
}

// AppState returns the app state from the genesis doc of the network.
func (dcp DynamicConfigProvider) AppState() (map[string]json.RawMessage, error) {
	codec := NewEncodingConfig(module.NewBasicManager(
		auth.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		genutil.AppModuleBasic{},
		bank.AppModuleBasic{},
	)).Codec

	genesisJSON, err := dcp.genesisByTemplate()
	if err != nil {
		return nil, errors.Wrap(err, "not able get genesis")
	}

	genesisDoc, err := tmtypes.GenesisDocFromJSON(genesisJSON)
	if err != nil {
		return nil, errors.Wrap(err, "not able to parse genesis json bytes")
	}

	var appState map[string]json.RawMessage
	if err := json.Unmarshal(genesisDoc.AppState, &appState); err != nil {
		return nil, errors.Wrap(err, "not able to parse genesis app state")
	}

	authState := authcosmostypes.GetGenesisStateFromAppState(codec, appState)
	accountState, err := authcosmostypes.UnpackAccounts(authState.Accounts)
	if err != nil {
		return nil, errors.Wrap(err, "not able to unpack auth accounts")
	}

	genutilState := genutiltypes.GetGenesisStateFromAppState(codec, appState)
	bankState := banktypes.GetGenesisStateFromAppState(codec, appState)

	if err := validateNoDuplicateFundedAccounts(dcp.FundedAccounts); err != nil {
		return nil, err
	}

	for _, fundedAcc := range dcp.FundedAccounts {
		accountState = applyFundedAccountToGenesis(fundedAcc, accountState, bankState)
	}

	genutilState.GenTxs = append(genutilState.GenTxs, dcp.GenTxs...)

	genutiltypes.SetGenesisStateInAppState(codec, appState, genutilState)
	authState.Accounts, err = authcosmostypes.PackAccounts(authcosmostypes.SanitizeGenesisAccounts(accountState))
	if err != nil {
		return nil, errors.Wrap(err, "not able to sanitize and pack accounts")
	}
	appState[authcosmostypes.ModuleName] = codec.MustMarshalJSON(&authState)

	bankState.Balances = banktypes.SanitizeGenesisBalances(bankState.Balances)
	appState[banktypes.ModuleName] = codec.MustMarshalJSON(bankState)

	return appState, nil
}

// GenesisDoc returns the genesis doc of the network.
func (dcp DynamicConfigProvider) genesisDoc() (*tmtypes.GenesisDoc, error) {
	genesisJSON, err := dcp.genesisByTemplate()
	if err != nil {
		return nil, errors.Wrap(err, "not able get genesis")
	}

	genesisDoc, err := tmtypes.GenesisDocFromJSON(genesisJSON)
	if err != nil {
		return nil, errors.Wrap(err, "not able to parse genesis json bytes")
	}

	appState, err := dcp.AppState()
	if err != nil {
		return nil, err
	}

	genesisDoc.AppState, err = json.MarshalIndent(appState, "", "  ")
	if err != nil {
		return nil, err
	}

	return genesisDoc, nil
}

func (dcp DynamicConfigProvider) clone() DynamicConfigProvider {
	dcp.FundedAccounts = append([]FundedAccount{}, dcp.FundedAccounts...)
	dcp.GenTxs = append([]json.RawMessage{}, dcp.GenTxs...)

	return dcp
}

func (dcp DynamicConfigProvider) genesisByTemplate() ([]byte, error) {
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
	}

	genesisBuf := new(bytes.Buffer)
	err := template.Must(template.New("genesis").Funcs(funcMap).Parse(dcp.GenesisTemplate)).Execute(genesisBuf, struct {
		GenesisTimeUTC     string
		ChainID            constant.ChainID
		Denom              string
		Gov                GovConfig
		CustomParamsConfig CustomParamsConfig
		BlockTimeIotaMS    int64
	}{
		GenesisTimeUTC:     dcp.GenesisTime.UTC().Format(time.RFC3339),
		ChainID:            dcp.ChainID,
		Denom:              dcp.Denom,
		Gov:                dcp.GovConfig,
		CustomParamsConfig: dcp.CustomParamsConfig,
		BlockTimeIotaMS:    dcp.BlockTimeIota.Milliseconds(),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to template genesis file")
	}

	return genesisBuf.Bytes(), nil
}

func validateNoDuplicateFundedAccounts(accounts []FundedAccount) error {
	accountsIndexMap := map[string]interface{}{}
	for _, fundEntry := range accounts {
		fundEntry := fundEntry
		_, exists := accountsIndexMap[fundEntry.Address]
		if exists {
			return errors.New("duplicate funded account is not allowed")
		}
		accountsIndexMap[fundEntry.Address] = true
	}

	return nil
}

func applyFundedAccountToGenesis(
	fa FundedAccount,
	accountState authcosmostypes.GenesisAccounts,
	bankState *banktypes.GenesisState,
) authcosmostypes.GenesisAccounts {
	accountAddress := sdk.MustAccAddressFromBech32(fa.Address)
	accountState = append(accountState, authcosmostypes.NewBaseAccount(accountAddress, nil, 0, 0))
	coins := fa.Balances
	bankState.Balances = append(
		bankState.Balances,
		banktypes.Balance{Address: accountAddress.String(), Coins: coins},
	)
	bankState.Supply = bankState.Supply.Add(coins...)

	return accountState
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

	codec := NewEncodingConfig(module.NewBasicManager(
		staking.AppModuleBasic{},
	)).Codec
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

// StaticConfigProvider provides configuration based on genesis JSON.
type StaticConfigProvider struct {
	content       []byte
	genesisDoc    *tmtypes.GenesisDoc
	denom         string
	addressPrefix string
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
func (scp StaticConfigProvider) EncodeGenesis() ([]byte, error) {
	return scp.content, nil
}

// AppState returns the app state from the genesis doc of the network.
func (scp StaticConfigProvider) AppState() (map[string]json.RawMessage, error) {
	var appState map[string]json.RawMessage
	if err := json.Unmarshal(scp.genesisDoc.AppState, &appState); err != nil {
		return nil, errors.Wrap(err, "not able to parse genesis app state")
	}

	return appState, nil
}
