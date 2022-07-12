// COPIED FROM https://github.com/ignite/cli/tree/e6a5efdaa2210fb72e33382d442268cdd466ae2d/ignite/pkg/cosmoscmd
// UNDER APACHE2.0 LICENSE
package cosmoscmd

import (
	"path/filepath"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/spf13/cobra"

	"github.com/ignite/cli/ignite/pkg/cosmosutil"
	"github.com/ignite/cli/ignite/pkg/ctxticker"
	"github.com/ignite/cli/ignite/pkg/gitpod"
	"github.com/ignite/cli/ignite/pkg/xchisel"
	"github.com/ignite/cli/ignite/services/network/networkchain"
)

const TunnelRerunDelay = 5 * time.Second

// startProxyForTunneledPeers hooks the `appd start` command to start an HTTP proxy server and HTTP proxy clients
// for each node that needs HTTP tunneling.
// HTTP tunneling is activated ** ONLY** if your app's `$APP_HOME/config` dir has an `spn.yml` file
// and only if this file has `tunneled_peers` field inside with a list of tunneled peers/nodes.
//
// If you're using SPN as coordinator and do not want to allow HTTP tunneling feature at all,
// you can prevent `spn.yml` file to being generated by not approving validator requests
// that has HTTP tunneling enabled instead of plain TCP connections.
func startProxyForTunneledPeers(clientCtx client.Context, cmd *cobra.Command) {
	if cmd.Name() != "start" {
		return
	}
	serverCtx := server.GetServerContextFromCmd(cmd)
	ctx := cmd.Context()

	spnConfigPath := filepath.Join(clientCtx.HomeDir, cosmosutil.ChainConfigDir, networkchain.SPNConfigFile)
	spnConfig, err := networkchain.GetSPNConfig(spnConfigPath)
	if err != nil {
		serverCtx.Logger.Error("Failed to open spn config file", "reason", err.Error())
		return
	}
	// exit if there aren't tunneled validators in the network
	if len(spnConfig.TunneledPeers) == 0 {
		return
	}

	for _, peer := range spnConfig.TunneledPeers {
		if peer.Name == networkchain.HTTPTunnelChisel {
			peer := peer
			go func() {
				ctxticker.DoNow(ctx, TunnelRerunDelay, func() error {
					serverCtx.Logger.Info("Starting chisel client", "tunnelAddress", peer.Address, "localPort", peer.LocalPort)
					err := xchisel.StartClient(ctx, peer.Address, peer.LocalPort, "26656")
					if err != nil {
						serverCtx.Logger.Error("Failed to start chisel client",
							"tunnelAddress", peer.Address,
							"localPort", peer.LocalPort,
							"reason", err.Error(),
						)
					}
					return nil
				})
			}()
		}
	}

	if gitpod.IsOnGitpod() {
		go func() {
			ctxticker.DoNow(ctx, TunnelRerunDelay, func() error {
				serverCtx.Logger.Info("Starting chisel server", "port", xchisel.DefaultServerPort)
				err := xchisel.StartServer(ctx, xchisel.DefaultServerPort)
				if err != nil {
					serverCtx.Logger.Error(
						"Failed to start chisel server",
						"port", xchisel.DefaultServerPort,
						"reason", err.Error(),
					)
				}
				return nil
			})
		}()
	}
}
