// Package cosmoscmd contains the root of the commands.
// The commands root.go copied from https://github.com/cosmos/cosmos-sdk/blob/v0.47.4/simapp/simd/cmd/root.go.
// under APACHE2.0 LICENSE
package cosmoscmd

import (
	"context"
	"io"
	"os"
	"time"

	"cosmossdk.io/log"
	confixcmd "cosmossdk.io/tools/confix/cmd"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	tmcfg "github.com/cometbft/cometbft/config"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	clientconfig "github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/server"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	tx "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	rosettaCmd "github.com/cosmos/rosetta/cmd"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/CoreumFoundation/coreum/v6/app"
	coreumclient "github.com/CoreumFoundation/coreum/v6/pkg/client"
	"github.com/CoreumFoundation/coreum/v6/pkg/config"
)

const ledgerAppName = "Coreum"

// NewRootCmd creates a new root command. It is called once in the
// main function.
func NewRootCmd() *cobra.Command {
	// we "pre"-instantiate the application for getting the injected/configured encoding configuration
	tempApp := app.New(log.NewNopLogger(), dbm.NewMemDB(), nil, true, simtestutil.NewAppOptionsWithFlagHome(tempDir()))
	encodingConfig := config.EncodingConfig{
		InterfaceRegistry: tempApp.InterfaceRegistry(),
		Codec:             tempApp.AppCodec(),
		TxConfig:          tempApp.TxConfig(),
		Amino:             tempApp.LegacyAmino(),
	}
	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome).
		WithViper("").
		WithKeyringOptions(func(options *keyring.Options) {
			options.LedgerAppName = ledgerAppName
		})

	rootCmd := &cobra.Command{
		Use:   app.Name + "d",
		Short: "Coreum App",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())
			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}
			initClientCtx, err = clientconfig.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}
			if !initClientCtx.Offline {
				enabledSignModes := make([]signingtypes.SignMode, 0)
				enabledSignModes = append(enabledSignModes, authtx.DefaultSignModes...)
				enabledSignModes = append(enabledSignModes, signingtypes.SignMode_SIGN_MODE_TEXTUAL)
				txConfigOpts := authtx.ConfigOptions{
					EnabledSignModes:           enabledSignModes,
					TextualCoinMetadataQueryFn: tx.NewGRPCCoinMetadataQueryFn(initClientCtx),
				}
				txConfig, err := authtx.NewTxConfigWithOptions(
					encodingConfig.Codec,
					txConfigOpts,
				)
				if err != nil {
					return err
				}

				initClientCtx = initClientCtx.WithTxConfig(txConfig)
			}
			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			initClientCtx = initClientCtx.WithCmdContext(cmd.Context())
			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			customTMConfig := app.ChosenNetwork.NodeConfig.TendermintNodeConfig(initTendermintConfig())

			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customTMConfig)
		},
	}

	initRootCmd(rootCmd, encodingConfig, tempApp.BasicModuleManager)
	// add keyring to autocli opts
	autoCliOpts := tempApp.AutoCliOpts()
	initClientCtx, _ = clientconfig.ReadDefaultValuesFromDefaultClientConfig(initClientCtx)
	autoCliOpts.ClientCtx = initClientCtx

	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	for _, cmd := range rootCmd.Commands() {
		if cmd.Use == "tx" {
			installAwaitBroadcastModeWrapper(cmd)
			addQueryGasPriceToAllLeafs(cmd)
			break
		}
	}

	return rootCmd
}

// initTendermintConfig helps to override default Tendermint Config values.
// return tmcfg.DefaultConfig if no custom configuration is required for the application.
func initTendermintConfig() *tmcfg.Config {
	cfg := tmcfg.DefaultConfig()

	// these values put a higher strain on node memory
	// cfg.P2P.MaxNumInboundPeers = 100
	// cfg.P2P.MaxNumOutboundPeers = 40

	return cfg
}

// initAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func initAppConfig() (string, interface{}) {
	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := serverconfig.DefaultConfig()
	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	//
	// In app, we set the min gas prices to 0.
	srvCfg.MinGasPrices = "0.00000000000000001" + app.ChosenNetwork.Denom()

	// WASMConfig defines configuration for the wasm module.
	type WASMConfig struct {
		// # This is the maximum sdk gas (wasm and storage) that we allow for any x/wasm "smart" queries
		QueryGasLimit uint64
		// This defines the memory size for Wasm modules that we can keep cached to speed-up instantiation
		// The value is in MiB not bytes
		MemoryCacheSize uint32
	}

	type CustomAppConfig struct {
		serverconfig.Config
		WASM WASMConfig
	}

	defaultWasmNodeConfig := wasmtypes.DefaultNodeConfig()
	customAppConfig := CustomAppConfig{
		Config: *srvCfg,
		WASM: WASMConfig{
			QueryGasLimit:   defaultWasmNodeConfig.SmartQueryGasLimit,
			MemoryCacheSize: defaultWasmNodeConfig.MemoryCacheSize,
		},
	}

	customAppTemplate := serverconfig.DefaultConfigTemplate + `
[wasm]
# This is the maximum sdk gas (wasm and storage) that we allow for any x/wasm "smart" queries
query_gas_limit = {{ .WASM.QueryGasLimit }}
# This defines the memory size for Wasm modules that we can keep cached to speed-up instantiation
# The value is in MiB not bytes
memory_cache_size = {{ .WASM.MemoryCacheSize }}
`

	return customAppTemplate, customAppConfig
}

func initRootCmd(
	rootCmd *cobra.Command,
	encodingConfig config.EncodingConfig,
	basicManager module.BasicManager,
) {
	cfg := sdk.GetConfig()
	cfg.Seal()

	rootCmd.AddCommand(
		InitCmd(basicManager, app.DefaultNodeHome),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		pruning.Cmd(newApp, app.DefaultNodeHome),
		snapshot.Cmd(newApp),
		GenerateGenesisCmd(basicManager),
	)

	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp, appExport, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, genesis, and tx child commands
	rootCmd.AddCommand(
		server.StatusCommand(),
		genesisCommand(encodingConfig.TxConfig, basicManager),
		queryCommand(),
		txCommand(),
		keys.Commands(),
	)

	// add rosetta
	rootCmd.AddCommand(rosettaCmd.RosettaCommand(encodingConfig.InterfaceRegistry, encodingConfig.Codec))

	overwriteFlagDefaults(rootCmd, map[string]string{
		flags.FlagChainID:        string(app.ChosenNetwork.ChainID()),
		flags.FlagKeyringBackend: "test",
	})
}

func addModuleInitFlags(startCmd *cobra.Command) {
	crisis.AddModuleInitFlags(startCmd)
	wasm.AddModuleInitFlags(startCmd)
}

func overwriteFlagDefaults(c *cobra.Command, defaults map[string]string) {
	set := func(s *pflag.FlagSet, key, val string) {
		if f := s.Lookup(key); f != nil {
			f.DefValue = val
			f.Value.Set(val) //nolint:errcheck
		}
	}
	for key, val := range defaults {
		set(c.Flags(), key, val)
		set(c.PersistentFlags(), key, val)
	}
	for _, c := range c.Commands() {
		overwriteFlagDefaults(c, defaults)
	}
}

// genesisCommand builds genesis-related `simd genesis` command.
// Users may provide application specific commands as a parameter.
func genesisCommand(txConfig client.TxConfig, basicManager module.BasicManager, cmds ...*cobra.Command) *cobra.Command {
	cmd := genutilcli.GenesisCoreCommand(txConfig, basicManager, app.DefaultNodeHome)

	for _, sub_cmd := range cmds { //nolint:revive,staticcheck // sdk code copy
		cmd.AddCommand(sub_cmd)
	}
	return cmd
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.ValidatorCommand(),
		rpc.WaitTxCmd(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		server.QueryBlocksCmd(),
		authcmd.QueryTxCmd(),
		server.QueryBlockResultsCmd(),
	)

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetSimulateCmd(),
	)

	return cmd
}

const broadcastModeBlock = "block"

type txWriter struct {
	cdc          codec.Codec
	parentWriter io.Writer
	txRes        *sdk.TxResponse
}

func (txw *txWriter) Write(p []byte) (int, error) {
	writer := txw.parentWriter
	if writer == nil {
		writer = os.Stdout
	}

	// If output does not contain transaction object, just print the original output.
	res := &sdk.TxResponse{}
	if err := txw.cdc.UnmarshalJSON(p, res); err != nil || res.TxHash == "" {
		return writer.Write(p)
	}

	// Store the tx hash for further processing.
	txw.txRes = res
	return len(p), nil
}

