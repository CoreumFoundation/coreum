package staking

import (
	"crypto/ed25519"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// PrepareTxStakingCreateValidator generates transaction of type MsgCreateValidator
func PrepareTxStakingCreateValidator(
	chainID config.ChainID,
	txConfig client.TxConfig,
	validatorPublicKey ed25519.PublicKey,
	stakerPrivateKey cosmossecp256k1.PrivKey,
	stakedBalance sdk.Coin,
) ([]byte, error) {
	// the passphrase here is the trick to import the private key into the keyring
	const passphrase = "tmp"

	commission := stakingtypes.CommissionRates{
		Rate:          sdk.MustNewDecFromStr("0.1"),
		MaxRate:       sdk.MustNewDecFromStr("0.2"),
		MaxChangeRate: sdk.MustNewDecFromStr("0.01"),
	}

	stakerAddress := sdk.AccAddress(stakerPrivateKey.PubKey().Address())
	msg, err := stakingtypes.NewMsgCreateValidator(sdk.ValAddress(stakerAddress), &cosmosed25519.PubKey{Key: validatorPublicKey}, stakedBalance, stakingtypes.Description{Moniker: stakerAddress.String()}, commission, sdk.OneInt())
	if err != nil {
		return nil, errors.Wrap(err, "not able to make CreateValidatorMessage")
	}

	if err := msg.ValidateBasic(); err != nil {
		return nil, errors.Wrap(err, "not able to validate CreateValidatorMessage")
	}

	inMemKeyring := keyring.NewInMemory()

	armor := crypto.EncryptArmorPrivKey(&stakerPrivateKey, passphrase, string(hd.Secp256k1Type))
	if err := inMemKeyring.ImportPrivKey(stakerAddress.String(), armor, passphrase); err != nil {
		return nil, errors.Wrap(err, "not able to import private key into new in memory keyring")
	}

	txf := tx.Factory{}.
		WithChainID(string(chainID)).
		WithKeybase(inMemKeyring).
		WithTxConfig(txConfig)

	txBuilder, err := txf.BuildUnsignedTx(msg)
	if err != nil {
		return nil, err
	}

	if err := tx.Sign(txf, stakerAddress.String(), txBuilder, true); err != nil {
		return nil, err
	}

	txBytes, err := txConfig.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	return txBytes, nil
}
