package cosmoscmd

import (
	"fmt"
	"os"

	cmtjson "github.com/cometbft/cometbft/libs/json"
	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/v6/pkg/config"
)

const (
	// FlagOutputPath defines an output path.
	FlagOutputPath = "output-path"

	// FlagInputPath defines an input path.
	FlagInputPath = "input-path"

	// FlagValidatorName defines a name of the validator.
	FlagValidatorName = "validator-name"
)

// GenerateGenesisCmd returns a cobra command that generates the gensis file, given an input config.
func GenerateGenesisCmd(basicManager module.BasicManager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-genesis",
		Short: "Generate gensis file",
		Long:  `Generate gensis file, which can be modified via input config file`,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			cosmosClientCtx := cosmosclient.GetClientContextFromCmd(cmd)

			inputPath, err := cmd.Flags().GetString(FlagInputPath)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to read %s flag", FlagInputPath))
			}

			inputContent, err := os.ReadFile(inputPath)
			if err != nil {
				return errors.Wrap(err, "failed to read file "+inputPath)
			}

			var genCfg config.GenesisInitConfig
			if err := cmtjson.Unmarshal(inputContent, &genCfg); err != nil {
				return errors.Wrap(err, fmt.Sprintf("error parsing input file, err: %s", err))
			}

			if genCfg.Denom != "" {
				sdk.DefaultBondDenom = genCfg.Denom
			}

			genDoc, err := config.GenDocFromInput(ctx, genCfg, cosmosClientCtx, basicManager)
			if err != nil {
				return err
			}

			outputPath, err := cmd.Flags().GetString(FlagOutputPath)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to read %s flag", FlagOutputPath))
			}

			return genDoc.SaveAs(outputPath)
		},
	}

	cmd.Flags().String(FlagOutputPath, "", "file path for the generated genesis file")
	cmd.Flags().String(FlagInputPath, "", "file path for the input config file")
	cmd.Flags().StringArray(FlagValidatorName, []string{}, "list of the validator names to generate")

	return cmd
}
