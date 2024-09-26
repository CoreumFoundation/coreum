package cli

import (
	"fmt"
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

	"github.com/CoreumFoundation/coreum/v4/x/dex/types"
)

const (
	// PriceFlag is price flag.
	PriceFlag = "price"
	// GoodTilBlockHeightFlag is good til block height flag.
	GoodTilBlockHeightFlag = "good-til-block-height"
	// GoodTilBlockTimeFlag is good til block time flag.
	GoodTilBlockTimeFlag = "good-til-block-time"
	// OrderTypeLimit is limit order type.
	OrderTypeLimit = "limit"
	// OrderTypeMarket is limit order market.
	OrderTypeMarket = "market"
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
//
//nolint:funlen // Despite the length function is still manageable
func CmdPlaceOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "place-order [type (limit,market)] [id] [base_denom] [quote_denom] [quantity] [side] --price 123e-2 --good-til-block-height 123 --good-til-block-time 1727124446 --from [sender]", //nolint:lll // string example
		Args:  cobra.ExactArgs(6),
		Short: "Place new order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Place new order.

Example:
$ %s tx %s place-order id1 denom1 denom2 123e-2 10000 buy --good-til-block-height 123 --from [sender]
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

			var orderType types.OrderType
			timeInForce := types.TIME_IN_FORCE_UNSPECIFIED
			switch args[0] {
			case OrderTypeLimit:
				orderType = types.ORDER_TYPE_LIMIT
				timeInForce = types.TIME_IN_FORCE_GTC
			case OrderTypeMarket:
				orderType = types.ORDER_TYPE_MARKET
			default:
				return errors.Errorf("unknown type '%s'", args[0])
			}

			id := args[1]
			baseDenom := args[2]
			quoteDenom := args[3]

			quantity, ok := sdkmath.NewIntFromString(args[4])
			if !ok {
				return sdkerrors.Wrap(err, "invalid quantity")
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

			msg := &types.MsgPlaceOrder{
				Sender:      sender.String(),
				Type:        orderType,
				ID:          id,
				BaseDenom:   baseDenom,
				QuoteDenom:  quoteDenom,
				Price:       price,
				Quantity:    quantity,
				Side:        types.Side(side),
				TimeInForce: timeInForce, // TODO(dzmitryhil) allow to modify
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
