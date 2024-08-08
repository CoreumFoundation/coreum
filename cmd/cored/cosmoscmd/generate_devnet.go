package cosmoscmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	sdkmath "cosmossdk.io/math"
	tmos "github.com/cometbft/cometbft/libs/os"
	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/go-bip39"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/v4/app"
	"github.com/CoreumFoundation/coreum/v4/pkg/config"
	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
)

const (
	// FlagOutputPath defines an output path.
	FlagOutputPath = "output-path"

	// FlagInputPath defines an input path.
	FlagInputPath = "input-path"

	// FlagValidatorName defines a name of the validator.
	FlagValidatorName = "validator-name"

	// mnemonicEntropySize used by cosmos SDK.
	mnemonicEntropySize = 256
)

// GenerateDevnetCmd returns a command that generates devnet files needed to start the devnet.
//
//nolint:funlen // breaking down this function will make it less maintainable.
func GenerateDevnetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-devnet",
		Short: "Generate devnet configuration files",
		Long:  `Generate devnet validators' and nodes' configuration files and genesis,`,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			network := app.ChosenNetwork
			if app.ChosenNetwork.ChainID() != constant.ChainIDDev {
				return errors.Errorf("the command supports the %s chain id only", constant.ChainIDDev)
			}

			validatorNames, err := cmd.Flags().GetStringArray(FlagValidatorName)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to read %s flag", FlagValidatorName))
			}
			if len(validatorNames) == 0 {
				return errors.Wrap(err, "at least one validator name must be provided")
			}
			duplicatedNames := lo.FindDuplicates(validatorNames)
			if len(duplicatedNames) != 0 {
				return errors.Wrap(
					err,
					fmt.Sprintf("it is prohibited to use duplicatd validator names, duplicates: %v", duplicatedNames),
				)
			}

			outputPath, err := cmd.Flags().GetString(FlagOutputPath)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to read %s flag", FlagOutputPath))
			}
			outputPath, err = filepath.Abs(outputPath)
			if err != nil {
				return errors.Wrap(err, "failed to get absolute path")
			}
			fmt.Printf("Generating devnet config to ouptut path: %s\n", outputPath)

			cfg := server.GetServerContextFromCmd(cmd).Config
			for _, validatorName := range validatorNames {
				validatorOutputPath := filepath.Join(outputPath, validatorName, string(network.ChainID()))
				cfg.SetRoot(validatorOutputPath)
				configDir := filepath.Join(validatorOutputPath, nodeConfigDirName)
				if err := os.MkdirAll(configDir, 0o700); err != nil {
					return errors.Wrap(err, "failed to make config directory")
				}

				network.NodeConfig.Name = validatorName
				cfg = network.NodeConfig.TendermintNodeConfig(cfg)
				nodeID, validatorPubKey, err := genutil.InitializeNodeValidatorFilesFromMnemonic(cfg, "")
				if err != nil {
					return errors.Wrap(err, "failed to init validator home")
				}
				if err := config.WriteTendermintConfigToFile(
					filepath.Join(validatorOutputPath, config.DefaultNodeConfigPath),
					cfg,
				); err != nil {
					return errors.Wrap(err, "failed to write tendermint config to file")
				}

				// prepare data for the genesis generation
				mnemonic, err := generateMnemonic()
				if err != nil {
					return errors.Wrap(err, "failed to generate mnemonic")
				}
				if network, err = addValidatorToNetwork(
					ctx,
					cosmosclient.GetClientContextFromCmd(cmd),
					network,
					validatorName,
					validatorPubKey,
					mnemonic,
				); err != nil {
					return err
				}
				fmt.Printf("\nGenerated validator `%s`:\nMnemonic: %s\nNode ID: %s\n", validatorName, mnemonic, nodeID)
			}

			genDocBytes, err := network.EncodeGenesis()
			if err != nil {
				return err
			}

			// write same genesis to all validators
			for _, validatorName := range validatorNames {
				validatorOutputPath := filepath.Join(outputPath, validatorName, string(network.ChainID()))
				genPath := filepath.Join(validatorOutputPath, nodeConfigDirName, genesisFileName)
				if tmos.FileExists(genPath) {
					return errors.Errorf("genesis already exists: %s", genPath)
				}
				if err := os.WriteFile(genPath, genDocBytes, 0644); err != nil {
					return errors.Wrap(err, "failed to write genesis bytes to file")
				}
			}
			fmt.Printf("\nAll validators are successfully generated.\n")

			return nil
		},
	}

	cmd.Flags().String(FlagOutputPath, "", "output directory for the generated files")
	cmd.Flags().StringArray(FlagValidatorName, []string{}, "list of the validator names to generate")

	return cmd
}

