package keyring

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	cosmcrypto "github.com/cosmos/cosmos-sdk/crypto"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	cosmtypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
)

var (
	defaultKeyringKeyName = "default"
	emptyCosmosAddress    = cosmtypes.AccAddress{}
)

// NewCosmosKeyring creates a new keyring from a variety of options. See ConfigOpt and related options.
func NewCosmosKeyring(opts ...ConfigOpt) (cosmtypes.AccAddress, keyring.Keyring, error) {
	config := &cosmosKeyringConfig{}
	for _, optFn := range opts {
		optFn(config)
	}

	switch {
	case len(config.PrivKeyHex) > 0:
		if config.UseLedger {
			err := errors.New("cannot combine ledger and privkey options")
			return emptyCosmosAddress, nil, err
		}

		pkBytes, err := hexToBytes(config.PrivKeyHex)
		if err != nil {
			err = errors.Wrap(err, "failed to hex-decode cosmos account privkey")
			return emptyCosmosAddress, nil, err
		}

		cosmosAccPk := &cosmossecp256k1.PrivKey{
			Key: pkBytes,
		}

		addressFromPk := cosmtypes.AccAddress(cosmosAccPk.PubKey().Address().Bytes())

		var keyName string

		// check that if cosmos 'From' specified separately, it must match the provided privkey,
		if len(config.KeyFrom) > 0 {
			addressFrom, err := cosmtypes.AccAddressFromBech32(config.KeyFrom)
			if err == nil {
				if !bytes.Equal(addressFrom.Bytes(), addressFromPk.Bytes()) {
					err = errors.Errorf("expected account address %s but got %s from the private key", addressFrom.String(), addressFromPk.String())
					return emptyCosmosAddress, nil, err
				}
			} else {
				// use it as a name then
				keyName = config.KeyFrom
			}
		}

		if len(keyName) == 0 {
			keyName = defaultKeyringKeyName
		}

		// wrap a PK into a Keyring
		kb, err := KeyringForPrivKey(keyName, cosmosAccPk)
		return addressFromPk, kb, err

	case len(config.KeyFrom) > 0:
		var fromIsAddress bool
		addressFrom, err := cosmtypes.AccAddressFromBech32(config.KeyFrom)
		if err == nil {
			fromIsAddress = true
		}

		var passReader io.Reader = os.Stdin
		if len(config.KeyPassphrase) > 0 {
			passReader = newPassReader(config.KeyPassphrase)
		}

		var absoluteKeyringDir string
		if filepath.IsAbs(config.KeyringDir) {
			absoluteKeyringDir = config.KeyringDir
		} else {
			absoluteKeyringDir, _ = filepath.Abs(config.KeyringDir)
		}

		kb, err := keyring.New(
			config.KeyringAppName,
			string(config.KeyringBackend),
			absoluteKeyringDir,
			passReader,
		)
		if err != nil {
			err = errors.Wrap(err, "failed to init keyring")
			return emptyCosmosAddress, nil, err
		}

		var keyInfo keyring.Info
		if fromIsAddress {
			if keyInfo, err = kb.KeyByAddress(addressFrom); err != nil {
				err = errors.Wrapf(err, "couldn't find an entry for the key %s in keybase", addressFrom.String())
				return emptyCosmosAddress, nil, err
			}
		} else {
			if keyInfo, err = kb.Key(config.KeyFrom); err != nil {
				err = errors.Wrapf(err, "could not find an entry for the key '%s' in keybase", config.KeyFrom)
				return emptyCosmosAddress, nil, err
			}
		}

		switch keyType := keyInfo.GetType(); keyType {
		case keyring.TypeLocal:
			// kb has a key and it's totally usable
			return keyInfo.GetAddress(), kb, nil
		case keyring.TypeLedger:
			// the kb stores references to ledger keys, so we must explicitly
			// check that. kb doesn't know how to scan HD keys - they must be added manually before
			if config.UseLedger {
				return keyInfo.GetAddress(), kb, nil
			}
			err := errors.Errorf("'%s' key is a ledger reference, enable ledger option", keyInfo.GetName())
			return emptyCosmosAddress, nil, err
		case keyring.TypeOffline:
			err := errors.Errorf("'%s' key is an offline key, not supported yet", keyInfo.GetName())
			return emptyCosmosAddress, nil, err
		case keyring.TypeMulti:
			err := errors.Errorf("'%s' key is an multisig key, not supported yet", keyInfo.GetName())
			return emptyCosmosAddress, nil, err
		default:
			err := errors.Errorf("'%s' key  has unsupported type: %s", keyInfo.GetName(), keyType)
			return emptyCosmosAddress, nil, err
		}

	default:
		err := errors.New("insufficient cosmos key details provided")
		return emptyCosmosAddress, nil, err
	}
}

func newPassReader(pass string) io.Reader {
	return &passReader{
		pass: pass,
		buf:  new(bytes.Buffer),
	}
}

type passReader struct {
	pass string
	buf  *bytes.Buffer
}

var _ io.Reader = &passReader{}

func (r *passReader) Read(p []byte) (n int, err error) {
	n, err = r.buf.Read(p)
	if err == io.EOF || n == 0 {
		r.buf.WriteString(r.pass + "\n")

		n, err = r.buf.Read(p)
	}

	return
}

// KeyringForPrivKey creates a temporary in-mem keyring for a PrivKey.
// Allows to init Context when the key has been provided in plaintext and parsed.
func KeyringForPrivKey(name string, privKey cryptotypes.PrivKey) (keyring.Keyring, error) {
	kb := keyring.NewInMemory()
	tmpPhrase := randPhrase(64)
	armored := cosmcrypto.EncryptArmorPrivKey(privKey, tmpPhrase, privKey.Type())
	err := kb.ImportPrivKey(name, armored, tmpPhrase)
	if err != nil {
		err = errors.Wrap(err, "failed to import privkey")
		return nil, err
	}

	return kb, nil
}

func hexToBytes(str string) ([]byte, error) {
	if strings.HasPrefix(str, "0x") {
		str = str[2:]
	}

	data, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func randPhrase(size int) string {
	buf := make([]byte, size)
	_, err := rand.Read(buf)
	orPanic(err)

	return string(buf)
}

func orPanic(err error) {
	if err != nil {
		log.Panicln()
	}
}
