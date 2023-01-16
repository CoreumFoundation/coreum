package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/x/asset/ft/types"
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

	cmd.AddCommand(CmdQueryToken())
	cmd.AddCommand(CmdQueryTokens())
	cmd.AddCommand(CmdQueryFrozenBalance())
	cmd.AddCommand(CmdQueryFrozenBalances())
	cmd.AddCommand(CmdQueryWhitelistedBalance())
	cmd.AddCommand(CmdQueryWhitelistedBalances())
	return cmd
}

// CmdQueryTokens return the QueryTokens cobra command.
//
//nolint:dupl // most code is identical, but reusing logic is not beneficial here.
func CmdQueryTokens() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tokens [issuer]",
		Args:  cobra.ExactArgs(1),
		Short: "Query fungible tokens",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query fungible tokens.

Example:
$ %[1]s query %s tokens [issuer]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			issuer := args[0]
			res, err := queryClient.Tokens(cmd.Context(), &types.QueryTokensRequest{
				Pagination: pageReq,
				Issuer:     issuer,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "tokens")

	return cmd
}

// CmdQueryToken return the QueryToken cobra command.
func CmdQueryToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token [denom]",
		Args:  cobra.ExactArgs(1),
		Short: "Query fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query fungible token details.

Example:
$ %[1]s query %s token [denom]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			denom := args[0]
			res, err := queryClient.Token(cmd.Context(), &types.QueryTokenRequest{
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

// CmdQueryFrozenBalances return the QueryFrozenBalances cobra command.
//
//nolint:dupl // most code is identical, but reusing logic is not beneficial here.
func CmdQueryFrozenBalances() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frozen-balances [account]",
		Args:  cobra.ExactArgs(1),
		Short: "Query fungible token frozen balances",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query frozen fungible token balances of an account.

Example:
$ %[1]s query %s frozen-balances [account]
`,
				version.AppName, types.ModuleName,
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

// CmdQueryFrozenBalance return the QueryFrozenBalance cobra command.
func CmdQueryFrozenBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frozen-balance [account] [denom]",
		Args:  cobra.ExactArgs(2),
		Short: "Query fungible token frozen balance",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query frozen fungible token balance of an account.

Example:
$ %[1]s query %s frozen-balance [account] [denom]
`,
				version.AppName, types.ModuleName,
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

// CmdQueryWhitelistedBalances return the QueryWhitelistedBalances cobra command.
//
//nolint:dupl // most code is identical, but reusing logic is not beneficial here.
func CmdQueryWhitelistedBalances() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whitelisted-balances [account]",
		Args:  cobra.ExactArgs(1),
		Short: "Query fungible token whitelisted balances",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query whitelisted fungible token balances of an account.

Example:
$ %[1]s query %s whitelisted-balances [account]
`,
				version.AppName, types.ModuleName,
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
			res, err := queryClient.WhitelistedBalances(cmd.Context(), &types.QueryWhitelistedBalancesRequest{
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
	flags.AddPaginationFlagsToCmd(cmd, "whitelisted balances")

	return cmd
}

// CmdQueryWhitelistedBalance return the QueryWhitelistedBalance cobra command.
func CmdQueryWhitelistedBalance() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "whitelisted-balance [account] [denom]",
		Args:  cobra.ExactArgs(2),
		Short: "Query fungible token whitelisted balance",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query whitelisted fungible token balance of an account.

Example:
$ %[1]s query %s whitelisted-balance [account] [denom]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			account := args[0]
			denom := args[1]
			res, err := queryClient.WhitelistedBalance(cmd.Context(), &types.QueryWhitelistedBalanceRequest{
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
