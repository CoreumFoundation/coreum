package networks_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// invalidSignatureTx has content of validator-0.json with a signature replaced by one in validator-1.json.
var invalidSignatureTx = []byte(`
{
  "body": {
    "messages": [
      {
        "@type": "/cosmos.staking.v1beta1.MsgCreateValidator",
        "description": {
          "moniker": "Mercury",
          "identity": "",
          "website": "",
          "security_contact": "",
          "details": ""
        },
        "commission": {
          "rate": "0.100000000000000000",
          "max_rate": "0.200000000000000000",
          "max_change_rate": "0.010000000000000000"
        },
        "min_self_delegation": "1",
        "delegator_address": "devcore15eqsya33vx9p5zt7ad8fg3k674tlsllk3pvqp6",
        "validator_address": "devcorevaloper15eqsya33vx9p5zt7ad8fg3k674tlsllkg7j9w0",
        "pubkey": {
          "@type": "/cosmos.crypto.secp256k1.PubKey",
          "key": "Ai+ulhQOTk78QqrOwr9ySdZQ+N6F5PzkRU7LsYXA24Gl"
        },
        "value": {
          "denom": "ducore",
          "amount": "10000000000000"
        }
      }
    ],
    "memo": "",
    "timeout_height": "0",
    "extension_options": [],
    "non_critical_extension_options": []
  },
  "auth_info": {
    "signer_infos": [
      {
        "public_key": {
          "@type": "/cosmos.crypto.secp256k1.PubKey",
          "key": "Ai+ulhQOTk78QqrOwr9ySdZQ+N6F5PzkRU7LsYXA24Gl"
        },
        "mode_info": {
          "single": {
            "mode": "SIGN_MODE_DIRECT"
          }
        },
        "sequence": "0"
      }
    ],
    "fee": {
      "amount": [],
      "gas_limit": "0",
      "payer": "",
      "granter": ""
    }
  },
  "signatures": [
    "C15uNvGlYjjTBzYLLUtKClKRVF1CiicRvB+vr2q4FQs5RQflaMvXkF15AHPamIIrZ7zvZKIg5p6/ZZl1Lkxl4Q=="
  ]
}
`)

func init() {
	network, _ := config.NetworkByChainID(config.Devnet)
	// Since we have a single network currently (devnet) we can seal config here.
	// The idea is to add SetSDKConfigNoSeal to Network once we need to validate txs for multiple networks.
	network.SetSDKConfig()
}

// The purpose of this test is to verify that validateGenesisTxSignature func works properly for negative case.
// Because invalidSignatureTx passes cosmos SDK `tx validate-signatures --offline` successfully even though
// the signature is obviously invalid.
func TestInvalidTxSignature(t *testing.T) {
	network, err := config.NetworkByChainID(config.Devnet)
	assert.NoError(t, err)

	clientContext := tx.NewClientContext(app.ModuleBasics).WithChainID(string(network.ChainID()))

	// Check on a sample tx with a wrong signature that signature verification works properly.
	sdkTx, err := clientContext.TxConfig().TxJSONDecoder()(invalidSignatureTx)
	assert.NoError(t, err)
	assert.ErrorContains(t, validateGenesisTxSignature(clientContext, sdkTx), "signature verification failed")
}

func TestNetworkTxSignatures(t *testing.T) {
	network, err := config.NetworkByChainID(config.Devnet)
	assert.NoError(t, err)

	clientContext := tx.NewClientContext(app.ModuleBasics).WithChainID(string(network.ChainID()))

	// Check real network txs.
	for _, rawTx := range network.GenTxs() {
		sdkTx, err := clientContext.TxConfig().TxJSONDecoder()(rawTx)
		assert.NoError(t, err)

		assert.NoError(t, validateGenesisTxSignature(clientContext, sdkTx))
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
