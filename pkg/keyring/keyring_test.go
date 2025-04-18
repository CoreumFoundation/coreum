package keyring

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/std"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/pkg/config/constant"
)

// TestMultisigAddressGeneration checks if the same multisig address is generated every time.
//
//nolint:lll // this code contains mnemonic that cannot be broken down.
func TestMultisigAddressGeneration(t *testing.T) {
	t.Parallel()

	interfaceRegistry := codectypes.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	cdc := codec.NewProtoCodec(interfaceRegistry)
	keystore := keyring.NewInMemory(cdc)

	accAddr1 := importMnemonic(keystore, "human scan federal dose project toward nominee chief wheel swamp drop pitch olympic job inner critic mask laundry corn dice fame expect brave feel")
	assert.Equal(t, "cosmos15lu0zdjkqzvh7x7pevp3n08anzt49sz87l7vry", accAddr1.String())

	signerKeyInfo1, err := keystore.KeyByAddress(accAddr1)
	require.NoError(t, err)

	accAddr2 := importMnemonic(keystore, "dinner liar trust decrease angry apart ladder dance leisure flock super hollow such much ridge planet pill crazy inherit limit submit size absurd drive")
	assert.Equal(t, "cosmos13ym5fg96sg442mgpta0xnd064dcv9tqskyj6mp", accAddr2.String())

	signerKeyInfo2, err := keystore.KeyByAddress(accAddr2)
	require.NoError(t, err)

	signer1PubKey, err := signerKeyInfo1.GetPubKey()
	require.NoError(t, err)
	signer2PubKey, err := signerKeyInfo2.GetPubKey()
	require.NoError(t, err)
	multisigPublicKey := multisig.NewLegacyAminoPubKey(2, []types.PubKey{
		signer1PubKey,
		signer2PubKey,
	})

	expectedMultisigAddr := "cosmos1gst5qagnzl36jx77r5gtcwg6gfcuyc2e66g64r"
	assert.Equal(t, expectedMultisigAddr, sdk.AccAddress(multisigPublicKey.Address()).String())
}

func importMnemonic(keystore keyring.Keyring, mnemonic string) sdk.AccAddress {
	keyInfo, err := keystore.NewAccount(
		uuid.New().String(),
		mnemonic,
		"",
		hd.CreateHDPath(constant.CoinType, 0, 0).String(),
		hd.Secp256k1,
	)
	if err != nil {
		panic(err)
	}

	address, err := keyInfo.GetAddress()
	if err != nil {
		panic(err)
	}

	return address
}
