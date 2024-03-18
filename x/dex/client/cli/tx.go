package cli

import (
	"fmt"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
		CmdTxCreateLimitOrder(),
	)

	return cmd
}

// CmdTxCreateLimitOrder returns CreateLimitOrder cobra command.
func CmdTxCreateLimitOrder() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-limit-order [offered-amount] [sell-price] --from [owner]",
		Args:  cobra.ExactArgs(2),
		Short: "Create limit order",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Creates limit-order.

Example:
$ %s tx %s create-limit-order 10wsatoshi 123.12wwei --from [issuer]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			owner := clientCtx.GetFromAddress()
			offeredAmount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid offered amount")
			}
			sellPrice, err := sdk.ParseDecCoin(args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid sell price")
			}

			msg := &types.MsgCreateLimitOrder{
				Owner:         owner.String(),
				OfferedAmount: offeredAmount,
				SellPrice:     sellPrice,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
