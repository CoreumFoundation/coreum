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
	)

	return cmd
}

// CmdTxIssueClass returns IssueClass cobra command.
func CmdTxIssueClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue-class [symbol] [name] [description] [uri] [uri_hash] --from [issuer]",
		Args:  cobra.ExactArgs(5),
		Short: "Issue new non-fungible token class",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Issue new non-fungible token class.

Example:
$ %s tx asset-nft issue-class abc "ABC Name" "ABC class description." https://my-class-meta.invalid/1 e000624 --from [issuer]
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
			name := args[1]
			description := args[2]
			uri := args[3]
			uriHash := args[4]

			msg := &types.MsgIssueClass{
				Issuer:      issuer.String(),
				Symbol:      symbol,
				Name:        name,
				Description: description,
				URI:         uri,
				URIHash:     uriHash,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

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
$ %s tx asset-nft mint abc-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8 id1 https://my-nft-meta.invalid/1 e000624 --from [sender]
`,
				version.AppName,
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
