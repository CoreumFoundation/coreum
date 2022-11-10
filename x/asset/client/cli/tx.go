package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// Flags defined on transactions
const (
	featuresFlag = "features"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		FTCmd(),
	)

	return cmd
}

// FTCmd returns the subcommands for the fungible tokens
func FTCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "ft",
		Short:                      "fungible token transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdTxIssueFungibleToken(),
		CmdTxFreezeFungibleToken(),
		CmdTxUnfreezeFungibleToken(),
	)

	return cmd
}

// CmdTxIssueFungibleToken returns IssueFungibleToken cobra command.
func CmdTxIssueFungibleToken() *cobra.Command {
	allowedFeatures := []string{}
	for _, n := range types.FungibleTokenFeature_name { //nolint:nosnakecase
		allowedFeatures = append(allowedFeatures, n)
	}
	sort.Strings(allowedFeatures)
	cmd := &cobra.Command{
		Use:   "issue [symbol] [description] [recipient_address] [initial_amount] --from [issuer] --features=freezable,mintable,...",
		Args:  cobra.ExactArgs(4),
		Short: "Issue new fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Issues new fungible token.

Example:
$ %s tx asset ft issue ABC "ABC Token" [recipient_address] 100000 --from [issuer]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			issuer := clientCtx.GetFromAddress()
			symbol := args[0]
			description := args[1]
			recipient := args[2]
			// if the recipient wasn't provided the signer is the recipient
			if recipient != "" {
				if _, err = sdk.AccAddressFromBech32(recipient); err != nil {
					return sdkerrors.Wrap(err, "invalid recipient")
				}
			} else {
				recipient = issuer.String()
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

			featuresString, err := cmd.Flags().GetStringSlice(featuresFlag)
			if err != nil {
				return errors.WithStack(err)
			}

			var features []types.FungibleTokenFeature
			for _, str := range featuresString {
				feature, ok := types.FungibleTokenFeature_value[str] //nolint:nosnakecase
				if !ok {
					return errors.Errorf("Unknown feature '%s',allowed features: %s", str, strings.Join(allowedFeatures, ","))
				}
				features = append(features, types.FungibleTokenFeature(feature))
			}

			msg := &types.MsgIssueFungibleToken{
				Issuer:        issuer.String(),
				Symbol:        symbol,
				Description:   description,
				Recipient:     recipient,
				InitialAmount: initialAmount,
				Features:      features,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().StringSlice(featuresFlag, []string{}, "Features to be enabled on fungible token. e.g --features=freezable,mintable.")

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxFreezeFungibleToken returns FreezeFungibleToken cobra command.
//
//nolint:dupl // most code is identical between Freeze/Unfreeze cmd, but reusing logic is not beneficial here.
func CmdTxFreezeFungibleToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "freeze [account_address] [amount] --from [issuer]",
		Args:  cobra.ExactArgs(2),
		Short: "Freeze a portion of fungible token on an account",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Freeze a portion of fungible token.

Example:
$ %s tx asset ft freeze [account_address] 100000ABC-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8-tEQ4 --from [issuer]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			issuer := clientCtx.GetFromAddress()
			account := args[0]
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid amount")
			}

			msg := &types.MsgFreezeFungibleToken{
				Issuer:  issuer.String(),
				Account: account,
				Coin:    amount,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxUnfreezeFungibleToken returns FreezeFungibleToken cobra command.
//
//nolint:dupl // most code is identical between Freeze/Unfreeze cmd, but reusing logic is not beneficial here.
func CmdTxUnfreezeFungibleToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfreeze [account_address] [amount] --from [issuer]",
		Args:  cobra.ExactArgs(2),
		Short: "Unfreeze a portion of the frozen fungible tokens",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unfreezes a portion of the frozen fungible token.

Example:
$ %s tx asset ft unfreeze [account_address] 100000ABC-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8-tEQ4 --from [issuer]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			issuer := clientCtx.GetFromAddress()
			account := args[0]
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid amount")
			}

			msg := &types.MsgUnfreezeFungibleToken{
				Issuer:  issuer.String(),
				Account: account,
				Coin:    amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
