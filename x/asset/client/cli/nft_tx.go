package cli

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/x/asset/types"
)

// NFTCmd returns the subcommands for the non-fungible tokens
func NFTCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "nft",
		Short:                      "non fungible token transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdTxCreateNonFungibleTokenClass(),
		CmdTxMintNonFungibleToken(),
	)

	return cmd
}

// CmdTxCreateNonFungibleTokenClass returns CreateNonFungibleTokenClass cobra command.
func CmdTxCreateNonFungibleTokenClass() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-class [symbol] [name] [description] [uri] [uri_hash] --from [creator]",
		Args:  cobra.ExactArgs(5),
		Short: "Create new non-fungible token class",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Create new non-fungible token class.

Example:
$ %s tx asset nft create-class abc "ABC Name" "ABC token description." metada-uri uri-hash --from [sender]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			creator := clientCtx.GetFromAddress()
			symbol := args[0]
			name := args[1]
			description := args[2]
			uri := args[3]
			uriHash := args[4]

			msg := &types.MsgCreateNonFungibleTokenClass{
				Creator:     creator.String(),
				Symbol:      symbol,
				Name:        name,
				Description: description,
				Uri:         uri,
				UriHash:     uriHash,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxMintNonFungibleToken returns MintNonFungibleToken cobra command.
func CmdTxMintNonFungibleToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint [class-id] [id] [uri] [uri_hash] --from [sender]",
		Args:  cobra.ExactArgs(4),
		Short: "Mint new non-fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Mint new non-fungible token.

Example:
$ %s tx asset nft mint abc-devcore1tr3w86yesnj8f290l6ve02cqhae8x4ze0nk0a8 id1 metada-uri uri-hash --from [sender]
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			sender := clientCtx.GetFromAddress()
			classID := args[0]
			ID := args[1]
			uri := args[2]
			uriHash := args[3]

			msg := &types.MsgMintNonFungibleToken{
				Sender:  sender.String(),
				ClassId: classID,
				Id:      ID,
				Uri:     uri,
				UriHash: uriHash,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
