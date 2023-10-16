package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/v3/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/v3/x/asset/nft/types"
)

// Flags defined on queries.
const (
	IssuerFlag = "issuer"
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
		CmdQueryClasses(),
		CmdQueryFrozen(),
		CmdQueryWhitelisted(),
		CmdQueryWhitelistedAccounts(),
		CmdQueryClassWhitelistedAccounts(),
		CmdQueryBurnt(),
		CmdQueryParams(),
	)

	return cmd
}

// CmdQueryClass return the QueryClass cobra command.
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

// CmdQueryClasses return the QueryClasses cobra command.
func CmdQueryClasses() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "classes",
		Args:  cobra.ExactArgs(0),
		Short: "Query non-fungible token classes",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query non-fungible token classes.

Example:
$ %[1]s query %s classes --issuer %s
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

			issuerString, err := cmd.Flags().GetString(IssuerFlag)
			if err != nil {
				return errors.WithStack(err)
			}

			res, err := queryClient.Classes(cmd.Context(), &types.QueryClassesRequest{
				Pagination: pageReq,
				Issuer:     issuerString,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	cmd.Flags().String(IssuerFlag, "", fmt.Sprintf("Class issuer address. e.g %s", constant.AddressSampleTest))
	flags.AddPaginationFlagsToCmd(cmd, "classes")
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
$ %s query %s whitelisted [class-id] [id] %s
`,
				version.AppName, types.ModuleName, constant.AddressSampleTest,
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
$ %s query %s whitelisted-accounts [class-id] [id]
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
	flags.AddPaginationFlagsToCmd(cmd, "whitelisted accounts")

	return cmd
}

// CmdQueryClassWhitelistedAccounts return the CmdQueryWhitelistedAccounts cobra command.
func CmdQueryClassWhitelistedAccounts() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "class-whitelisted-accounts [class-id]",
		Args:  cobra.ExactArgs(1),
		Short: "Query for whitelisted accounts for a class of non-fungible tokens",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for whitelisted accounts for a class of non-fungible tokens.

Example:
$ %s query %s class-whitelisted-accounts [class-id]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			classID := args[0]

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.ClassWhitelistedAccounts(cmd.Context(), &types.QueryClassWhitelistedAccountsRequest{
				Pagination: pageReq,
				ClassId:    classID,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "whitelisted accounts")

	return cmd
}

// CmdQueryParams implements a command to fetch assetnft parameters.
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

// CmdQueryBurnt return the CmdQueryBurnt cobra command.
func CmdQueryBurnt() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burnt [class-id] [id]",
		Args:  cobra.RangeArgs(1, 2),
		Short: "Query for the burnt NFTs",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Query for the burnt NFTs in a class.

Example:
$ %s query %s burnt [class-id] [id]
$ %s query %s burnt [class-id] 
`,
				version.AppName, types.ModuleName,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			queryClient := types.NewQueryClient(clientCtx)

			classID := args[0]

			if len(args) == 2 {
				id := args[1]
				res, err := queryClient.BurntNFT(cmd.Context(), &types.QueryBurntNFTRequest{
					ClassId: classID,
					NftId:   id,
				})
				if err != nil {
					return err
				}
				return clientCtx.PrintProto(res)
			}

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			res, err := queryClient.BurntNFTsInClass(cmd.Context(), &types.QueryBurntNFTsInClassRequest{
				Pagination: pageReq,
				ClassId:    classID,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)
	flags.AddPaginationFlagsToCmd(cmd, "burnt NFTs")

	return cmd
}
