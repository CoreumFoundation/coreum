package cosmoscmd

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	feecli "github.com/CoreumFoundation/coreum/x/feemodel/client/cli"
)

const (
	autoValue = "auto"
)

func mergeRunEs(runEs ...func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		for _, runE := range runEs {
			if runE != nil {
				if err := runE(cmd, args); err != nil {
					return err
				}
			}
		}
		return nil
	}
}

func queryGasPriceRunE(cmd *cobra.Command, args []string) error {
	gasPriceFlag := cmd.LocalFlags().Lookup(flags.FlagGasPrices)
	if gasPriceFlag == nil {
		return nil
	}

	if gasPriceFlag.Changed && gasPriceFlag.Value.String() != autoValue {
		return nil
	}

	feeFlag := cmd.LocalFlags().Lookup(flags.FlagFees)
	if feeFlag != nil && feeFlag.Changed {
		// if both fee flag and gas price flag is provided, it is an error
		if gasPriceFlag.Changed {
			return errors.New("cannot provide both fees and gas prices")
		}

		// if only fee flag is provided, we should not query for gas prices
		return nil
	}

	params, err := feecli.QueryParams(cmd)
	if err != nil {
		return err
	}

	gasPrice, err := feecli.QueryGasPrice(cmd)
	if err != nil {
		return err
	}

	gasPriceWithOverhead := sdk.DecCoin{
		Denom:  gasPrice.MinGasPrice.Denom,
		Amount: params.GetParams().Model.InitialGasPrice,
	}
	return gasPriceFlag.Value.Set(gasPriceWithOverhead.String())
}

// addQueryGasPriceToAllLeafs adds the logic to PreRunE function of all leaf commands
// in the tree of the provided command. This function assumes that only the leaf commands
// will contain logic to execute transactions to be executed.
func addQueryGasPriceToAllLeafs(cmd *cobra.Command) {
	if !cmd.HasSubCommands() {
		cmd.PreRunE = mergeRunEs(queryGasPriceRunE, cmd.PreRunE)
		return
	}

	for _, cmd := range cmd.Commands() {
		addQueryGasPriceToAllLeafs(cmd)
	}
}
