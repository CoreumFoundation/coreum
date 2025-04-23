package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/v6/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/v6/x/dex/types"
)

// GetQueryCmd returns the cli query commands for the module.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdQueryOrder())
	cmd.AddCommand(CmdQueryOrders())
	cmd.AddCommand(CmdQueryOrderBooks())
	cmd.AddCommand(CmdQueryOrderBookParams())
	cmd.AddCommand(CmdQueryOrderBookOrders())
	cmd.AddCommand(CmdQueryAccountDenomOrdersCount())

	return cmd
}

// CmdQueryParams implements a command to fetch dex parameters.
func CmdQueryParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "params",
		Short: fmt.Sprintf("Query the current %s parameters", types.ModuleName),
		Args:  cobra.NoArgs,
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query parameters for the %s module:

Example:
$ %[1]s query %s params
`,
				types.ModuleName, version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryParamsRequest{}
			res, err := queryClient.Params(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// CmdQueryOrder returns the QueryOrder cobra command.
func CmdQueryOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order [creator] [id]",
		Args:  cobra.ExactArgs(2),
		Short: "Query order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query order.

Example:
$ %[1]s query %s order %s id1
`,
				version.AppName, types.ModuleName, constant.AddressSampleTest,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Order(cmd.Context(), &types.QueryOrderRequest{
				Creator: args[0],
				Id:      args[1],
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

// CmdQueryOrders returns the QueryOrders cobra command.
func CmdQueryOrders() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "orders [creator]",
		Args:  cobra.ExactArgs(1),
		Short: "Query orders",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query orders.

Example:
$ %[1]s query %s orders %s
`,
				version.AppName, types.ModuleName, constant.AddressSampleTest,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.Orders(cmd.Context(), &types.QueryOrdersRequest{
				Creator:    args[0],
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "orders")

	return cmd
}

// CmdQueryOrderBooks returns the QueryOrderBooks cobra command.
func CmdQueryOrderBooks() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order-books",
		Args:  cobra.NoArgs,
		Short: "Query order books",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query order books.

Example:
$ %[1]s query %s order-books
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

			res, err := queryClient.OrderBooks(cmd.Context(), &types.QueryOrderBooksRequest{
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "order-books")

	return cmd
}

// CmdQueryOrderBookParams returns the QueryOrderBookParams cobra command.
func CmdQueryOrderBookParams() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order-book-params [base_denom] [quote_denom]",
		Args:  cobra.ExactArgs(2),
		Short: "Query order book params",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query order book params.

Example:
$ %[1]s query %s order-book-params denom1 denom2 
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.OrderBookParams(cmd.Context(), &types.QueryOrderBookParamsRequest{
				BaseDenom:  args[0],
				QuoteDenom: args[1],
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

// CmdQueryOrderBookOrders returns the QueryOrderBookOrders cobra command.
func CmdQueryOrderBookOrders() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order-book-orders [base_denom] [quote_denom] [side]",
		Args:  cobra.ExactArgs(3),
		Short: "Query order book orders",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query order book orders.

Example:
$ %[1]s query %s order-book-orders denom1 denom2 buy 
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

			side, ok := types.Side_value[args[2]]
			if !ok {
				return errors.Errorf("unknown side '%s'", args[2])
			}

			res, err := queryClient.OrderBookOrders(cmd.Context(), &types.QueryOrderBookOrdersRequest{
				BaseDenom:  args[0],
				QuoteDenom: args[1],
				Side:       types.Side(side),
				Pagination: pageReq,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "order-book-orders")

	return cmd
}

// CmdQueryAccountDenomOrdersCount returns the QueryAccountDenomOrdersCount cobra command.
func CmdQueryAccountDenomOrdersCount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account-denom-orders-count [account] [denom]",
		Args:  cobra.ExactArgs(2),
		Short: "Query account denom orders count",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query account denom orders count.

Example:
$ %[1]s query %s account-denom-orders-count %s denom1
`,
				version.AppName, types.ModuleName, constant.AddressSampleTest,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.AccountDenomOrdersCount(cmd.Context(), &types.QueryAccountDenomOrdersCountRequest{
				Account: args[0],
				Denom:   args[1],
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
