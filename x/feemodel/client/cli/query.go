package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/v2/x/feemodel/types"
)

const afterFlag = "after"

// GetQueryCmd returns the parent command for all x/feemodel CLI query commands. The
// provided clientCtx should have, at a minimum, a verifier, Tendermint RPC client,
// and marshaler set.
func GetQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Querying commands for the feemodel module",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		GetMinGasPriceCmd(),
		GetRecommendedGasPriceCmd(),
	)

	return cmd
}

// GetMinGasPriceCmd returns command for getting minimum gas price required by the network.
func GetMinGasPriceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "min-gas-price",
		Short: "Query for minimum gas price for current block required by the network",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			res, err := QueryGasPrice(cmd)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(&res.MinGasPrice)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

// QueryGasPrice queries the gas price.
func QueryGasPrice(cmd *cobra.Command) (*types.QueryMinGasPriceResponse, error) {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return nil, err
	}

	queryClient := types.NewQueryClient(clientCtx)

	ctx := cmd.Context()
	return queryClient.MinGasPrice(ctx, &types.QueryMinGasPriceRequest{})
}

// QueryParams queries the params.
func QueryParams(cmd *cobra.Command) (*types.QueryParamsResponse, error) {
	clientCtx, err := client.GetClientQueryContext(cmd)
	if err != nil {
		return nil, err
	}

	queryClient := types.NewQueryClient(clientCtx)

	ctx := cmd.Context()
	return queryClient.Params(ctx, &types.QueryParamsRequest{})
}

// GetRecommendedGasPriceCmd returns command for getting recommended gas price.
func GetRecommendedGasPriceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "recommended-gas-price",
		Short: fmt.Sprintf("Query for recommended gas price for `%s` blocks in future", afterFlag),
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			after, err := cmd.Flags().GetUint32(afterFlag)
			if err != nil {
				return err
			}

			res, err := queryClient.RecommendedGasPrice(cmd.Context(), &types.QueryRecommendedGasPriceRequest{
				AfterBlocks: after,
			})
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	cmd.Flags().Uint32(afterFlag, 10, "how many blocks in future to estimate gas price for.")

	return cmd
}
