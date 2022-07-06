package config

import (
	"bytes"
	"crypto/ed25519"
	_ "embed"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/CoreumFoundation/coreum/pkg/types"
	"github.com/pkg/errors"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"
)

// Genesis is responsible for creating genesis configuration for coreum network
type Genesis struct {
	codec        codec.Codec
	genesisDoc   *tmtypes.GenesisDoc
	mu           sync.Mutex
	appState     map[string]json.RawMessage
	accountState authtypes.GenesisAccounts
	bankState    *banktypes.GenesisState
	genutilState *genutiltypes.GenesisState
	authState    authtypes.GenesisState
}

// ChainID returns ID of chain
func (g *Genesis) ChainID() string {
	return g.genesisDoc.ChainID
}

// FundAccount funds address with balances at genesis
func (g *Genesis) FundAccount(publicKey types.Secp256k1PublicKey, balances string) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	pubKey := cosmossecp256k1.PubKey{Key: publicKey}
	accountAddress := sdk.AccAddress(pubKey.Address())
	g.accountState = append(g.accountState, authtypes.NewBaseAccount(accountAddress, nil, 0, 0))
	coins, err := sdk.ParseCoinsNormalized(balances)
	if err != nil {
		return errors.Wrapf(err, "not able to parse balances %s", balances)
	}

	g.bankState.Balances = append(g.bankState.Balances, banktypes.Balance{Address: accountAddress.String(), Coins: coins})
	g.bankState.Supply = g.bankState.Supply.Add(coins...)
	return nil
}

// AddGenesisTx adds transaction to the genesis file
func (g *Genesis) AddGenesisTx(signedTx json.RawMessage) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.genutilState.GenTxs = append(g.genutilState.GenTxs, signedTx)
}

// EncodeAsJSON returns json encoded representation
func (g *Genesis) EncodeAsJSON() ([]byte, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	genutiltypes.SetGenesisStateInAppState(g.codec, g.appState, g.genutilState)
	var err error
	g.authState.Accounts, err = authtypes.PackAccounts(authtypes.SanitizeGenesisAccounts(g.accountState))
	if err != nil {
		return nil, errors.Wrap(err, "not able to sanitize and pack accounts")
	}
	g.appState[authtypes.ModuleName] = g.codec.MustMarshalJSON(&g.authState)

	g.bankState.Balances = banktypes.SanitizeGenesisBalances(g.bankState.Balances)
	g.appState[banktypes.ModuleName] = g.codec.MustMarshalJSON(g.bankState)

	g.genesisDoc.AppState, err = json.MarshalIndent(g.appState, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "not able to marshal app state")
	}

	bs, err := tmjson.MarshalIndent(g.genesisDoc, "", "  ")
	if err != nil {
		return nil, errors.Wrap(err, "not able to marshal genesis doc")
	}

	return bs, nil
}

// Save saves genesis configuration
func (g *Genesis) Save(homeDir string) error {
	genDocBytes, err := g.EncodeAsJSON()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(homeDir+"/config", 0o700); err != nil {
		return errors.Wrap(err, "unable to make config directory")
	}

	err = ioutil.WriteFile(homeDir+"/config/genesis.json", genDocBytes, 0644)
	return errors.Wrap(err, "unable to write genesis bytes to file")
}

// GenerateAddValidatorTx generates transaction of type MsgCreateValidator
func GenerateAddValidatorTx(
	clientCtx client.Context,
	validatorPublicKey ed25519.PublicKey,
	stakerPrivateKey types.Secp256k1PrivateKey,
	stakedBalance string,
) ([]byte, error) {
	amount, err := sdk.ParseCoinNormalized(stakedBalance)
	if err != nil {
		return nil, errors.Wrapf(err, "not able to parse stake balances %s", stakedBalance)
	}

	commission := stakingtypes.CommissionRates{
		Rate:          sdk.MustNewDecFromStr("0.1"),
		MaxRate:       sdk.MustNewDecFromStr("0.2"),
		MaxChangeRate: sdk.MustNewDecFromStr("0.01"),
	}

	valPubKey := &cosmosed25519.PubKey{Key: validatorPublicKey}
	stakerPrivKey := &cosmossecp256k1.PrivKey{Key: stakerPrivateKey}
	stakerAddress := sdk.AccAddress(stakerPrivKey.PubKey().Address())

	msg, err := stakingtypes.NewMsgCreateValidator(sdk.ValAddress(stakerAddress), valPubKey, amount, stakingtypes.Description{Moniker: stakerAddress.String()}, commission, sdk.OneInt())
	if err != nil {
		return nil, errors.Wrap(err, "not able to make CreateValidatorMessage")
	}

	tx, err := signTx(clientCtx, stakerPrivateKey, 0, 0, msg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to sign transaction")
	}
	encodedTx, err := clientCtx.TxConfig.TxJSONEncoder()(tx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode transaction")
	}
	return encodedTx, nil
}

//go:embed genesis/genesis.tmpl.json
var genesisTemplate string

func genesis(n Network) ([]byte, error) {
	genesisBuf := new(bytes.Buffer)
	err := template.Must(template.New("genesis").Parse(genesisTemplate)).Execute(genesisBuf, struct {
		GenesisTimeUTC string
		ChainID        ChainID
		TokenSymbol    string
	}{
		GenesisTimeUTC: n.genesisTime.UTC().Format(time.RFC3339),
		ChainID:        n.chainID,
		TokenSymbol:    n.tokenSymbol,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to template genesis file")
	}
	return genesisBuf.Bytes(), nil
}

func signTx(clientCtx client.Context, signerKey types.Secp256k1PrivateKey, accNum, accSeq uint64, msg sdk.Msg) (authsigning.Tx, error) {
	privKey := &cosmossecp256k1.PrivKey{Key: signerKey}
	txBuilder := clientCtx.TxConfig.NewTxBuilder()
	txBuilder.SetGasLimit(200000)
	err := txBuilder.SetMsgs(msg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set message on tx builder")
	}

	signerData := authsigning.SignerData{
		ChainID:       clientCtx.ChainID,
		AccountNumber: accNum,
		Sequence:      accSeq,
	}
	sigData := &signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   privKey.PubKey(),
		Data:     sigData,
		Sequence: accSeq,
	}
	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set signature on tx builder")
	}

	bytesToSign, err := clientCtx.TxConfig.SignModeHandler().GetSignBytes(signing.SignMode_SIGN_MODE_DIRECT, signerData, txBuilder.GetTx())
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode bytes to sign")
	}
	sigBytes, err := privKey.Sign(bytesToSign)
	if err != nil {
		return nil, errors.Wrap(err, "unable to sign")
	}

	sigData.Signature = sigBytes

	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to set signature on tx builder")
	}

	return txBuilder.GetTx(), nil
}
