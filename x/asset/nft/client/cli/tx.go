package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/v4/x/asset/nft/types"
)

// Flags defined on transactions.
const (
	AuthzFileFlag   = "auth-file"
	ExpirationFlag  = "expiration"
	FeaturesFlag    = "features"
	RoyaltyRateFlag = "royalty-rate"
	RecipientFlag   = "recipient"
	URIFlag         = "uri"
	URIHashFlag     = "uri_hash"
	DataFileFlag    = "data-file"
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
		CmdTxIssueClass(),
		CmdTxMint(),
		CmdTxBurn(),
		CmdTxFreeze(),
		CmdTxUnfreeze(),
		CmdTxClassFreeze(),
		CmdTxClassUnfreeze(),
		CmdTxWhitelist(),
		CmdTxUnwhitelist(),
		CmdTxClassWhitelist(),
		CmdTxClassUnwhitelist(),
		CmdGrantAuthorization(),
	)

	return cmd
}

// CmdTxIssueClass returns IssueClass cobra command.
func CmdTxIssueClass() *cobra.Command {
	allowedFeatures := make([]string, 0, len(types.ClassFeature_name))
	for _, n := range types.ClassFeature_name {
		allowedFeatures = append(allowedFeatures, n)
	}
	allowedFeaturesString := strings.Join(allowedFeatures, ",")

	cmd := &cobra.Command{
		//nolint:lll // breaking this down will make it look worse when printed to user screen.
		Use:   fmt.Sprintf("issue-class [symbol] [name] [description] --from [issuer] --%s=%s --uri https://my-token-meta.invalid/1 --uri_hash e000624 --data-file [path]", FeaturesFlag, allowedFeaturesString),
		Args:  cobra.ExactArgs(3),
		Short: "Issue new non-fungible token class",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Issue new non-fungible token class.

Example:
$ %s tx %s issue-class abc "ABC Name" "ABC class description." --from [issuer] --data-file [path] --%s=%s"
`,
				version.AppName, types.ModuleName, FeaturesFlag, allowedFeaturesString,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			issuer := clientCtx.GetFromAddress()
			symbol := args[0]
			name := args[1]
			description := args[2]
			royaltyStr, err := cmd.Flags().GetString(RoyaltyRateFlag)
			if err != nil {
				return errors.WithStack(err)
			}
			royaltyRate, err := sdk.NewDecFromStr(royaltyStr)
			if err != nil {
				return errors.WithStack(err)
			}

			featuresString, err := cmd.Flags().GetStringSlice(FeaturesFlag)
			if err != nil {
				return errors.WithStack(err)
			}

			var features []types.ClassFeature
			for _, str := range featuresString {
				feature, ok := types.ClassFeature_value[str]
				if !ok {
					return errors.Errorf("unknown feature '%s',allowed allowedFeatures: %s", str, allowedFeaturesString)
				}
				features = append(features, types.ClassFeature(feature))
			}

			uri, err := cmd.Flags().GetString(URIFlag)
			if err != nil {
				return errors.WithStack(err)
			}

			uriHash, err := cmd.Flags().GetString(URIHashFlag)
			if err != nil {
				return errors.WithStack(err)
			}

			data, err := getDataFromFile(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgIssueClass{
				Issuer:      issuer.String(),
				Symbol:      symbol,
				Name:        name,
				Description: description,
				URI:         uri,
				URIHash:     uriHash,
				Data:        data,
				Features:    features,
				RoyaltyRate: royaltyRate,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	cmd.Flags().StringSlice(
		FeaturesFlag,
		[]string{},
		fmt.Sprintf("Features to be enabled on non-fungible token. e.g --%s=%s", FeaturesFlag, allowedFeaturesString),
	)
	//nolint:lll // breaking this down will make it look worse when printed to user screen.
	cmd.Flags().String(RoyaltyRateFlag, "0", fmt.Sprintf("%s is a number between 0 and 1, and will be used to determine royalties sent to issuer, when an nft in this class is traded.", RoyaltyRateFlag))
	cmd.Flags().String(URIFlag, "", "Class URI.")
	cmd.Flags().String(URIHashFlag, "", "Class URI hash.")
	cmd.Flags().String(DataFileFlag, "", "path to the file containing data.")

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxMint returns Mint cobra command.
func CmdTxMint() *cobra.Command {
	cmd := &cobra.Command{
		//nolint:lll // breaking this down will make it look worse when printed to user screen.
		Use:   "mint [class-id] [id] --from [sender] --uri https://my-token-meta.invalid/1 --uri_hash e000624 --data-file [path]",
		Args:  cobra.ExactArgs(2),
		Short: "Mint new non-fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Mint new non-fungible token.

Example:
$ %s tx %s mint abc-%s id1 --from [sender] --data-file [path]
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
			recipient, err := cmd.Flags().GetString(RecipientFlag)
			if err != nil {
				return errors.WithStack(err)
			}

			classID := args[0]
			ID := args[1]

			uri, err := cmd.Flags().GetString(URIFlag)
			if err != nil {
				return errors.WithStack(err)
			}

			uriHash, err := cmd.Flags().GetString(URIHashFlag)
			if err != nil {
				return errors.WithStack(err)
			}

			data, err := getDataFromFile(cmd)
			if err != nil {
				return err
			}

			msg := &types.MsgMint{
				Sender:    sender.String(),
				Recipient: recipient,
				ClassID:   classID,
				ID:        ID,
				URI:       uri,
				URIHash:   uriHash,
				Data:      data,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	//nolint:lll // breaking it down will make it look worse when printed to user screen
	cmd.Flags().String(RecipientFlag, "", "Address to send minted token to, if not specified minted token is sent to the class issuer")
	cmd.Flags().String(URIFlag, "", "NFT URI.")
	cmd.Flags().String(URIHashFlag, "", "NFT URI hash.")
	cmd.Flags().String(DataFileFlag, "", "path to the file containing data.")

	return cmd
}

// CmdTxBurn returns Burn cobra command.
func CmdTxBurn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn [class-id] [id] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Burn non-fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Burn non-fungible token.

Example:
$ %s tx %s burn abc-%s id1 --from [sender]
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
			classID := args[0]
			ID := args[1]

			msg := &types.MsgBurn{
				Sender:  sender.String(),
				ClassID: classID,
				ID:      ID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxFreeze returns Freeze cobra command.
func CmdTxFreeze() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "freeze [class-id] [id] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Freeze a non-fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Freeze a non-fungible token.

Example:
$ %s tx %s freeze abc-%s id1 --from [sender]
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
			classID := args[0]
			ID := args[1]

			msg := &types.MsgFreeze{
				Sender:  sender.String(),
				ClassID: classID,
				ID:      ID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxUnfreeze returns Unfreeze cobra command.
func CmdTxUnfreeze() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfreeze [class-id] [id] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Unfreeze a non-fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unfreeze a non-fungible token.

Example:
$ %s tx %s unfreeze abc-%s id1 --from [sender]
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
			classID := args[0]
			ID := args[1]

			msg := &types.MsgUnfreeze{
				Sender:  sender.String(),
				ClassID: classID,
				ID:      ID,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxWhitelist returns Whitelist cobra command.
func CmdTxWhitelist() *cobra.Command { //nolint:dupl // all CLI commands are similar.
	cmd := &cobra.Command{
		Use:   "whitelist [class-id] [id] [account] --from [sender]",
		Args:  cobra.ExactArgs(3),
		Short: "Whitelist an account for a non-fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Whitelist an account for a non-fungible token.

Example:
$ %s tx %s whitelist abc-%[3]s id1 %[3]s --from [sender]
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
			classID := args[0]
			ID := args[1]
			account := args[2]

			msg := &types.MsgAddToWhitelist{
				Sender:  sender.String(),
				ClassID: classID,
				ID:      ID,
				Account: account,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxUnwhitelist returns Unwhitelist cobra command.
func CmdTxUnwhitelist() *cobra.Command { //nolint:dupl // all CLI commands are similar.
	cmd := &cobra.Command{
		Use:   "unwhitelist [class-id] [id] [account] --from [sender]",
		Args:  cobra.ExactArgs(3),
		Short: "Unwhitelist an account for a non-fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unwhitelist an account for a non-fungible token.

Example:
$ %s tx %s unwhitelist abc-%[3]s id1 %[3]s --from [sender]
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
			classID := args[0]
			ID := args[1]
			account := args[2]

			msg := &types.MsgRemoveFromWhitelist{
				Sender:  sender.String(),
				ClassID: classID,
				ID:      ID,
				Account: account,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxClassWhitelist returns ClassWhitelist cobra command.
func CmdTxClassWhitelist() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "class-whitelist [class-id] [account] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Whitelist an account for a class of non-fungible tokens",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Whitelist an account for a class of non-fungible tokens.

Example:
$ %s tx %s class-whitelist abc-%[3]s %[3]s --from [sender]
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
			classID := args[0]
			account := args[1]

			msg := &types.MsgAddToClassWhitelist{
				Sender:  sender.String(),
				ClassID: classID,
				Account: account,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxClassUnwhitelist returns ClassUnwhitelist cobra command.
func CmdTxClassUnwhitelist() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "class-unwhitelist [class-id] [account] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Unwhitelist an account for a class of non-fungible tokens",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unwhitelist an account for a class of non-fungible tokens.

Example:
$ %s tx %s class-unwhitelist abc-%[3]s %[3]s --from [sender]
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
			classID := args[0]
			account := args[1]

			msg := &types.MsgRemoveFromClassWhitelist{
				Sender:  sender.String(),
				ClassID: classID,
				Account: account,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxClassFreeze returns ClassFreeze cobra command.
func CmdTxClassFreeze() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "class-freeze [class-id] [account] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Freeze an account for a class of non-fungible tokens",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Freeze an account for a class of non-fungible tokens.

Example:
$ %s tx %s class-freeze abc-%[3]s %[3]s --from [sender]
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
			classID := args[0]
			account := args[1]

			msg := &types.MsgClassFreeze{
				Sender:  sender.String(),
				ClassID: classID,
				Account: account,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxClassUnfreeze returns ClassUnfreeze cobra command.
func CmdTxClassUnfreeze() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "class-unfreeze [class-id] [account] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Unfreeze an account for a class of non-fungible tokens",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unfreeze an account for a class of non-fungible tokens.

Example:
$ %s tx %s class-unfreeze abc-%[3]s %[3]s --from [sender]
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
			classID := args[0]
			account := args[1]

			msg := &types.MsgClassUnfreeze{
				Sender:  sender.String(),
				ClassID: classID,
				Account: account,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdGrantAuthorization returns a CLI command handler for creating a MsgGrant transaction.
func CmdGrantAuthorization() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant [grantee] [message_type=\"send\"] --from <granter> --auth-file=path/to/authz.json",
		Short: "Grant authorization to an address",
		Long: fmt.Sprintf(`Grant authorization to an address.
Examples:
$ %s tx grant <grantee_addr> send --expiration 1667979596 --auth-file=./authz.json

Where authz.json for send grant contains:

{
	"nfts":[
		{
			"class_id":"class1-%[3]s",
			"id": "nft-id-1"
		}
	]
}
`, version.AppName, version.AppName, constant.AddressSampleTest),
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			grantee, err := sdk.AccAddressFromBech32(args[0])
			if err != nil {
				return err
			}

			expire, err := getExpireTime(cmd)
			if err != nil {
				return err
			}

			var authorization authz.Authorization
			switch args[1] {
			case "send":
				path, err := cmd.Flags().GetString(AuthzFileFlag)
				if err != nil {
					return err
				}

				contents, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				authorization = &types.SendAuthorization{}
				err = json.Unmarshal(contents, authorization)
				if err != nil {
					return err
				}
			default:
				return errors.Errorf("invalid authorization types, %s", args[1])
			}

			grantMsg, err := authz.NewMsgGrant(clientCtx.GetFromAddress(), grantee, authorization, expire)
			if err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), grantMsg)
		},
	}
	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().Int64(ExpirationFlag, 0, "Expire time as Unix timestamp. Set zero (0) for no expiry.")
	cmd.Flags().String(AuthzFileFlag, "", "path to the file containing auth content.")
	return cmd
}

func getExpireTime(cmd *cobra.Command) (*time.Time, error) {
	exp, err := cmd.Flags().GetInt64(ExpirationFlag)
	if err != nil {
		return nil, err
	}
	if exp == 0 {
		return nil, nil //nolint:nilnil //the intent of this function is to simplify return nil time.
	}
	e := time.Unix(exp, 0)
	return &e, nil
}

func getDataFromFile(cmd *cobra.Command) (*codectypes.Any, error) {
	if !cmd.Flags().Changed(DataFileFlag) {
		return nil, nil //nolint:nilnil
	}

	dataFilePath, err := cmd.Flags().GetString(DataFileFlag)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	data, err := os.ReadFile(dataFilePath)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// FIXME update mint CLI
	// FIXME add update CLI

	dataAny, err := codectypes.NewAnyWithValue(&types.DataBytes{Data: data})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return dataAny, nil
}
