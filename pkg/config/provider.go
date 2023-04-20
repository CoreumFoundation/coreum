package config

import (
	"bytes"
	"encoding/json"
	"strings"
	"text/template"
	"time"

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
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/CoreumFoundation/coreum/pkg/config/constant"
)

// NetworkConfigProvider specifies methods required by config consumer.
type NetworkConfigProvider interface {
	GetChainID() constant.ChainID
	GetDenom() string
	GenesisDoc() (*tmtypes.GenesisDoc, error)
}

// DirectConfigProvider provides configuration generated from fields in this structure.
type DirectConfigProvider struct {
	ChainID            constant.ChainID
	GenesisTime        time.Time
	GovConfig          GovConfig
	CustomParamsConfig CustomParamsConfig
	FundedAccounts     []FundedAccount
	GenTxs             []json.RawMessage

	Denom string
}

// WithAccount funds address with balances at genesis.
func (dcp DirectConfigProvider) WithAccount(accAddress sdk.AccAddress, balances sdk.Coins) DirectConfigProvider {
	dcp = dcp.clone()
	dcp.FundedAccounts = append(dcp.FundedAccounts, FundedAccount{
		Address:  accAddress.String(),
		Balances: balances,
	})
	return dcp
}

// WithGenesisTx adds transaction to the genesis file.
func (dcp DirectConfigProvider) WithGenesisTx(signedTx json.RawMessage) DirectConfigProvider {
	dcp = dcp.clone()
	dcp.GenTxs = append(dcp.GenTxs, signedTx)
	return dcp
}

// GetChainID returns chain ID.
func (dcp DirectConfigProvider) GetChainID() constant.ChainID {
	return dcp.ChainID
}

// GetDenom returns denom.
func (dcp DirectConfigProvider) GetDenom() string {
	return dcp.Denom
}

// GenesisDoc returns the genesis doc of the network.
func (dcp DirectConfigProvider) GenesisDoc() (*tmtypes.GenesisDoc, error) {
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

	if err = json.Unmarshal(genesisDoc.AppState, &appState); err != nil {
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

	genesisDoc.AppState, err = json.MarshalIndent(appState, "", "  ")
	if err != nil {
		return nil, err
	}

	return genesisDoc, nil
}

func (dcp DirectConfigProvider) clone() DirectConfigProvider {
	dcp.FundedAccounts = append([]FundedAccount{}, dcp.FundedAccounts...)
	dcp.GenTxs = append([]json.RawMessage{}, dcp.GenTxs...)

	return dcp
}

func (dcp DirectConfigProvider) genesisByTemplate() ([]byte, error) {
	funcMap := template.FuncMap{
		"ToUpper": strings.ToUpper,
	}

	genesisBuf := new(bytes.Buffer)
	err := template.Must(template.New("genesis").Funcs(funcMap).Parse(genesisTemplate)).Execute(genesisBuf, struct {
		GenesisTimeUTC     string
		ChainID            constant.ChainID
		Gov                GovConfig
		CustomParamsConfig CustomParamsConfig
	}{
		GenesisTimeUTC:     dcp.GenesisTime.UTC().Format(time.RFC3339),
		ChainID:            dcp.ChainID,
		Gov:                dcp.GovConfig,
		CustomParamsConfig: dcp.CustomParamsConfig,
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

// NewJSONConfigProvider creates new JSONConfigProvider.
func NewJSONConfigProvider(content []byte) JSONConfigProvider {
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

	provider := JSONConfigProvider{
		genesisDoc: genesisDoc,
		denom:      stakingGenesisState.Params.BondDenom,
	}

	return provider
}

// JSONConfigProvider provides configuration based on genesis JSON.
type JSONConfigProvider struct {
	genesisDoc *tmtypes.GenesisDoc
	denom      string
}

// GetChainID returns chain ID.
func (jcp JSONConfigProvider) GetChainID() constant.ChainID {
	return constant.ChainID(jcp.genesisDoc.ChainID)
}

// GetDenom returns denom.
func (jcp JSONConfigProvider) GetDenom() string {
	return jcp.denom
}

// GenesisDoc returns the genesis doc of the network.
func (jcp JSONConfigProvider) GenesisDoc() (*tmtypes.GenesisDoc, error) {
	return jcp.genesisDoc, nil
}
