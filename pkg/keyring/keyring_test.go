package keyring

import (
	"encoding/hex"
	"fmt"
	"os"
	"testing"

	"github.com/CoreumFoundation/coreum/cmd/cored/cosmoscmd"
	cosmcrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	cosmkeyring "github.com/cosmos/cosmos-sdk/crypto/keyring"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	cosmoscmd.SetPrefixes("core")
	os.Exit(m.Run())
}

func TestKeyFrom(t *testing.T) {
	requireT := require.New(t)

	accAddr, kb, err := NewCosmosKeyring(
		WithPrivKeyHex("61cb29d4cfe4c82eec31effa4b65495ef3dfbf343d659a283bafd15e0bc50fb2"),
	)
	requireT.NoError(err)
	requireT.Equal("core1lqepwqlvn5jnam0gadgtzhkjgwuyt8jnlydtes", accAddr.String())

	info, err := kb.KeyByAddress(accAddr)
	requireT.NoError(err)
	requireT.Equal(cosmkeyring.TypeLocal, info.GetType())
	requireT.Equal(hd.Secp256k1Type, info.GetAlgo())

	// showPrivkey(kb, accAddr)

	res, pubkey, err := kb.SignByAddress(accAddr, []byte("test"))
	requireT.NoError(err)
	requireT.Equal(info.GetPubKey(), pubkey)
	requireT.Equal(testSig, res)
}

func TestKeyringFile(t *testing.T) {
	requireT := require.New(t)

	accAddr, kb, err := NewCosmosKeyring(
		WithKeyringBackend(BackendFile),
		WithKeyringDir("./test_fixtures"),
		WithKeyFrom("test"),
		WithKeyPassphrase("test12345678"),
	)
	requireT.NoError(err)
	requireT.Equal("core1lqepwqlvn5jnam0gadgtzhkjgwuyt8jnlydtes", accAddr.String())

	info, err := kb.KeyByAddress(accAddr)
	requireT.NoError(err)
	requireT.Equal(cosmkeyring.TypeLocal, info.GetType())
	requireT.Equal(hd.Secp256k1Type, info.GetAlgo())
	requireT.Equal("test", info.GetName())

	// showPrivkey(kb, accAddr)

	res, pubkey, err := kb.SignByAddress(accAddr, []byte("test"))
	requireT.NoError(err)
	requireT.Equal(info.GetPubKey(), pubkey)
	requireT.Equal(testSig, res)
}

var testSig = []byte{
	0x55, 0x92, 0xf7, 0x93, 0x81, 0x7b, 0x16, 0xf3, 0xe7,
	0x90, 0xe5, 0x83, 0xf6, 0xb9, 0x57, 0x29, 0x58, 0xb6,
	0x47, 0x94, 0xba, 0xe6, 0x8, 0xd4, 0x24, 0x19, 0x65, 0x2e,
	0x1c, 0xdc, 0xf3, 0x34, 0xa, 0x11, 0x54, 0x69, 0xe4, 0xff,
	0xd, 0xa7, 0x35, 0xd7, 0xe4, 0x85, 0x4d, 0x22, 0x89, 0xf4,
	0x14, 0x53, 0x9b, 0xa2, 0xbc, 0x8, 0x25, 0x4b, 0x64, 0x30,
	0x66, 0xc3, 0xfe, 0x98, 0xc0, 0xa0,
}

func showPrivkey(kb cosmkeyring.Keyring, accAddr sdk.AccAddress) {
	armor, _ := kb.ExportPrivKeyArmorByAddress(accAddr, "")
	privKey, _, _ := cosmcrypto.UnarmorDecryptPrivKey(armor, "")
	fmt.Println("[PRIV]", hex.EncodeToString(privKey.Bytes()))
}
