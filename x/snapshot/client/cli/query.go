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
		CmdQueryPendingFreezeRequests(),
		CmdQueryFrozenSnapshots(),
	)
	return cmd
}

func CmdQueryPendingFreezeRequests() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pending [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query pending freeze requests",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query pending freeze requests.

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
			res, err := queryClient.PendingFreezeRequests(cmd.Context(), &types.QueryPendingFreezeRequestsRequest{
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

func CmdQueryFrozenSnapshots() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frozen [address]",
		Args:  cobra.ExactArgs(1),
		Short: "Query frozen snapshots",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query frozen snapshots.

Example:
$ %[1]s query snapshot frozen [address]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			address := args[0]
			res, err := queryClient.FrozenSnapshots(cmd.Context(), &types.QueryFrozenSnapshotsRequest{
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
