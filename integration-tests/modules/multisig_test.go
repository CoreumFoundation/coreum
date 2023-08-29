//go:build integrationtests

package modules

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	integrationtests "github.com/CoreumFoundation/coreum/v2/integration-tests"
)

// TestMultisigAddressGeneration checks if the same multisig address is generated every time.
func TestMultisigAddressGeneration(t *testing.T) {
	t.Parallel()
	_, chain := integrationtests.NewCoreumTestingContext(t)

	accAddr1 := chain.ImportMnemonic("human scan federal dose project toward nominee chief wheel swamp drop pitch olympic job inner critic mask laundry corn dice fame expect brave feel")
	assert.Equal(t, "devcore15lu0zdjkqzvh7x7pevp3n08anzt49sz8l0t42r", accAddr1.String())

	signerKeyInfo1, err := chain.ClientContext.Keyring().KeyByAddress(accAddr1)
	assert.Nil(t, err)

	accAddr2 := chain.ImportMnemonic("dinner liar trust decrease angry apart ladder dance leisure flock super hollow such much ridge planet pill crazy inherit limit submit size absurd drive")
	assert.Equal(t, "devcore13ym5fg96sg442mgpta0xnd064dcv9tqsh58rjx", accAddr2.String())

	signerKeyInfo2, err := chain.ClientContext.Keyring().KeyByAddress(accAddr2)
	assert.Nil(t, err)

	multisigPublicKey := multisig.NewLegacyAminoPubKey(2, []types.PubKey{
		signerKeyInfo1.GetPubKey(),
		signerKeyInfo2.GetPubKey(),
	})

	expectedMultisigAddr := "devcore1gst5qagnzl36jx77r5gtcwg6gfcuyc2em2aruy"
	assert.Equal(t, expectedMultisigAddr, sdk.AccAddress(multisigPublicKey.Address()).String())
}
