package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/x/airdrop/types"
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
		CmdTxCreate(),
		CmdTxClaim(),
	)

	return cmd
}

func CmdTxCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [required_denom] [offer] [height] [description] --from [owner]",
		Args:  cobra.ExactArgs(4),
		Short: "Creates an airdrop",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Creates an airdrop.

Example:
$ %s tx airdrop create [required_denom] 0.1denom1,100denom2 100 "Airdrop 2022" --from [sender]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			sender := clientCtx.GetFromAddress()
			requiredDenom := args[0]
			offer, err := sdk.ParseDecCoins(args[1])
			if err != nil {
				return errors.Wrapf(err, "parsing offered coins failed")
			}
			height, err := strconv.ParseInt(args[2], 10, 64)
			if err != nil {
				return errors.WithStack(err)
			}
			description := args[3]

			msg := &types.MsgCreate{
				Sender:        sender.String(),
				Height:        height,
				Description:   description,
				RequiredDenom: requiredDenom,
				Offer:         offer,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdTxClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim [denom] [airdrop_id] --from [recipient]",
		Args:  cobra.ExactArgs(2),
		Short: "Claims an airdrop",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Claims an airdrop.

Example:
$ %s tx airdrop claim [denom] 10 --from [recipient]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			recipient := clientCtx.GetFromAddress()
			denom := args[0]
			airdropID, err := strconv.ParseInt(args[1], 10, 64)
			if err != nil {
				return errors.WithStack(err)
			}

			msg := &types.MsgClaim{
				Recipient: recipient.String(),
				Denom:     denom,
				Id:        uint64(airdropID),
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
