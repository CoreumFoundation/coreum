package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/x/snapshot/types"
)

func GetQueryCmd() *cobra.Command {
	// Group asset queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryPending(),
		CmdQuerySnapshots(),
	)
	return cmd
}

func CmdQueryPending() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query pending snapshot requests",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query pending snapshot requests.

Example:
$ %[1]s query snapshot pending [address]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			address := args[0]
			res, err := queryClient.Pending(cmd.Context(), &types.QueryPendingRequest{
				Address: address,
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

func CmdQuerySnapshots() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query snapshots",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query snapshots.

Example:
$ %[1]s query snapshot list [address]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			address := args[0]
			res, err := queryClient.Snapshots(cmd.Context(), &types.QuerySnapshotsRequest{
				Address: address,
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