func installAwaitBroadcastModeWrapper(cmd *cobra.Command) {
	// Read values of broadcast mode and output format set by the user.
	const flagHelp = "help"
	flagSet := pflag.NewFlagSet("pre-process", pflag.ExitOnError)
	flagSet.ParseErrorsWhitelist.UnknownFlags = true
	broadcastMode := flagSet.StringP(flags.FlagBroadcastMode, "b", "", "")
	originalOutputFormat := flagSet.StringP(flags.FlagOutput, "o", "", "")
	dryRun := flagSet.Bool(flags.FlagDryRun, false, "")
	// Dummy flag to turn off printing usage of this flag set
	flagSet.BoolP(flagHelp, "h", false, "")
	//nolint:errcheck // since we have set ExitOnError on flagset, we don't need to check for errors here
	flagSet.Parse(os.Args[1:])

	if *originalOutputFormat == "" {
		*originalOutputFormat = "text"
	}

	// If broadcast mode is "block", we need to set output format to json and broadcast mode to sync, so our
	// wrapper behaves correctly.
	if *broadcastMode == broadcastModeBlock {
		removeFlag(os.Args, "-b")
		os.Args = append(removeFlag(os.Args, "--"+flags.FlagBroadcastMode), "--"+flags.FlagBroadcastMode, flags.BroadcastSync)

		if !*dryRun {
			removeFlag(os.Args, "-o")
			os.Args = append(removeFlag(os.Args, "--"+flags.FlagOutput), "--"+flags.FlagOutput, "json")
		}
	}

	// Iterate over all the "tx" subcommands.
	cmds := []*cobra.Command{cmd}
	for len(cmds) > 0 {
		cmd := cmds[len(cmds)-1]
		cmds = cmds[:len(cmds)-1]
		cmds = append(cmds, cmd.Commands()...)

		// Modify description of "--broadcast-mode" flag to add "block" to available values.
		if broadcastModeFlag := cmd.LocalFlags().Lookup(flags.FlagBroadcastMode); broadcastModeFlag != nil {
			broadcastModeFlag.Usage = `Transaction broadcasting mode (sync|async|block)`
		}

		// We install our wrapper only if this is "block" broadcast mode and not a dry run.
		if *broadcastMode != broadcastModeBlock || *dryRun {
			continue
		}

		// Install wrapper for the command.
		originalRunE := cmd.RunE
		cmd.RunE = func(cmd *cobra.Command, args []string) error {
			// Set output handler in the client context.
			clientCtx := client.GetClientContextFromCmd(cmd)
			originalOutput := clientCtx.Output
			writer := &txWriter{
				cdc:          clientCtx.Codec,
				parentWriter: originalOutput,
			}
			clientCtx.Output = writer
			cmd.SetOut(writer)
			if err := client.SetCmdClientContext(cmd, clientCtx); err != nil {
				return errors.WithStack(err)
			}

			// Execute original command handler.
			if err := originalRunE(cmd, args); err != nil {
				return err
			}

			if writer.txRes.Code != 0 {
				clientCtx.Output = originalOutput
				cmd.SetOut(originalOutput)
				clientCtx.OutputFormat = *originalOutputFormat
				return errors.WithStack(clientCtx.PrintProto(writer.txRes))
			}

			// Once we read tx hash from the output produced by cosmos sdk we may await the transaction.
			awaitClientCtx := coreumclient.NewContextFromCosmosContext(coreumclient.DefaultContextConfig(), clientCtx).
				WithGRPCClient(clientCtx.GRPCClient).WithClient(clientCtx.Client)
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()

			res, err := coreumclient.AwaitTx(ctx, awaitClientCtx, writer.txRes.TxHash)
			if err != nil {
				return err
			}

			// Restore original output and print the transaction.
			clientCtx.Output = originalOutput
			clientCtx.OutputFormat = *originalOutputFormat
			return errors.WithStack(clientCtx.PrintProto(res))
		}
	}
}

// newApp creates the application.
func newApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	baseappOptions := server.DefaultBaseappOptions(appOpts)

	return app.New(
		logger, db, traceStore, true,
		appOpts,
		baseappOptions...,
	)
}

// appExport creates a new app (optionally at a given height) and exports state.
func appExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	var coreumApp *app.App

	// this check is necessary as we use the flag in x/upgrade.
	// we can exit more gracefully by checking the flag here.
	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	viperAppOpts, ok := appOpts.(*viper.Viper)
	if !ok {
		return servertypes.ExportedApp{}, errors.New("appOpts is not viper.Viper")
	}

	// overwrite the FlagInvCheckPeriod
	viperAppOpts.Set(server.FlagInvCheckPeriod, 1)
	appOpts = viperAppOpts

	if height != -1 {
		coreumApp = app.New(logger, db, traceStore, false, appOpts)

		if err := coreumApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		coreumApp = app.New(logger, db, traceStore, true, appOpts)
	}

	return coreumApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}

func tempDir() string {
	dir, err := os.MkdirTemp("", "cored")
	if err != nil {
		panic("failed to create temp dir: " + err.Error())
	}
	defer os.RemoveAll(dir) //nolint:errcheck // we don't care

	return dir
}
