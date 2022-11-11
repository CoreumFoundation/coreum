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

	cmd.AddCommand(GetFTQueryCmd())
	return cmd
}

// GetFTQueryCmd returns the cli query commands for fungible tokens.
func GetFTQueryCmd() *cobra.Command {
	// Group asset queries under a subcommand
	cmd := &cobra.Command{
		Use:                        "ft",
		Short:                      "Querying commands for fungible tokens",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryFungibleToken())
	cmd.AddCommand(CmdQueryFungibleTokenFrozenBalance())
	cmd.AddCommand(CmdQueryFungibleTokenFrozenBalances())
	return cmd
}

// CmdQueryFungibleToken return the QueryFungibleToken cobra command.
func CmdQueryFungibleToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query fungible token details.

Example:
$ %[1]s query asset ft info [denom]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			denom := args[0]
			res, err := queryClient.FungibleToken(cmd.Context(), &types.QueryFungibleTokenRequest{
				Denom: denom,
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

// CmdQueryFungibleTokenFrozenBalances return the QueryFungibleToken cobra command.
func CmdQueryFungibleTokenFrozenBalances() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frozen-balances [account]",
		Args:  cobra.ExactArgs(1),
		Short: "Query fungible token frozen balances",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query frozen fungible token balances of an account.

Example:
$ %[1]s query asset ft frozen-balances [account]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			account := args[0]
			res, err := queryClient.FrozenBalances(cmd.Context(), &types.QueryFrozenBalancesRequest{
				Account:    account,
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "frozen balances")

	return cmd
}

// CmdQueryFungibleTokenFrozenBalance return the QueryFungibleToken cobra command.
func CmdQueryFungibleTokenFrozenBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frozen-balance [account] [denom]",
		Args:  cobra.ExactArgs(2),
		Short: "Query fungible token frozen balance",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query frozen fungible token balance of an account.

Example:
$ %[1]s query asset ft frozen-balance [account] [denom]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			account := args[0]
			denom := args[1]
			res, err := queryClient.FrozenBalance(cmd.Context(), &types.QueryFrozenBalanceRequest{
				Account: account,
				Denom:   denom,
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
