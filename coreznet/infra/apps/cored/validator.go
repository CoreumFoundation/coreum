package cored

import (
	"crypto/ed25519"
	"encoding/hex"
	"os"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/tendermint/tendermint/config"
	tmed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
)

// SaveIdentityFiles saves files with private keys required by validator
func SaveIdentityFiles(homeDir string, nodePrivateKey ed25519.PrivateKey, validatorPrivateKey ed25519.PrivateKey) {
	must.OK(os.MkdirAll(homeDir+"/config", 0o700))
	must.OK(os.MkdirAll(homeDir+"/data", 0o700))

	must.OK((&p2p.NodeKey{
		PrivKey: tmed25519.PrivKey(nodePrivateKey),
	}).SaveAs(homeDir + "/config/node_key.json"))

	privval.NewFilePV(tmed25519.PrivKey(validatorPrivateKey), homeDir+"/config/priv_validator_key.json", homeDir+"/data/priv_validator_state.json").Save()

	cfg := config.DefaultConfig()
	// set addr_book_strict to false so nodes connecting from non-routable hosts are added to address book
	cfg.P2P.AddrBookStrict = false
	cfg.P2P.AllowDuplicateIP = true
	config.WriteConfigFile(homeDir+"/config/config.toml", cfg)
}

// NodeID computes node ID from node public key
func NodeID(pubKey ed25519.PublicKey) string {
	return hex.EncodeToString(tmed25519.PubKey(pubKey).Address())
}
