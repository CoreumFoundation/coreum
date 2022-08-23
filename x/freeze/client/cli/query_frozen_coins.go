package cli

import (
	"strconv"

	"github.com/CoreumFoundation/coreum/x/freeze/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdFrozenCoins() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frozen-coins [account]",
		Short: "Query frozenCoins",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqAccount := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryFrozenCoinsRequest{

				Account: reqAccount,
			}

			res, err := queryClient.FrozenCoins(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
