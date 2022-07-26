package main

import (
	"bufio"
	"path/filepath"

	"github.com/CoreumFoundation/coreum/app"
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
)

// used flags
const (
	// FlagOverwrite defines a flag to overwrite an existing genesis JSON file.
	FlagOverwrite = "overwrite"

	// FlagRecover defines a flag which determines whether to init private keys from mnemonic or generate new ones.
	FlagRecover = "recover"
)

func initCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [moniker]",
		Short: "Initialize configuration files for private validator, p2p, genesis, and application",
		Long:  `Initialize validators's and node's configuration files.`,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			chainID, _ := cmd.Flags().GetString(flags.FlagChainID)
			if chainID == "" {
				return errors.New("chain id not provided")
			}

			clientCtx := client.GetClientContextFromCmd(cmd)

			config := server.GetServerContextFromCmd(cmd).Config
			config.SetRoot(clientCtx.HomeDir)

			// Get bip39 mnemonic
			var mnemonic string
			isRecover, _ := cmd.Flags().GetBool(FlagRecover)
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

			genFile := config.GenesisFile()
			overwrite, _ := cmd.Flags().GetBool(FlagOverwrite)

			if !overwrite && tmos.FileExists(genFile) {
				return errors.Errorf("genesis.json file already exists: %v", genFile)
			}

			network, err := app.NetworkByChainID(app.ChainID(chainID))
			if err != nil {
				return err
			}

			err = network.SaveGenesis(clientCtx.HomeDir)
			if err != nil {
				return err
			}

			networkNodeConfig := network.NodeConfig()
			networkNodeConfig.Name = args[0]
			config = network.NodeConfig().TendermintNodeConfig(config)

			_, _, err = genutil.InitializeNodeValidatorFilesFromMnemonic(config, mnemonic)
			if err != nil {
				return err
			}

			app.WriteTendermintConfigToFile(
				filepath.Join(config.RootDir, app.DefaultNodeConfigPath),
				config,
			)

			return nil
		},
	}

	cmd.Flags().String(cli.HomeFlag, defaultNodeHome, "node's home directory")
	cmd.Flags().String(flags.FlagChainID, string(app.DefaultChainID), "genesis file chain-id, if left blank will be randomly created")
	cmd.Flags().BoolP(FlagOverwrite, "o", false, "overwrite the genesis.json file")
	cmd.Flags().Bool(FlagRecover, false, "provide seed phrase to recover existing key instead of creating")

	return cmd
}
