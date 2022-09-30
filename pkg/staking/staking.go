package staking

import (
	"crypto/ed25519"

	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// PrepareTxStakingCreateValidator generates transaction of type MsgCreateValidator
func PrepareTxStakingCreateValidator(
	clientCtx tx.ClientContext,
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

	signedTx, err := tx.Sign(clientCtx, tx.BaseInput{Signer: types.Wallet{Key: stakerPrivateKey}}, msg)
	if err != nil {
		return nil, errors.Wrap(err, "unable to sign transaction")
	}
	encodedTx, err := clientCtx.TxConfig().TxJSONEncoder()(signedTx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to encode transaction")
	}
	return encodedTx, nil
}
