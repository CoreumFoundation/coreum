package cli

import (
    "strconv"
	
	"github.com/spf13/cobra"
    "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/CoreumFoundation/coreum/x/freeze/types"
)

var _ = strconv.Itoa(0)

func CmdUnfreezeCoin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfreeze-coin [address] [denom]",
		Short: "Broadcast message unfreezeCoin",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
      		 argAddress := args[0]
             argDenom := args[1]
            
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgUnfreezeCoin(
				clientCtx.GetFromAddress().String(),
				argAddress,
				argDenom,
				
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