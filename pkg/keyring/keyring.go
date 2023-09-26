package keyring

import (
	"sync"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ keyring.Keyring = ConcurrentSafeKeyring{}

// ConcurrentSafeKeyring wraps standard Cosmos SDK keyring implementation to make it concurrent safe.
// Since we run our integration tests in parallel from time to time we get: "concurrent map read and map write"
// so ConcurrentSafeKeyring wraps all methods of keyring and locks mutex before calling method.
type ConcurrentSafeKeyring struct {
	mu *sync.RWMutex
	kr keyring.Keyring
}

// NewConcurrentSafeKeyring returns new instance of ConcurrentSafeKeyring.
func NewConcurrentSafeKeyring(kr keyring.Keyring) ConcurrentSafeKeyring {
	return ConcurrentSafeKeyring{
		mu: &sync.RWMutex{},
		kr: kr,
	}
}

// Read operations:

// SupportedAlgorithms supported signing algorithms for Keyring and Ledger respectively.
func (csk ConcurrentSafeKeyring) SupportedAlgorithms() (keyring.SigningAlgoList, keyring.SigningAlgoList) {
	return csk.kr.SupportedAlgorithms()
}

// List lists all keys.
func (csk ConcurrentSafeKeyring) List() ([]*keyring.Record, error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.List()
}

// Key returns keys by uid.
func (csk ConcurrentSafeKeyring) Key(uid string) (*keyring.Record, error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.Key(uid)
}

// KeyByAddress return keys by address.
func (csk ConcurrentSafeKeyring) KeyByAddress(address sdk.Address) (*keyring.Record, error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.KeyByAddress(address)
}

// ExportPubKeyArmor exports public key armor by uid.
func (csk ConcurrentSafeKeyring) ExportPubKeyArmor(uid string) (string, error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.ExportPubKeyArmor(uid)
}

// ExportPubKeyArmorByAddress exports public key armor by address.
func (csk ConcurrentSafeKeyring) ExportPubKeyArmorByAddress(address sdk.Address) (string, error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.ExportPubKeyArmorByAddress(address)
}

// ExportPrivKeyArmor exports priv key armor by uid.
func (csk ConcurrentSafeKeyring) ExportPrivKeyArmor(uid, encryptPassphrase string) (armor string, err error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.ExportPrivKeyArmor(uid, encryptPassphrase)
}

// ExportPrivKeyArmorByAddress exports priv key armor by address.
func (csk ConcurrentSafeKeyring) ExportPrivKeyArmorByAddress(address sdk.Address, encryptPassphrase string) (armor string, err error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.ExportPrivKeyArmorByAddress(address, encryptPassphrase)
}

// Sign signs byte messages with a user key.
func (csk ConcurrentSafeKeyring) Sign(uid string, msg []byte) ([]byte, types.PubKey, error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.Sign(uid, msg)
}

// SignByAddress sign byte messages with a user key providing the address.
func (csk ConcurrentSafeKeyring) SignByAddress(address sdk.Address, msg []byte) ([]byte, types.PubKey, error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.SignByAddress(address, msg)
}

// Write operations:

// Delete deletes keys from the keyring by uid.
func (csk ConcurrentSafeKeyring) Delete(uid string) error {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.Delete(uid)
}

// DeleteByAddress deletes keys from the keyring by address.
func (csk ConcurrentSafeKeyring) DeleteByAddress(address sdk.Address) error {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.DeleteByAddress(address)
}

// NewMnemonic generates a new mnemonic, derives a hierarchical deterministic key from it, and
// persists the key to storage. Returns the generated mnemonic and the key Info.
// It returns an error if it fails to generate a key for the given algo type, or if
// another key is already stored under the same name or address.
//
// A passphrase set to the empty string will set the passphrase to the DefaultBIP39Passphrase value.
func (csk ConcurrentSafeKeyring) NewMnemonic(uid string, language keyring.Language, hdPath, bip39Passphrase string, algo keyring.SignatureAlgo) (*keyring.Record, string, error) {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.NewMnemonic(uid, language, hdPath, bip39Passphrase, algo)
}

// NewAccount converts a mnemonic to a private key and BIP-39 HD Path and persists it.
// It fails if there is an existing key Info with the same address.
func (csk ConcurrentSafeKeyring) NewAccount(uid, mnemonic, bip39Passphrase, hdPath string, algo keyring.SignatureAlgo) (*keyring.Record, error) {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.NewAccount(uid, mnemonic, bip39Passphrase, hdPath, algo)
}

// SaveLedgerKey retrieves a public key reference from a Ledger device and persists it.
func (csk ConcurrentSafeKeyring) SaveLedgerKey(uid string, algo keyring.SignatureAlgo, hrp string, coinType, account, index uint32) (*keyring.Record, error) {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.SaveLedgerKey(uid, algo, hrp, coinType, account, index)
}

// SaveMultisig stores and returns a new multsig (offline) key reference.
func (csk ConcurrentSafeKeyring) SaveMultisig(uid string, pubkey types.PubKey) (*keyring.Record, error) {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.SaveMultisig(uid, pubkey)
}

// ImportPrivKey imports ASCII armored passphrase-encrypted private keys.
func (csk ConcurrentSafeKeyring) ImportPrivKey(uid, armor, passphrase string) error {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.ImportPrivKey(uid, armor, passphrase)
}

// ImportPrivKeyHex imports hex encoded keys.
func (csk ConcurrentSafeKeyring) ImportPrivKeyHex(uid, privKey, algoStr string) error {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.ImportPrivKeyHex(uid, privKey, algoStr)
}

// ImportPubKey imports ASCII armored public keys.
func (csk ConcurrentSafeKeyring) ImportPubKey(uid, armor string) error {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.ImportPubKey(uid, armor)
}

// Backend returns the backend type used in the keyring config: "file", "os", "kwallet", "pass", "test", "memory".
func (csk ConcurrentSafeKeyring) Backend() string {
	return csk.kr.Backend()
}

// Rename renames an existing key from the Keyring.
func (csk ConcurrentSafeKeyring) Rename(from, to string) error {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.Rename(from, to)
}

// SaveOfflineKey stores a public key and returns the persisted Info structure.
func (csk ConcurrentSafeKeyring) SaveOfflineKey(uid string, pubkey types.PubKey) (*keyring.Record, error) {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.SaveOfflineKey(uid, pubkey)
}

// MigrateAll migrates keys from amino to proto.
func (csk ConcurrentSafeKeyring) MigrateAll() ([]*keyring.Record, error) {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.MigrateAll()
}
