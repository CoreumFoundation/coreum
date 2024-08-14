package cli

import (
	"fmt"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdPlaceOrder(),
		CmdCancelOrder(),
	)

	return cmd
}

// CmdPlaceOrder returns PlaceOrder cobra command.
func CmdPlaceOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place-order [id] [base_denom] [quote_denom] [price] [quantity] [side] --from [sender]",
		Args:  cobra.ExactArgs(6),
		Short: "Place new order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Place new order.

Example:
$ %s tx %s place-order id1 denom1 denom2 123e-2 10000 buy --from [sender]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			sender := clientCtx.GetFromAddress()
			id := args[0]
			baseDenom := args[1]
			quoteDenom := args[2]

			price, err := types.NewPriceFromString(args[3])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid price")
			}

			quantity, ok := sdkmath.NewIntFromString(args[4])
			if !ok {
				return sdkerrors.Wrap(err, "invalid quantity")
			}

			side, ok := types.Side_value[args[5]]
			if !ok {
				return errors.Errorf("unknown side '%s'", args[5])
			}

			msg := &types.MsgPlaceOrder{
				Sender:     sender.String(),
				ID:         id,
				BaseDenom:  baseDenom,
				QuoteDenom: quoteDenom,
				Price:      price,
				Quantity:   quantity,
				Side:       types.Side(side),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdCancelOrder returns CancelOrder cobra command.
func CmdCancelOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-order [id] --from [sender]",
		Args:  cobra.ExactArgs(1),
		Short: "Cancel order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel order.

Example:
$ %s tx %s cancel-order id1 --from [sender]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			sender := clientCtx.GetFromAddress()
			id := args[0]

			msg := &types.MsgCancelOrder{
				Sender: sender.String(),
				ID:     id,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
