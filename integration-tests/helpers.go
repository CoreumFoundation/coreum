package integrationtests

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

// SkipUnsafe will skip the tests that are not safe to run against a real running chain.
// unsafe tests can only be run against a locally running chain since they modify parameters
// of the chain.
func SkipUnsafe(t *testing.T) {
	if !cfg.RunUnsafe {
		t.SkipNow()
	}
}

// GenRandomAddress generates a random secp256k1 bech32 encoded address that starts with
// the given prefix.
func GenRandomAddress(prefix string) (string, error) {
	privateKey := secp256k1.GenPrivKey()
	return bech32.ConvertAndEncode(prefix, privateKey.PubKey().Address())
}
