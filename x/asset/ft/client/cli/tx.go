package cli

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

// Flags defined on transactions.
const (
	FeaturesFlag           = "features"
	BurnRateFlag           = "burn-rate"
	SendCommissionRateFlag = "send-commission-rate"
	IBCEnabledFlag         = "ibc-enabled"
	MintLimitFlag          = "mint-limit"
	BurnLimitFlag          = "burn-limit"
	ExpirationFlag         = "expiration"
	RecipientFlag          = "recipient"
	URIFlag                = "uri"
	URIHashFlag            = "uri_hash"
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
		CmdTxIssue(),
		CmdTxMint(),
		CmdTxBurn(),
		CmdTxFreeze(),
		CmdTxUnfreeze(),
		CmdTxSetFrozen(),
		CmdTxGloballyFreeze(),
		CmdTxGloballyUnfreeze(),
		CmdTxClawback(),
		CmdTxSetWhitelistedLimit(),
		CmdTxUpgradeV1(),
		CmdGrantAuthorization(),
	)

	return cmd
}

// CmdTxIssue returns Issue cobra command.
//
//nolint:funlen // Despite the length function is still manageable
func CmdTxIssue() *cobra.Command {
	var allowedFeatures []string
	for _, n := range types.Feature_name {
		allowedFeatures = append(allowedFeatures, n)
	}
	sort.Strings(allowedFeatures)
	cmd := &cobra.Command{
		//nolint:lll // breaking this down will make it look worse when printed to user screen.
		Use:   "issue [symbol] [subunit] [precision] [initial_amount] [description] --from [issuer] --features=" + strings.Join(allowedFeatures, ",") + " --burn-rate=0.12 --send-commission-rate=0.2 --uri https://my-token-meta.invalid/1 --uri_hash e000624",
		Args:  cobra.ExactArgs(5),
		Short: "Issue new fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Issues new fungible token.

Example:
$ %s tx %s issue WBTC wsatoshi 8 100000 "Wrapped Bitcoin Token" --from [issuer]
`,
				version.AppName, types.ModuleName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			issuer := clientCtx.GetFromAddress()
			symbol := args[0]
			subunit := args[1]
			precision, err := strconv.ParseUint(args[2], 10, 32)
			if err != nil {
				return sdkerrors.Wrap(err, "invalid precision")
			}

			// if the initial amount wasn't provided the amount is zero
			initialAmount := sdkmath.ZeroInt()
			if args[3] != "" {
				var ok bool
				initialAmount, ok = sdkmath.NewIntFromString(args[3])
				if !ok {
					return sdkerrors.Wrap(err, "invalid initial_amount")
				}
			}

			featuresString, err := cmd.Flags().GetStringSlice(FeaturesFlag)
			if err != nil {
				return errors.WithStack(err)
			}

			burnRate := sdk.NewDec(0)
			burnRateStr, err := cmd.Flags().GetString(BurnRateFlag)
			if err != nil {
				return errors.WithStack(err)
			}
			if len(burnRateStr) > 0 {
				burnRate, err = sdk.NewDecFromStr(burnRateStr)
				if err != nil {
					return errors.Wrapf(err, "invalid burn-rate")
				}
			}

			sendCommissionRate := sdk.NewDec(0)
			sendCommissionFeeStr, err := cmd.Flags().GetString(SendCommissionRateFlag)
			if err != nil {
				return errors.WithStack(err)
			}
			if len(sendCommissionFeeStr) > 0 {
				sendCommissionRate, err = sdk.NewDecFromStr(sendCommissionFeeStr)
				if err != nil {
					return errors.Wrapf(err, "invalid send-commission-rate")
				}
			}

			var features []types.Feature
			for _, str := range featuresString {
				feature, ok := types.Feature_value[str]
				if !ok {
					return errors.Errorf("unknown feature '%s',allowed features: %s", str, strings.Join(allowedFeatures, ","))
				}
				features = append(features, types.Feature(feature))
			}
			description := args[4]

			uri, err := cmd.Flags().GetString(URIFlag)
			if err != nil {
				return errors.WithStack(err)
			}
			uriHash, err := cmd.Flags().GetString(URIHashFlag)
			if err != nil {
				return errors.WithStack(err)
			}

			msg := &types.MsgIssue{
				Issuer:             issuer.String(),
				Symbol:             symbol,
				Subunit:            subunit,
				Precision:          uint32(precision),
				InitialAmount:      initialAmount,
				Description:        description,
				Features:           features,
				BurnRate:           burnRate,
				SendCommissionRate: sendCommissionRate,
				URI:                uri,
				URIHash:            uriHash,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	//nolint:lll // breaking this down will make it look worse when printed to user screen.
	cmd.Flags().StringSlice(FeaturesFlag, []string{}, "Features to be enabled on fungible token. e.g --features="+strings.Join(allowedFeatures, ","))
	//nolint:lll // breaking this down will make it look worse when printed to user screen.
	cmd.Flags().String(BurnRateFlag, "0", "Indicates the rate at which coins will be burnt on top of the sent amount in every send action. Must be between 0 and 1.")
	//nolint:lll // breaking this down will make it look worse when printed to user screen.
	cmd.Flags().String(SendCommissionRateFlag, "0", "Indicates the rate at which coins will be sent to the issuer on top of the sent amount in every send action. Must be between 0 and 1.")
	cmd.Flags().String(URIFlag, "", "Token URI.")
	cmd.Flags().String(URIHashFlag, "", "Token URI hash.")

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxMint returns Mint cobra command.
func CmdTxMint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mint [amount] --from [sender] --recipient [recipient]",
		Args:  cobra.ExactArgs(1),
		Short: "mint new amount of fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Mint new amount of fungible token.

Example:
$ %s tx %s mint 100000ABC-%s --from [sender]
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

			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid amount")
			}

			msg := &types.MsgMint{
				Sender:    sender.String(),
				Recipient: recipient,
				Coin:      amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(
		RecipientFlag,
		"",
		"Address to send minted tokens to, if not specified minted tokens are sent to the issuer",
	)

	return cmd
}

// CmdTxBurn returns Burn cobra command.
func CmdTxBurn() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "burn [amount] --from [sender]",
		Args:  cobra.ExactArgs(1),
		Short: "burn some amount of fungible token",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Burn some amount of fungible token.

Example:
$ %s tx %s burn 100000ABC-%s --from [sender]
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
			amount, err := sdk.ParseCoinNormalized(args[0])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid amount")
			}

			msg := &types.MsgBurn{
				Sender: sender.String(),
				Coin:   amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxFreeze returns Freeze cobra command.
//
//nolint:dupl // most code is identical between Freeze/Unfreeze cmd, but reusing logic is not beneficial here.
func CmdTxFreeze() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "freeze [account_address] [amount] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Freeze any amount of fungible token for the specific account",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Freeze a portion of fungible token.

Example:
$ %s tx %s freeze [account_address] 100000ABC-%s --from [sender]
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
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid amount")
			}

			msg := &types.MsgFreeze{
				Sender:  sender.String(),
				Account: account,
				Coin:    amount,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxUnfreeze returns Unfreeze cobra command.
//
//nolint:dupl // most code is identical between Freeze/Unfreeze cmd, but reusing logic is not beneficial here.
func CmdTxUnfreeze() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unfreeze [account_address] [amount] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Unfreeze a portion of the frozen fungible tokens",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unfreezes a portion of the frozen fungible token.

Example:
$ %s tx %s unfreeze [account_address] 100000ABC-%s --from [sender]
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
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid amount")
			}

			msg := &types.MsgUnfreeze{
				Sender:  sender.String(),
				Account: account,
				Coin:    amount,
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxSetFrozen returns SetFrozen cobra command.
//
//nolint:dupl // most code is identical between Freeze/Unfreeze cmd, but reusing logic is not beneficial here.
func CmdTxSetFrozen() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-frozen [account_address] [amount] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Set absolute frozen amount for the specific account",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Set absolute frozen amount for the specific account.

Example:
$ %s tx %s set-frozen [account_address] 100000ABC-%s --from [sender]
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
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid amount")
			}

			msg := &types.MsgSetFrozen{
				Sender:  sender.String(),
				Account: account,
				Coin:    amount,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxClawback returns Clawback cobra command.
func CmdTxClawback() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clawback [account_address] [amount] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Confiscates any amount of fungible token from the specific account",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Confiscate a portion of fungible token.

Example:
$ %s tx %s clawback [account_address] 100000ABC-%s --from [sender]
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
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid amount")
			}

			msg := &types.MsgClawback{
				Sender:  sender.String(),
				Account: account,
				Coin:    amount,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxSetWhitelistedLimit returns SetWhitelistedLimit cobra command.
//
//nolint:dupl // most code is identical, but reusing logic is not beneficial here.
func CmdTxSetWhitelistedLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-whitelisted-limit [account_address] [amount] --from [sender]",
		Args:  cobra.ExactArgs(2),
		Short: "Set whitelisted limit on an account",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Set whitelisted limit on an account.

Example:
$ %s tx %s set-whitelisted-limit [account_address] 100000ABC-%s --from [sender]
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
			amount, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return sdkerrors.Wrap(err, "invalid amount")
			}

			msg := &types.MsgSetWhitelistedLimit{
				Sender:  sender.String(),
				Account: account,
				Coin:    amount,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxGloballyFreeze returns GlobalFreeze cobra command.
func CmdTxGloballyFreeze() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "globally-freeze [denom] --from [sender]",
		Args:  cobra.ExactArgs(1),
		Short: "freezes fungible token so no operations are allowed with it before unfrozen",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Freezes fungible token so no operations are allowed with it before unfrozen.
This operation is idempotent so global freeze of already frozen token does nothing.

Example:
$ %s tx %s globally-freeze ABC-%s --from [sender]
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
			denom := args[0]

			msg := &types.MsgGloballyFreeze{
				Sender: sender.String(),
				Denom:  denom,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxGloballyUnfreeze returns GlobalUnfreeze cobra command.
func CmdTxGloballyUnfreeze() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "globally-unfreeze [denom] --from [sender]",
		Args:  cobra.ExactArgs(1),
		Short: "unfreezes fungible token and unblocks basic operations on it",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Unfreezes fungible token and unblocks basic operations on it.
This operation is idempotent so global unfreezing of non-frozen token does nothing.

Example:
$ %s tx %s globally-unfreeze ABC-%s --from [sender]
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
			denom := args[0]

			msg := &types.MsgGloballyUnfreeze{
				Sender: sender.String(),
				Denom:  denom,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdTxUpgradeV1 returns UpgradeV1 cobra command.
func CmdTxUpgradeV1() *cobra.Command {
	var ibcEnabled bool
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("upgrade-v1 [denom] --%s=true --from [sender]", IBCEnabledFlag),
		Args:  cobra.ExactArgs(1),
		Short: "upgrades denom to version v1 and specifies if IBC should be enabled or disabled",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Upgrades denom to version v1 and specifies if IBC should be enabled or disabled.
This is a one-time operation!!! Once executed, it can never be done again.

Example:
$ %s tx %s upgrade-v1 ABC-%s --%s=true --from [sender]
`,
				version.AppName, types.ModuleName, constant.AddressSampleTest, IBCEnabledFlag,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed(IBCEnabledFlag) {
				return errors.Errorf("flag --%s must be explicitly set", IBCEnabledFlag)
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return errors.WithStack(err)
			}

			sender := clientCtx.GetFromAddress()
			denom := args[0]

			msg := &types.MsgUpgradeTokenV1{
				Sender:     sender.String(),
				Denom:      denom,
				IbcEnabled: ibcEnabled,
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
	cmd.Flags().BoolVar(&ibcEnabled, IBCEnabledFlag, false, "Specifies if IBC should be enabled or disabled for the token")
	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

// CmdGrantAuthorization returns a CLI command handler for creating a MsgGrant transaction.
func CmdGrantAuthorization() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant [grantee] [message_type=\"mint\"|\"burn\"] --from <granter> --burn-limit 10ucore --mint-limit 10ucore",
		Short: "Grant authorization to an address",
		Long: fmt.Sprintf(`Grant authorization to an address.
Examples:
$ %s tx grant <grantee_addr> mint --mint-limit 100ucore --expiration 1667979596

$ %s tx grant <grantee_addr> burn --burn-limit 100ucore --expiration 1667979596
`, version.AppName, version.AppName),
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

			var authorization authz.Authorization
			switch args[1] {
			case "mint":
				limit, err := cmd.Flags().GetString(MintLimitFlag)
				if err != nil {
					return err
				}

				limitCoins, err := sdk.ParseCoinsNormalized(limit)
				if err != nil {
					return err
				}

				if !limitCoins.IsAllPositive() {
					return fmt.Errorf("mint-limit should be greater than zero")
				}
				authorization = types.NewMintAuthorization(limitCoins)
			case "burn":
				limit, err := cmd.Flags().GetString(BurnLimitFlag)
				if err != nil {
					return err
				}

				limitCoins, err := sdk.ParseCoinsNormalized(limit)
				if err != nil {
					return err
				}

				if !limitCoins.IsAllPositive() {
					return fmt.Errorf("burn-limit should be greater than zero")
				}
				authorization = types.NewBurnAuthorization(limitCoins)
			default:
				return errors.Errorf("invalid authorization types, %s", args[1])
			}

			expire, err := getExpireTime(cmd)
			if err != nil {
				return err
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
	cmd.Flags().String(BurnLimitFlag, "", "The Amount that is allowed to be burnt.")
	cmd.Flags().String(MintLimitFlag, "", "The Amount that is allowed to be minted.")
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
