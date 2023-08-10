package config

import (
	"crypto/ed25519"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tmcfg "github.com/cometbft/cometbft/config"
	tmed25519 "github.com/cometbft/cometbft/crypto/ed25519"
	"github.com/cometbft/cometbft/p2p"
	"github.com/cometbft/cometbft/privval"
)

// DefaultNodeConfigPath is the default path there the config.toml is saved.
var DefaultNodeConfigPath = filepath.Join("config", "config.toml")

// NodeConfig saves files with private keys and config required by node.
type NodeConfig struct {
	Name           string
	PrometheusPort int
	NodeKey        ed25519.PrivateKey
	ValidatorKey   ed25519.PrivateKey
	SeedPeers      []string
}

// SavePrivateKeys saves private keys to files.
func (nc NodeConfig) SavePrivateKeys(homeDir string) error {
	err := os.MkdirAll(homeDir+"/config", 0o700)
	if err != nil {
		return err
	}

	err = (&p2p.NodeKey{
		PrivKey: tmed25519.PrivKey(nc.NodeKey),
	}).SaveAs(homeDir + "/config/node_key.json")
	if err != nil {
		return err
	}

	if nc.ValidatorKey != nil {
		err = os.MkdirAll(homeDir+"/data", 0o700)
		if err != nil {
			return err
		}

		privval.
			NewFilePV(
				tmed25519.PrivKey(nc.ValidatorKey),
				homeDir+"/config/priv_validator_key.json",
				homeDir+"/data/priv_validator_state.json").
			Save()
	}
	return nil
}

// TendermintNodeConfig applies node's tendermint config.
func (nc NodeConfig) TendermintNodeConfig(cfg *tmcfg.Config) *tmcfg.Config {
	if cfg == nil {
		cfg = tmcfg.DefaultConfig()
	}

	if nc.Name != "" {
		cfg.Moniker = nc.Name
	}

	if nc.PrometheusPort > 0 {
		cfg.Instrumentation.Prometheus = true
		cfg.Instrumentation.PrometheusListenAddr = net.JoinHostPort(net.IPv4zero.String(), strconv.Itoa(nc.PrometheusPort))
	}

	if len(nc.SeedPeers) > 0 {
		cfg.P2P.Seeds = strings.Join(nc.SeedPeers, ",")
	}

	// Update the default consensus config
	cfg.Consensus.TimeoutCommit = time.Second

	return cfg
}

func (nc NodeConfig) clone() NodeConfig {
	copied := NodeConfig{
		Name:           nc.Name,
		PrometheusPort: nc.PrometheusPort,
		NodeKey:        make([]byte, len(nc.NodeKey)),
		ValidatorKey:   make([]byte, len(nc.ValidatorKey)),
		SeedPeers:      make([]string, len(nc.SeedPeers)),
	}

	copy(copied.NodeKey, nc.NodeKey)
	copy(copied.ValidatorKey, nc.ValidatorKey)
	copy(copied.SeedPeers, nc.SeedPeers)

	return copied
}

// WriteTendermintConfigToFile saves tendermint config to file.
func WriteTendermintConfigToFile(filePath string, cfg *tmcfg.Config) error {
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmcfg.WriteConfigFile(filePath, cfg)
	return nil
}
