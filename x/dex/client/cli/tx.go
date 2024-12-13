package cli

import (
	"fmt"
	"sort"
	"strings"
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/v5/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/v5/x/dex/types"
)

const (
	// PriceFlag is price flag.
	PriceFlag = "price"
	// GoodTilBlockHeightFlag is good til block height flag.
	GoodTilBlockHeightFlag = "good-til-block-height"
	// GoodTilBlockTimeFlag is good til block time flag.
	GoodTilBlockTimeFlag = "good-til-block-time"
	// TimeInForce is time-in-force flag.
	TimeInForce = "time-in-force"
)

// GetTxCmd returns the transaction commands for this module.
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      types.ModuleName + " transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdPlaceOrder(),
		CmdCancelOrder(),
		CmdCancelOrdersByDenom(),
	)

	return cmd
}

// CmdPlaceOrder returns PlaceOrder cobra command.
//
//nolint:funlen // Despite the length function is still manageable
func CmdPlaceOrder() *cobra.Command {
	availableTimeInForces := lo.Values(types.TimeInForce_name)
	sort.Strings(availableTimeInForces)
	availableOrderTypes := lo.Values(types.OrderType_name)
	sort.Strings(availableTimeInForces)
	availableSides := lo.Values(types.Side_name)
	sort.Strings(availableTimeInForces)
	cmd := &cobra.Command{
		Use:   "place-order [type (" + strings.Join(availableOrderTypes, ",") + ")] [id] [base_denom] [quote_denom] [quantity] [side (" + strings.Join(availableSides, ",") + ")] --price 123e-2 --time-in-force=" + strings.Join(availableTimeInForces, ",") + " --good-til-block-height=123 --good-til-block-time=1727124446 --from [sender]", //nolint:lll // string example
		Args:  cobra.ExactArgs(6),
		Short: "Place new order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Place new order.

Example:
$ %s tx %s cored tx dex place-order ORDER_TYPE_LIMIT "my-order-id1" denom1 denom2 1000 SIDE_SELL --price 12e-1 --time-in-force=TIME_IN_FORCE_GTC --from [sender]
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

			orderType, ok := types.OrderType_value[args[0]]
			if !ok {
				return errors.Errorf("unknown type '%s'", args[0])
			}

			id := args[1]
			baseDenom := args[2]
			quoteDenom := args[3]

			quantity, ok := sdkmath.NewIntFromString(args[4])
			if !ok {
				return errors.New("invalid quantity")
			}

			side, ok := types.Side_value[args[5]]
			if !ok {
				return errors.Errorf("unknown side '%s'", args[5])
			}

			priceStr, err := cmd.Flags().GetString(PriceFlag)
			if err != nil {
				return errors.WithStack(err)
			}
			var price *types.Price
			if priceStr != "" {
				priceV, err := types.NewPriceFromString(priceStr)
				if err != nil {
					return sdkerrors.Wrap(err, "invalid price")
				}
				price = &priceV
			}

			goodTilBlockHeight, err := cmd.Flags().GetUint64(GoodTilBlockHeightFlag)
			if err != nil {
				return errors.WithStack(err)
			}

			goodTilBlockTimeNum, err := cmd.Flags().GetInt64(GoodTilBlockTimeFlag)
			if err != nil {
				return errors.WithStack(err)
			}
			var goodTilBlockTime *time.Time
			if goodTilBlockTimeNum > 0 {
				goodTilBlockTime = lo.ToPtr(time.Unix(goodTilBlockTimeNum, 0))
			}

			timeInForceString, err := cmd.Flags().GetString(TimeInForce)
			timeInForceInt, ok := types.TimeInForce_value[timeInForceString]
			if !ok {
				return errors.Errorf(
					"unknown TimeInForce '%s',available TimeInForces: %s",
					timeInForceString, strings.Join(availableTimeInForces, ","),
				)
			}
			if err != nil {
				return errors.WithStack(err)
			}
			timeInForce := types.TimeInForce(timeInForceInt)

			msg := &types.MsgPlaceOrder{
				Sender:      sender.String(),
				Type:        types.OrderType(orderType),
				ID:          id,
				BaseDenom:   baseDenom,
				QuoteDenom:  quoteDenom,
				Price:       price,
				Quantity:    quantity,
				Side:        types.Side(side),
				TimeInForce: timeInForce,
			}

			if goodTilBlockHeight != 0 || goodTilBlockTime != nil {
				msg.GoodTil = &types.GoodTil{
					GoodTilBlockHeight: goodTilBlockHeight,
					GoodTilBlockTime:   goodTilBlockTime,
				}
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().String(PriceFlag, "", "Order price.")
	cmd.Flags().Uint64(GoodTilBlockHeightFlag, 0, "Good til block height.")
	cmd.Flags().Int64(GoodTilBlockTimeFlag, 0, "Good til block time.")
	cmd.Flags().String(TimeInForce, types.TIME_IN_FORCE_UNSPECIFIED.String(), "Time in force.")

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

// CmdCancelOrdersByDenom returns CancelOrdersByDenom cobra command.
func CmdCancelOrdersByDenom() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-orders-by-denom [account] [denom] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Cancel orders by denom",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Cancel orders by denom.

Example:
$ %s tx %s cancel-orders-by-denom %s denom1 --from [sender]
`,
				version.AppName, types.ModuleName, constant.AddressSampleTest,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			sender := clientCtx.GetFromAddress()
			account := args[0]
			denom := args[1]

			msg := &types.MsgCancelOrdersByDenom{
				Sender:  sender.String(),
				Account: account,
				Denom:   denom,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
