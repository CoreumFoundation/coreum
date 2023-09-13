package cosmoscmd

// The command init.go copied from https://github.com/cosmos/cosmos-sdk/blob/v0.47.4/x/genutil/client/cli/init.go.

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/cometbft/cometbft/libs/cli"
	tmos "github.com/cometbft/cometbft/libs/os"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/input"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/v3/app"
	"github.com/CoreumFoundation/coreum/v3/pkg/config"
)

//nolint:tagliatelle,tagalign // default structure
type printInfo struct {
	Moniker    string          `json:"moniker" yaml:"moniker"`
	ChainID    string          `json:"chain_id" yaml:"chain_id"`
	NodeID     string          `json:"node_id" yaml:"node_id"`
	GenTxsDir  string          `json:"gentxs_dir" yaml:"gentxs_dir"`
	AppMessage json.RawMessage `json:"app_message" yaml:"app_message"`
}

func newPrintInfo(moniker, chainID, nodeID, genTxsDir string, appMessage json.RawMessage) printInfo {
	return printInfo{
		Moniker:    moniker,
		ChainID:    chainID,
		NodeID:     nodeID,
		GenTxsDir:  genTxsDir,
		AppMessage: appMessage,
	}
}

func displayInfo(info printInfo) error {
	out, err := json.MarshalIndent(info, "", " ")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(os.Stderr, "%s\n", sdk.MustSortJSON(out))

	return err
}

// InitCmd returns a command that initializes all files needed for Tendermint
// and the respective application.
func InitCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize private validator, p2p, genesis, and application configuration files",
		Long:  `Initialize validators's and node's configuration files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			network := app.ChosenNetwork

			clientCtx := client.GetClientContextFromCmd(cmd)

			cfg := server.GetServerContextFromCmd(cmd).Config
			cfg.SetRoot(clientCtx.HomeDir)

			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			switch {
			case chainID != "":
			case clientCtx.ChainID != "":
				chainID = clientCtx.ChainID
			default:
				return errors.Errorf("undefined chain ID %s", chainID)
			}

			// Get bip39 mnemonic
			var mnemonic string
			isRecover, err := cmd.Flags().GetBool(genutilcli.FlagRecover)
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
			overwrite, _ := cmd.Flags().GetBool(genutilcli.FlagOverwrite)

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

			nodeID, _, err := genutil.InitializeNodeValidatorFilesFromMnemonic(cfg, mnemonic)
			if err != nil {
				return err
			}

			if err := config.WriteTendermintConfigToFile(
				filepath.Join(cfg.RootDir, config.DefaultNodeConfigPath),
				cfg,
			); err != nil {
				return err
			}

			return displayInfo(newPrintInfo(cfg.Moniker, chainID, nodeID, "", genDocBytes))
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().BoolP(genutilcli.FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().Bool(genutilcli.FlagRecover, false, "provide seed phrase to recover existing key instead of creating")
	cmd.Flags().String(flags.FlagChainID, "", "genesis file chain-id, if left blank will be randomly created")

	return cmd
}
