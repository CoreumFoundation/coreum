package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

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
		CmdTxIssueClass(),
		CmdTxMint(),
		CmdTxBurn(),
		CmdTxFreeze(),
		CmdTxUnfreeze(),
	)

	return cmd
}

// CmdTxIssueClass returns IssueClass cobra command.
//
//nolint:nosnakecase // generated variable
func CmdTxIssueClass() *cobra.Command {
	allowedFeatures := make([]string, 0, len(types.ClassFeature_name))
	for _, n := range types.ClassFeature_name {
		allowedFeatures = append(allowedFeatures, n)
	}
	allowedFeaturesString := strings.Join(allowedFeatures, ",")

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("issue-class [symbol] [name] [description] [uri] [uri_hash] --from [issuer] --%s=%s", featuresFlag, allowedFeaturesString),
		Args:  cobra.ExactArgs(5),
		Short: "Issue new non-fungible token class",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Issue new non-fungible token class.

Example:
$ %s tx %s issue-class abc "ABC Name" "ABC class description." https://my-class-meta.invalid/1 e000624 --from [issuer] --%s=%s"
`,
				version.AppName, types.ModuleName, featuresFlag, allowedFeaturesString,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			issuer := clientCtx.GetFromAddress()
			symbol := args[0]
			name := args[1]
			description := args[2]
			uri := args[3]
			uriHash := args[4]

			featuresString, err := cmd.Flags().GetStringSlice(featuresFlag)
			if err != nil {
				return errors.WithStack(err)
			}

			var features []types.ClassFeature
			for _, str := range featuresString {
				feature, ok := types.ClassFeature_value[str] //nolint:nosnakecase
				if !ok {
					return errors.Errorf("unknown feature '%s',allowed allowedFeatures: %s", str, allowedFeaturesString)
				}
				features = append(features, types.ClassFeature(feature))
			}

			msg := &types.MsgIssueClass{
				Issuer:      issuer.String(),
				Symbol:      symbol,
				Name:        name,
				Description: description,
				URI:         uri,
				URIHash:     uriHash,
				Features:    features,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().StringSlice(featuresFlag, []string{}, fmt.Sprintf("Features to be enabled on non-fungible token. e.g --%s=%s", featuresFlag, allowedFeaturesString))
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxMint returns Mint cobra command.
func CmdTxMint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint [class-id] [id] [uri] [uri_hash] --from [sender]",
		Args:  cobra.ExactArgs(4),
		Short: "Mint new non-fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Mint new non-fungible token.

Example:
$ %s tx %s mint abc-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8 id1 https://my-nft-meta.invalid/1 e000624 --from [sender]
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
			classID := args[0]
			ID := args[1]
			uri := args[2]
			uriHash := args[3]

			msg := &types.MsgMint{
				Sender:  sender.String(),
				ClassID: classID,
				ID:      ID,
				URI:     uri,
				URIHash: uriHash,
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
		Use:   "burn [class-id] [id] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Burn non-fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Burn non-fungible token.

Example:
$ %s tx %s burn abc-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8 id1 --from [sender]
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
			classID := args[0]
			ID := args[1]

			msg := &types.MsgBurn{
				Sender:  sender.String(),
				ClassID: classID,
				ID:      ID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxFreeze returns Freeze cobra command.
func CmdTxFreeze() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "freeze [class-id] [id] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Freeze a non-fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Freeze a non-fungible token.

Example:
$ %s tx %s freeze abc-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8 id1 --from [sender]
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
			classID := args[0]
			ID := args[1]

			msg := &types.MsgFreeze{
				Sender:  sender.String(),
				ClassID: classID,
				ID:      ID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxUnfreeze returns Unfreeze cobra command.
func CmdTxUnfreeze() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfreeze [class-id] [id] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Unfreeze a non-fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unfreeze a non-fungible token.

Example:
$ %s tx %s unfreeze abc-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8 id1 --from [sender]
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
			classID := args[0]
			ID := args[1]

			msg := &types.MsgUnfreeze{
				Sender:  sender.String(),
				ClassID: classID,
				ID:      ID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
