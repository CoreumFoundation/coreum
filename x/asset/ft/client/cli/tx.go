package cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
)

// Flags defined on transactions.
const (
	FeaturesFlag           = "features"
	BurnRateFlag           = "burn-rate"
	SendCommissionRateFlag = "send-commission-rate"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdTxIssue(),
		CmdTxMint(),
		CmdTxBurn(),
		CmdTxFreeze(),
		CmdTxUnfreeze(),
		CmdTxGloballyFreeze(),
		CmdTxGloballyUnfreeze(),
		CmdTxSetWhitelistedLimit(),
	)

	return cmd
}

// CmdTxIssue returns Issue cobra command.
//
//nolint:funlen // Despite the length function is still manageable
func CmdTxIssue() *cobra.Command {
	var allowedFeatures []string
	for _, n := range types.Feature_name {
		allowedFeatures = append(allowedFeatures, n)
	}
	sort.Strings(allowedFeatures)
	cmd := &cobra.Command{
		Use:   "issue [symbol] [subunit] [precision] [initial_amount] [description] --from [issuer] --features=" + strings.Join(allowedFeatures, ",") + " --burn-rate=0.12 --send-commission-rate=0.2",
		Args:  cobra.ExactArgs(5),
		Short: "Issue new fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Issues new fungible token.

Example:
$ %s tx %s issue WBTC wsatoshi 8 100000 "Wrapped Bitcoin Token" --from [issuer]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			issuer := clientCtx.GetFromAddress()
			symbol := args[0]
			subunit := args[1]
			precision, err := strconv.ParseUint(args[2], 10, 32)
			if err != nil {
				return sdkerrors.Wrap(err, "invalid precision")
			}

			// if the initial amount wasn't provided the amount is zero
			initialAmount := sdk.ZeroInt()
			if args[3] != "" {
				var ok bool
				initialAmount, ok = sdk.NewIntFromString(args[3])
				if !ok {
					return sdkerrors.Wrap(err, "invalid initial_amount")
				}
			}

			featuresString, err := cmd.Flags().GetStringSlice(FeaturesFlag)
			if err != nil {
				return errors.WithStack(err)
			}

			burnRate := sdk.NewDec(0)
			burnRateStr, err := cmd.Flags().GetString(BurnRateFlag)
			if err != nil {
				return errors.WithStack(err)
			}
			if len(burnRateStr) > 0 {
				burnRate, err = sdk.NewDecFromStr(burnRateStr)
				if err != nil {
					return errors.Wrapf(err, "invalid burn-rate")
				}
			}

			sendCommissionRate := sdk.NewDec(0)
			sendCommissionFeeStr, err := cmd.Flags().GetString(SendCommissionRateFlag)
			if err != nil {
				return errors.WithStack(err)
			}
			if len(sendCommissionFeeStr) > 0 {
				sendCommissionRate, err = sdk.NewDecFromStr(sendCommissionFeeStr)
				if err != nil {
					return errors.Wrapf(err, "invalid send-commission-rate")
				}
			}

			var features []types.Feature
			for _, str := range featuresString {
				feature, ok := types.Feature_value[str]
				if !ok {
					return errors.Errorf("unknown feature '%s',allowed features: %s", str, strings.Join(allowedFeatures, ","))
				}
				features = append(features, types.Feature(feature))
			}
			description := args[4]

			msg := &types.MsgIssue{
				Issuer:             issuer.String(),
				Symbol:             symbol,
				Subunit:            subunit,
				Precision:          uint32(precision),
				InitialAmount:      initialAmount,
				Description:        description,
				Features:           features,
				BurnRate:           burnRate,
				SendCommissionRate: sendCommissionRate,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().StringSlice(FeaturesFlag, []string{}, "Features to be enabled on fungible token. e.g --features="+strings.Join(allowedFeatures, ","))
	cmd.Flags().String(BurnRateFlag, "0", "Indicates the rate at which coins will be burned on top of the sent amount in every send action. Must be between 0 and 1.")
	cmd.Flags().String(SendCommissionRateFlag, "0", "Indicates the rate at which coins will be sent to the issuer on top of the sent amount in every send action. Must be between 0 and 1.")

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxMint returns Mint cobra command.
func CmdTxMint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint [amount] --from [sender]",
		Args:  cobra.ExactArgs(1),
		Short: "mint new amount of fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Mint new amount of fungible token.

Example:
$ %s tx %s mint 100000ABC-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8 --from [sender]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			sender := clientCtx.GetFromAddress()
			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid amount")
			}

			msg := &types.MsgMint{
				Sender: sender.String(),
				Coin:   amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxBurn returns Burn cobra command.
func CmdTxBurn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn [amount] --from [sender]",
		Args:  cobra.ExactArgs(1),
		Short: "burn some amount of fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Burn some amount of fungible token.

Example:
$ %s tx %s burn 100000ABC-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8 --from [sender]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			sender := clientCtx.GetFromAddress()
			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid amount")
			}

			msg := &types.MsgBurn{
				Sender: sender.String(),
				Coin:   amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxFreeze returns Freeze cobra command.
//
//nolint:dupl // most code is identical between Freeze/Unfreeze cmd, but reusing logic is not beneficial here.
func CmdTxFreeze() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "freeze [account_address] [amount] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Freeze a portion of fungible token on an account",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Freeze a portion of fungible token.

Example:
$ %s tx %s freeze [account_address] 100000ABC-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8 --from [sender]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			sender := clientCtx.GetFromAddress()
			account := args[0]
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid amount")
			}

			msg := &types.MsgFreeze{
				Sender:  sender.String(),
				Account: account,
				Coin:    amount,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxUnfreeze returns Unfreeze cobra command.
//
//nolint:dupl // most code is identical between Freeze/Unfreeze cmd, but reusing logic is not beneficial here.
func CmdTxUnfreeze() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfreeze [account_address] [amount] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Unfreeze a portion of the frozen fungible tokens",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unfreezes a portion of the frozen fungible token.

Example:
$ %s tx %s unfreeze [account_address] 100000ABC-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8 --from [sender]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			sender := clientCtx.GetFromAddress()
			account := args[0]
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid amount")
			}

			msg := &types.MsgUnfreeze{
				Sender:  sender.String(),
				Account: account,
				Coin:    amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxSetWhitelistedLimit returns SetWhitelistedLimit cobra command.
//
//nolint:dupl // most code is identical, but reusing logic is not beneficial here.
func CmdTxSetWhitelistedLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-whitelisted-limit [account_address] [amount] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Set whitelisted limit on an account",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Set whitelisted limit on an account.

Example:
$ %s tx %s set-whitelisted-limit [account_address] 100000ABC-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8-tEQ4 --from [sender]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			sender := clientCtx.GetFromAddress()
			account := args[0]
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid amount")
			}

			msg := &types.MsgSetWhitelistedLimit{
				Sender:  sender.String(),
				Account: account,
				Coin:    amount,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxGloballyFreeze returns GlobalFreeze cobra command.
func CmdTxGloballyFreeze() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "globally-freeze [denom] --from [sender]",
		Args:  cobra.ExactArgs(1),
		Short: "freezes fungible token so no operations are allowed with it before unfrozen",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Freezes fungible token so no operations are allowed with it before unfrozen.
This operation is idempotent so global freeze of already frozen token does nothing.

Example:
$ %s tx %s globally-freeze ABC-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8 --from [sender]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			sender := clientCtx.GetFromAddress()
			denom := args[0]

			msg := &types.MsgGloballyFreeze{
				Sender: sender.String(),
				Denom:  denom,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxGloballyUnfreeze returns GlobalUnfreeze cobra command.
func CmdTxGloballyUnfreeze() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "globally-unfreeze [denom] --from [sender]",
		Args:  cobra.ExactArgs(1),
		Short: "unfreezes fungible token and unblocks basic operations on it",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unfreezes fungible token and unblocks basic operations on it.
This operation is idempotent so global unfreezing of non-frozen token does nothing.

Example:
$ %s tx %s globally-unfreeze ABC-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8 --from [sender]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			sender := clientCtx.GetFromAddress()
			denom := args[0]

			msg := &types.MsgGloballyUnfreeze{
				Sender: sender.String(),
				Denom:  denom,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
