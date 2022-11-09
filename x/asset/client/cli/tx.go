package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdTxIssueFungibleToken(),
	)

	return cmd
}

// CmdTxIssueFungibleToken returns issue IssueFungibleToken cobra command.
func CmdTxIssueFungibleToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "issue-ft [symbol] [recipient_address] [initial_amount] [description] --from [issuer]",
		Args:  cobra.ExactArgs(4),
		Short: "Issue new fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Issues new fungible token.

Example:
$ %s tx asset issue-ft WBTC [recipient_address] 21000000 "Wrapped Bitcoin" --from [issuer]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			issuer := clientCtx.GetFromAddress()
			symbol := args[0]
			recipient := args[1]
			// if the recipient wasn't provided the signer is the recipient
			if recipient != "" {
				if _, err = sdk.AccAddressFromBech32(recipient); err != nil {
					return sdkerrors.Wrap(err, "invalid recipient")
				}
			} else {
				recipient = issuer.String()
			}

			// if the initial amount wasn't provided the amount is zero
			initialAmount := sdk.ZeroInt()
			if args[2] != "" {
				var ok bool
				initialAmount, ok = sdk.NewIntFromString(args[2])
				if !ok {
					return sdkerrors.Wrap(err, "invalid initial_amount")
				}
			}

			description := args[3]

			msg := &types.MsgIssueFungibleToken{
				Issuer:        issuer.String(),
				Symbol:        symbol,
				Recipient:     recipient,
				InitialAmount: initialAmount,
				Description:   description,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
