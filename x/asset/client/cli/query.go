package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// GetQueryCmd returns the cli query commands for the module.
func GetQueryCmd() *cobra.Command {
	// Group asset queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryAsset())
	return cmd
}

// CmdQueryAsset return the QueryAsset cobra command.
func CmdQueryAsset() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show [id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query one asset",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query asset details by id.

Example:
$ %[1]s query asset show {id}
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			id := args[0]
			res, err := queryClient.Asset(cmd.Context(), &types.QueryAssetRequest{
				Id: id,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
