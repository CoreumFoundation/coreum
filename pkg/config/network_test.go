package config_test

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
	"unsafe"

	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	feemodeltypes "github.com/CoreumFoundation/coreum/x/feemodel/types"
)

func TestAddFundsToGenesis(t *testing.T) {
	assertT := assert.New(t)
	requireT := require.New(t)

	n, err := config.NetworkByChainID(constant.ChainIDDev)
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

	requireT.Len(n.FundedAccounts(), 6)
	requireT.Len(n.GenTxs(), 5)

	// re-init the config and check that length remains the same
	n, err = config.NetworkByChainID(constant.ChainIDDev)
	requireT.NoError(err)
	requireT.Len(n.FundedAccounts(), 5)
	requireT.Len(n.GenTxs(), 4)
}

func TestValidateAllGenesis(t *testing.T) {
	assertT := assert.New(t)
	encCfg := config.NewEncodingConfig(app.ModuleBasics)

	for _, n := range config.EnabledNetworks() {
		unsealConfig()
		n.SetSDKConfig()
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

func TestValidateAllGenTxs(t *testing.T) {
	for _, n := range config.EnabledNetworks() {
		unsealConfig()
		n.SetSDKConfig()

		clientCtx := tx.NewClientContext(app.ModuleBasics).WithChainID(string(n.ChainID()))

		// Check real n txs.
		for _, rawTx := range n.GenTxs() {
			sdkTx, err := clientCtx.TxConfig().TxJSONDecoder()(rawTx)
			assert.NoError(t, err)

			assert.NoError(t, validateGenesisTxSignature(clientCtx, sdkTx))
		}
	}
}

// https://github.com/cosmos/cosmos-sdk/tree/v0.45.5/x/auth/client/cli/validate_sigs.go:L61
// Original code was significantly refactored & simplified to cover our use-case.
// Note that this func handles only genesis txs signature validation because of
// hardcoded account number & sequence to avoid real network requests.
func validateGenesisTxSignature(clientCtx tx.ClientContext, tx sdk.Tx) error {
	signedTx, ok := tx.(authsigning.SigVerifiableTx)
	if !ok {
		return errors.New("failed to convert Tx to SigVerifiableTx")
	}

	sigs, err := signedTx.GetSignaturesV2()
	if err != nil {
		return errors.Wrap(err, "failed to get tx signature")
	}

	signers := signedTx.GetSigners()
	signModeHandler := clientCtx.TxConfig().SignModeHandler()

	for i, sig := range sigs {
		pubKey := sig.PubKey

		if i >= len(signers) || !sdk.AccAddress(pubKey.Address()).Equals(signers[i]) {
			return errors.New("signature does not match its respective signer")
		}

		// AccountNumber & Sequence is set to 0 because txs we validate here are genesis txs.
		signingData := authsigning.SignerData{
			ChainID:       clientCtx.ChainID(),
			AccountNumber: 0,
			Sequence:      0,
		}
		err = authsigning.VerifySignature(pubKey, signingData, sig.Data, signModeHandler, signedTx)
		if err != nil {
			return errors.Wrap(err, "signature verification failed")
		}
	}

	return nil
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
