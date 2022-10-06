package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/x/asset/types"
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
		CmdTxIssueFTAsset(),
	)

	return cmd
}

// CmdTxIssueFTAsset return issue IssueAsset cobra command.
func CmdTxIssueFTAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue-ft [recipient_address] [code] [description] [precision] [initial_amount] --from [signer_address]",
		Args:  cobra.ExactArgs(5),
		Short: "Issue new asset",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Issues new asset.

Example:
$ %s tx asset issue-ft [recipient_address] BTC "BTC Token" 18 100000 --from [signer_address]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			recipient := args[0]
			code := args[1]
			description := args[2]
			precision, err := strconv.ParseUint(args[3], 10, 32)
			if err != nil {
				return err
			}
			initialAmount, ok := sdk.NewIntFromString(args[4])
			if !ok {
				return sdkerrors.Wrap(err, "invalid initial_amount")
			}
			from := clientCtx.GetFromAddress()

			msg := &types.MsgIssueAsset{
				From: from.String(),
				Definition: &types.AssetDefinition{
					Recipient:   recipient,
					Type:        types.AssetType_FT, //nolint:nosnakecase // protogen
					Code:        code,
					Description: description,
					Ft: &types.FTCustomDefinition{
						Precision:     uint32(precision),
						InitialAmount: initialAmount,
					},
				},
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
