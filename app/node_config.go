package app

import (
	"crypto/ed25519"
	"net"
	"os"
	"strconv"

	"github.com/tendermint/tendermint/config"
	tmed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/p2p"
	"github.com/tendermint/tendermint/privval"
)

// NodeConfig saves files with private keys and config required by node
type NodeConfig struct {
	Name           string
	PrometheusPort int
	NodeKey        ed25519.PrivateKey
	ValidatorKey   ed25519.PrivateKey
}

// Save saves files required by validator
func (nc NodeConfig) Save(homeDir string) error {
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

	cfg := config.DefaultConfig()
	cfg.Moniker = nc.Name
	// set addr_book_strict to false so nodes connecting from non-routable hosts are added to address book
	cfg.P2P.AddrBookStrict = false
	cfg.P2P.AllowDuplicateIP = true
	cfg.P2P.MaxNumOutboundPeers = 100
	cfg.P2P.MaxNumInboundPeers = 100
	cfg.RPC.MaxSubscriptionClients = 10000
	cfg.RPC.MaxOpenConnections = 10000
	cfg.RPC.GRPCMaxOpenConnections = 10000
	cfg.RPC.MaxSubscriptionsPerClient = 10000
	cfg.Mempool.Size = 50000
	cfg.Mempool.MaxTxsBytes = 5368709120
	cfg.Instrumentation.Prometheus = true
	cfg.Instrumentation.PrometheusListenAddr = net.JoinHostPort(net.IPv4zero.String(), strconv.Itoa(nc.PrometheusPort))
	config.WriteConfigFile(homeDir+"/config/config.toml", cfg)
	return nil
}
