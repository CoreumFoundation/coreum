package integrationtests

import (
	"sync"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// concurrentSafeKeyring wraps standard Cosmos SDK keyring implementation to make it concurrent safe.
// Since we run our integration tests in parallel from time to time we get: "concurrent map read and map write"
// so concurrentSafeKeyring wraps all methods of keyring and locks mutex before calling method.
type concurrentSafeKeyring struct {
	mu *sync.RWMutex
	kr keyring.Keyring
}

func newConcurrentSafeKeyring(kr keyring.Keyring) concurrentSafeKeyring {
	return concurrentSafeKeyring{
		mu: &sync.RWMutex{},
		kr: kr,
	}
}

func (csk concurrentSafeKeyring) SupportedAlgorithms() (keyring.SigningAlgoList, keyring.SigningAlgoList) {
	return csk.kr.SupportedAlgorithms()
}

// Read operations:

func (csk concurrentSafeKeyring) List() ([]keyring.Info, error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.List()
}

func (csk concurrentSafeKeyring) Key(uid string) (keyring.Info, error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.Key(uid)
}

func (csk concurrentSafeKeyring) KeyByAddress(address sdk.Address) (keyring.Info, error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.KeyByAddress(address)
}

func (csk concurrentSafeKeyring) ExportPubKeyArmor(uid string) (string, error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.ExportPubKeyArmor(uid)
}

func (csk concurrentSafeKeyring) ExportPubKeyArmorByAddress(address sdk.Address) (string, error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.ExportPubKeyArmorByAddress(address)
}

func (csk concurrentSafeKeyring) ExportPrivKeyArmor(uid, encryptPassphrase string) (armor string, err error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.ExportPrivKeyArmor(uid, encryptPassphrase)
}

func (csk concurrentSafeKeyring) ExportPrivKeyArmorByAddress(address sdk.Address, encryptPassphrase string) (armor string, err error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.ExportPrivKeyArmorByAddress(address, encryptPassphrase)
}

func (csk concurrentSafeKeyring) Sign(uid string, msg []byte) ([]byte, types.PubKey, error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.Sign(uid, msg)
}

func (csk concurrentSafeKeyring) SignByAddress(address sdk.Address, msg []byte) ([]byte, types.PubKey, error) {
	csk.mu.RLock()
	defer csk.mu.RUnlock()

	return csk.kr.SignByAddress(address, msg)
}

// Write operations:

func (csk concurrentSafeKeyring) Delete(uid string) error {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.Delete(uid)
}

func (csk concurrentSafeKeyring) DeleteByAddress(address sdk.Address) error {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.DeleteByAddress(address)
}

func (csk concurrentSafeKeyring) NewMnemonic(uid string, language keyring.Language, hdPath, bip39Passphrase string, algo keyring.SignatureAlgo) (keyring.Info, string, error) {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.NewMnemonic(uid, language, hdPath, bip39Passphrase, algo)
}

func (csk concurrentSafeKeyring) NewAccount(uid, mnemonic, bip39Passphrase, hdPath string, algo keyring.SignatureAlgo) (keyring.Info, error) {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.NewAccount(uid, mnemonic, bip39Passphrase, hdPath, algo)
}

func (csk concurrentSafeKeyring) SaveLedgerKey(uid string, algo keyring.SignatureAlgo, hrp string, coinType, account, index uint32) (keyring.Info, error) {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.SaveLedgerKey(uid, algo, hrp, coinType, account, index)
}

func (csk concurrentSafeKeyring) SavePubKey(uid string, pubkey types.PubKey, algo hd.PubKeyType) (keyring.Info, error) {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.SavePubKey(uid, pubkey, algo)
}

func (csk concurrentSafeKeyring) SaveMultisig(uid string, pubkey types.PubKey) (keyring.Info, error) {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.SaveMultisig(uid, pubkey)
}

func (csk concurrentSafeKeyring) ImportPrivKey(uid, armor, passphrase string) error {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.ImportPrivKey(uid, armor, passphrase)
}

func (csk concurrentSafeKeyring) ImportPubKey(uid, armor string) error {
	csk.mu.Lock()
	defer csk.mu.Unlock()

	return csk.kr.ImportPubKey(uid, armor)
}