func generateMnemonic() (string, error) {
	entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
	if err != nil {
		return "", err
	}

	return bip39.NewMnemonic(entropySeed)
}

func addValidatorToNetwork(
	ctx context.Context,
	clientCtx cosmosclient.Context,
	network config.NetworkConfig,
	validatorName string,
	validatorPubKey cryptotypes.PubKey,
	mnemonic string,
) (config.NetworkConfig, error) {
	networkProvider, ok := network.Provider.(config.DynamicConfigProvider)
	if !ok {
		return config.NetworkConfig{}, errors.New("failed to cast network.Provider to  config.DynamicConfigProvider")
	}

	const signerKeyName = "signer"
	clientCtx = clientCtx.WithFrom(signerKeyName)
	inMemKeyring := keyring.NewInMemory(config.NewEncodingConfig(app.ModuleBasics).Codec)
	k, err := inMemKeyring.NewAccount(
		signerKeyName,
		mnemonic,
		"",
		hd.CreateHDPath(sdk.GetConfig().GetCoinType(), 0, 0).String(),
		hd.Secp256k1,
	)
	if err != nil {
		return config.NetworkConfig{}, errors.Wrap(err, "failed to import account with mnemonic")
	}

	stakerAddress, err := k.GetAddress()
	if err != nil {
		return config.NetworkConfig{}, errors.Wrap(err, "failed get staker address from key")
	}

	// 10m delegated and 1m extra to the txs
	networkProvider = networkProvider.WithAccount(
		stakerAddress,
		sdk.NewCoins(sdk.NewCoin(constant.DenomDev, sdkmath.NewInt(11_000_000_000_000))),
	)
	stakerSelfDelegationAmount := sdk.NewCoin(constant.DenomDev, sdkmath.NewInt(10_000_000_000_000))
	commission := stakingtypes.CommissionRates{
		Rate:          sdkmath.LegacyMustNewDecFromStr("0.1"),
		MaxRate:       sdkmath.LegacyMustNewDecFromStr("0.2"),
		MaxChangeRate: sdkmath.LegacyMustNewDecFromStr("0.01"),
	}

	msg, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(stakerAddress).String(),
		validatorPubKey,
		stakerSelfDelegationAmount,
		stakingtypes.Description{
			Moniker: validatorName,
		},
		commission,
		stakerSelfDelegationAmount.Amount,
	)
	if err != nil {
		return config.NetworkConfig{}, errors.Wrap(err, "failed create MsgCreateValidator transaction")
	}

	txf := tx.Factory{}.
		WithChainID(string(network.ChainID())).
		WithKeybase(inMemKeyring).
		WithTxConfig(clientCtx.TxConfig)
	txBuilder, err := txf.BuildUnsignedTx(msg)
	if err != nil {
		return config.NetworkConfig{}, errors.Wrap(err, "failed to build MsgCreateValidator transaction")
	}
	if err := tx.Sign(ctx, txf, signerKeyName, txBuilder, true); err != nil {
		return config.NetworkConfig{}, errors.Wrap(err, "failed to sign MsgCreateValidator transaction")
	}
	txBytes, err := clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
	if err != nil {
		return config.NetworkConfig{}, errors.Wrap(err, "failed to encode MsgCreateValidator transaction")
	}
	networkProvider = networkProvider.WithGenesisTx(txBytes)
	// update provider
	network.Provider = networkProvider

	return network, nil
}
