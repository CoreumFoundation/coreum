package cosmoscmd

import (
	"bufio"
	"os"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/tendermint/tendermint/libs/cli"
	tmos "github.com/tendermint/tendermint/libs/os"

	"github.com/CoreumFoundation/coreum/v2/app"
	"github.com/CoreumFoundation/coreum/v2/pkg/config"
)

// Used flags.
const (
	// FlagOverwrite defines a flag to overwrite an existing genesis JSON file.
	FlagOverwrite = "overwrite"

	// FlagRecover defines a flag which determines whether to init private keys from mnemonic or generate new ones.
	FlagRecover = "recover"
)

// InitCmd returns the init cobra command.
func InitCmd(network config.NetworkConfig, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize configuration files for private validator, p2p, genesis, and application",
		Long:  `Initialize validators's and node's configuration files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			cfg := server.GetServerContextFromCmd(cmd).Config
			cfg.SetRoot(clientCtx.HomeDir)

			// Get bip39 mnemonic
			var mnemonic string
			isRecover, err := cmd.Flags().GetBool(FlagRecover)
			if err != nil {
				return errors.Wrapf(err, "got error parsing recover flag")
			}
			if isRecover {
				inBuf := bufio.NewReader(cmd.InOrStdin())
				value, err := input.GetString("Enter your bip39 mnemonic", inBuf)
				if err != nil {
					return err
				}

				mnemonic = value
				if !bip39.IsMnemonicValid(mnemonic) {
					return errors.New("invalid mnemonic")
				}
			}

			genFile := cfg.GenesisFile()
			overwrite, _ := cmd.Flags().GetBool(FlagOverwrite)

			if !overwrite && tmos.FileExists(genFile) {
				return errors.Errorf("genesis.json file already exists: %v", genFile)
			}

			genDocBytes, err := network.EncodeGenesis()
			if err != nil {
				return err
			}

			configDir := filepath.Join(clientCtx.HomeDir, "config")
			if err := os.MkdirAll(configDir, 0o700); err != nil {
				return errors.Wrap(err, "unable to make config directory")
			}

			if err := os.WriteFile(filepath.Join(configDir, "genesis.json"), genDocBytes, 0644); err != nil {
				return errors.Wrap(err, "unable to write genesis bytes to file")
			}

			network.NodeConfig.Name = args[0]
			cfg = network.NodeConfig.TendermintNodeConfig(cfg)

			_, _, err = genutil.InitializeNodeValidatorFilesFromMnemonic(cfg, mnemonic)
			if err != nil {
				return err
			}

			return config.WriteTendermintConfigToFile(
				filepath.Join(cfg.RootDir, config.DefaultNodeConfigPath),
				cfg,
			)
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flags.FlagChainID, string(app.DefaultChainID), "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().BoolP(FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().Bool(FlagRecover, false, "provide seed phrase to recover existing key instead of creating")

	return cmd
}
