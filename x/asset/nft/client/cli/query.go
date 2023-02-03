package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

// GetQueryCmd returns the cli query commands for the module.
func GetQueryCmd() *cobra.Command {
	// Group nft-asset queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryClass(),
		CmdQueryFrozen(),
		CmdQueryWhitelisted(),
		CmdQueryWhitelistedAccounts(),
	)
	return cmd
}

// CmdQueryClass return the QueryToken cobra command.
func CmdQueryClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "class [id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query non-fungible token class",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query non-fungible token class details.

Example:
$ %[1]s query %s class [id]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			id := args[0]
			res, err := queryClient.Class(cmd.Context(), &types.QueryClassRequest{
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

// CmdQueryFrozen return the CmdQueryFrozen cobra command.
func CmdQueryFrozen() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frozen [class-id] [id]",
		Args:  cobra.ExactArgs(2),
		Short: "Query if non-fungible token is frozen",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query if non-fungible token is frozen.

Example:
$ %[1]s query %s frozen [class-id] [id]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			classID := args[0]
			id := args[1]
			res, err := queryClient.Frozen(cmd.Context(), &types.QueryFrozenRequest{
				Id:      id,
				ClassId: classID,
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

// CmdQueryWhitelisted return the CmdQueryWhitelisted cobra command.
func CmdQueryWhitelisted() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whitelisted [class-id] [id] [account]",
		Args:  cobra.ExactArgs(3),
		Short: "Query if account is whitelisted for non-fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query if account is whitelisted for non-fungible token.

Example:
$ %s query %s whitelisted [class-id] [id] devcore15nc50svxu8xakdvks2hzd586xs9xmyqu9ws5ta
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			classID := args[0]
			id := args[1]
			account := args[2]
			res, err := queryClient.Whitelisted(cmd.Context(), &types.QueryWhitelistedRequest{
				Id:      id,
				ClassId: classID,
				Account: account,
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

// CmdQueryWhitelistedAccounts return the CmdQueryWhitelistedAccounts cobra command.
func CmdQueryWhitelistedAccounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whitelisted-accounts [class-id] [id]",
		Args:  cobra.ExactArgs(2),
		Short: "Query for the list whitelisted accounts for non-fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for whitelisted accounts for non-fungible token.

Example:
$ %s query %s whitelisted [class-id] [id]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			classID := args[0]
			id := args[1]

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.WhitelistedAccountsForNFT(cmd.Context(), &types.QueryWhitelistedAccountsForNFTRequest{
				Pagination: pageReq,
				Id:         id,
				ClassId:    classID,
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
