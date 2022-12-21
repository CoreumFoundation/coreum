package cosmoscmd

import (
	"github.com/cosmos/cosmos-sdk/client/flags"
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
	if flag := cmd.LocalFlags().Lookup(flags.FlagGasPrices); flag != nil && (!flag.Changed || flag.Value.String() == autoValue) {
		gasPrice, err := feecli.QueryGasPrice(cmd)
		if err != nil {
			return err
		}

		if err = flag.Value.Set(gasPrice.MinGasPrice.String()); err != nil {
			return err
		}
	}
	return nil
}

func addQueryGasPriceToAllLeafs(cmd *cobra.Command) {
	if cmd.Run != nil || cmd.RunE != nil {
		cmd.PreRunE = mergeRunEs(queryGasPriceRunE, cmd.PreRunE)
		return
	}

	for _, cmd := range cmd.Commands() {
		addQueryGasPriceToAllLeafs(cmd)
	}
}
