// Package cosmoscmd contains the root of the commands.
// The commands root.go copied from https://github.com/cosmos/cosmos-sdk/blob/v0.47.4/simapp/simd/cmd/root.go.
// under APACHE2.0 LICENSE
package cosmoscmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	rosettaCmd "cosmossdk.io/tools/rosetta/cmd"
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	dbm "github.com/cometbft/cometbft-db"
	tmcfg "github.com/cometbft/cometbft/config"
	"github.com/cometbft/cometbft/libs/log"
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
	sdk "github.com/cosmos/cosmos-sdk/types"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/CoreumFoundation/coreum/v4/app"
	coreumclient "github.com/CoreumFoundation/coreum/v4/pkg/client"
	"github.com/CoreumFoundation/coreum/v4/pkg/config"
)

const ledgerAppName = "Coreum"

// NewRootCmd creates a new root command for simd. It is called once in the
// main function.
func NewRootCmd() *cobra.Command {
	// we "pre"-instantiate the application for getting the injected/configured encoding configuration
	encodingConfig := config.NewEncodingConfig(app.ModuleBasics)
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

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			customTMConfig := app.ChosenNetwork.NodeConfig.TendermintNodeConfig(initTendermintConfig())

			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customTMConfig)
		},
	}

	initRootCmd(rootCmd, encodingConfig)

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
	srvCfg.MinGasPrices = fmt.Sprintf("0.00000000000000001%s", app.ChosenNetwork.Denom())

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

	defaultWasmConfig := wasmtypes.DefaultWasmConfig()
	customAppConfig := CustomAppConfig{
		Config: *srvCfg,
		WASM: WASMConfig{
			QueryGasLimit:   defaultWasmConfig.SmartQueryGasLimit,
			MemoryCacheSize: defaultWasmConfig.MemoryCacheSize,
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

func initRootCmd(rootCmd *cobra.Command, encodingConfig config.EncodingConfig) {
	cfg := sdk.GetConfig()
	cfg.Seal()

	rootCmd.AddCommand(
		InitCmd(app.DefaultNodeHome),
		debug.Cmd(),
		clientconfig.Cmd(),
		pruning.Cmd(newApp, app.DefaultNodeHome),
		snapshot.Cmd(newApp),
		GenerateDevnetCmd(),
	)

	server.AddCommands(rootCmd, app.DefaultNodeHome, newApp, appExport, addModuleInitFlags)

	// add keybase, auxiliary RPC, query, genesis, and tx child commands
	rootCmd.AddCommand(
		rpc.StatusCommand(),
		genesisCommand(encodingConfig),
		queryCommand(),
		txCommand(),
		keys.Commands(app.DefaultNodeHome),
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
func genesisCommand(encodingConfig config.EncodingConfig, cmds ...*cobra.Command) *cobra.Command {
	cmd := genutilcli.GenesisCoreCommand(encodingConfig.TxConfig, app.ModuleBasics, app.DefaultNodeHome)

	for _, sub_cmd := range cmds { //nolint:revive,stylecheck // sdk code copy
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
		authcmd.GetAccountCmd(),
		rpc.ValidatorCommand(),
		rpc.BlockCommand(),
		authcmd.QueryTxsByEventsCmd(),
		authcmd.QueryTxCmd(),
	)

	app.ModuleBasics.AddQueryCommands(cmd)

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

	app.ModuleBasics.AddTxCommands(cmd)
	addQueryGasPriceToAllLeafs(cmd)

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetAuxToFeeCommand(),
	)

	installAwaitBroadcastModeWrapper(cmd)

	return cmd
}

const broadcastModeBlock = "block"

type txWriter struct {
	cdc          codec.Codec
	parentWriter io.Writer
	txHash       string
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
	txw.txHash = res.TxHash
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
			if err := client.SetCmdClientContext(cmd, clientCtx); err != nil {
				return errors.WithStack(err)
			}

			// Execute original command handler.
			if err := originalRunE(cmd, args); err != nil {
				return err
			}

			// Once we read tx hash from the output produced by cosmos sdk we may await the transaction.
			awaitClientCtx := coreumclient.NewContext(coreumclient.DefaultContextConfig(), app.ModuleBasics).
				WithGRPCClient(clientCtx.GRPCClient).WithClient(clientCtx.Client)
			ctx, cancel := context.WithTimeout(cmd.Context(), 10*time.Second)
			defer cancel()

			res, err := coreumclient.AwaitTx(ctx, awaitClientCtx, writer.txHash)
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
