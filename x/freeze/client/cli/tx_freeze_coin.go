package cli

import (
	"strconv"

	"github.com/CoreumFoundation/coreum/x/freeze/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdFreezeCoin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "freeze-coin [address] [coin]",
		Short: "Broadcast message freezeCoin",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argAddress := args[0]
			argCoin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgFreezeCoin(
				clientCtx.GetFromAddress().String(),
				argAddress,
				argCoin,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
