package cosmoscmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	sdkmath "cosmossdk.io/math"
	cometbftcrypto "github.com/cometbft/cometbft/crypto"
	cmtjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/types"
	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/CoreumFoundation/coreum/v4/app"
	"github.com/CoreumFoundation/coreum/v4/pkg/config"
	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	customparamstypes "github.com/CoreumFoundation/coreum/v4/x/customparams/types"
)

// GenesisInitConfig is used to pass genesis creating paramers to cored.
//
//nolint:tagliatelle
type GenesisInitConfig struct {
	ChainID       constant.ChainID `json:"chain_id"`
	Denom         string           `json:"denom"`
	DisplayDenom  string           `json:"display_denom"`
	AddressPrefix string           `json:"address_prefix"`
	GenesisTime   time.Time        `json:"genesis_time"`
	GovConfig     struct {
		MinDeposit   sdk.Coins     `json:"min_deposit"`
		VotingPeriod time.Duration `json:"voting_period"`
	} `json:"gov_config"`
	CustomParamsConfig struct {
		MinSelfDelegation sdkmath.Int `json:"min_self_delegation"`
	} `json:"custom_params_config"`
	BankBalances []banktypes.Balance `json:"bank_balances"`
	Validators   []struct {
		DelegatorMnemonic string                `json:"delegator_mnemonic"`
		Pubkey            cometbftcrypto.PubKey `json:"pub_key"`
		ValidatorName     string                `json:"validator_name"`
	} `json:"validators"`
}

// GenerateGenesisCmd returns a cobra command that generates the gensis file, given an input config.
func GenerateGenesisCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate-genesis",
		Short: "Generate gensis file",
		Long:  `Generate gensis file, which can be modified via input config file`,
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			cosmosClientCtx := cosmosclient.GetClientContextFromCmd(cmd)

			inputPath, err := cmd.Flags().GetString(FlagInputPath)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to read %s flag", FlagInputPath))
			}

			inputContent, err := os.ReadFile(inputPath)
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("failed to read file %s", inputPath))
			}

			var genCfg GenesisInitConfig
			if err := cmtjson.Unmarshal(inputContent, &genCfg); err != nil {
				return errors.Wrap(err, fmt.Sprintf("error parsing input file, err: %s", err))
			}

			if genCfg.Denom != "" {
				sdk.DefaultBondDenom = genCfg.Denom
			}

			genDoc, err := genDocFromInput(genCfg, cosmosClientCtx)
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

func genDocFromInput(cfg GenesisInitConfig, cosmosClientCtx cosmosclient.Context) (types.GenesisDoc, error) {
	cdc := cosmosClientCtx.Codec
	appGenState := app.ModuleBasics.DefaultGenesis(cdc)

	// set gov config
	govGenesis := govtypesv1.DefaultGenesisState()
	if len(cfg.GovConfig.MinDeposit) > 0 {
		govGenesis.Params.MinDeposit = cfg.GovConfig.MinDeposit
	}
	if cfg.GovConfig.VotingPeriod > 0 {
		govGenesis.Params.VotingPeriod = &cfg.GovConfig.VotingPeriod
	}
	fourteenDays := 14 * 24 * time.Hour
	govGenesis.Params.MaxDepositPeriod = &fourteenDays
	appGenState[govtypes.ModuleName] = cdc.MustMarshalJSON(govGenesis)

	// set custom params
	customparamsGenesis := customparamstypes.DefaultGenesisState()
	if !cfg.CustomParamsConfig.MinSelfDelegation.IsNil() {
		if cfg.CustomParamsConfig.MinSelfDelegation.IsNegative() {
			return types.GenesisDoc{}, errors.Errorf("min self delegation cannot be negative")
		}
		customparamsGenesis.StakingParams.MinSelfDelegation = cfg.CustomParamsConfig.MinSelfDelegation
	}
	appGenState[customparamstypes.ModuleName] = cdc.MustMarshalJSON(customparamsGenesis)

	// assetft params
	assetftGenesis := assetfttypes.DefaultGenesis()
	assetftGenesis.Params.IssueFee.Amount = sdk.NewInt(10_000_000)
	appGenState[assetfttypes.ModuleName] = cdc.MustMarshalJSON(assetftGenesis)

	// bank and auth params
	authGenesis, bankGenesis, err := defaultAuthAndBankParams(cfg)
	if err != nil {
		return types.GenesisDoc{}, errors.Wrapf(err, "error creating bank and auth genesis, err:%s", err)
	}
	appGenState[banktypes.ModuleName] = cdc.MustMarshalJSON(bankGenesis)
	appGenState[authtypes.ModuleName] = cdc.MustMarshalJSON(authGenesis)

	// crisis params
	crisisGenesis := crisistypes.DefaultGenesisState()
	crisisGenesis.ConstantFee.Amount = sdk.NewInt(500_000_000_000)
	appGenState[crisistypes.ModuleName] = cdc.MustMarshalJSON(crisisGenesis)

	// distribution params
	distributionGenesis := distributiontypes.DefaultGenesisState()
	distributionGenesis.Params.CommunityTax = sdk.MustNewDecFromStr("0.050000000000000000")
	appGenState[distributiontypes.ModuleName] = cdc.MustMarshalJSON(distributionGenesis)

	// mint params
	mintGenesis := minttypes.DefaultGenesisState()
	mintGenesis.Params.BlocksPerYear = 6311520
	mintGenesis.Params.InflationMin = sdk.MustNewDecFromStr("0.07")
	mintGenesis.Params.InflationMax = sdk.MustNewDecFromStr("0.20")
	mintGenesis.Minter.Inflation = sdk.MustNewDecFromStr("0.13")

	appGenState[minttypes.ModuleName] = cdc.MustMarshalJSON(mintGenesis)

	// genutil state
	genutilState := genutiltypes.DefaultGenesisState()
	for _, validatorInfo := range cfg.Validators {
		pubKey, err := cryptocodec.FromTmPubKeyInterface(validatorInfo.Pubkey)
		if err != nil {
			return types.GenesisDoc{}, errors.Wrapf(err, "error converting public key, err:%s", err)
		}

		genTx, err := signedCreateValidatorTxBytes(
			cosmosClientCtx,
			string(cfg.ChainID),
			validatorInfo.ValidatorName,
			pubKey,
			validatorInfo.DelegatorMnemonic,
		)
		if err != nil {
			return types.GenesisDoc{}, errors.Wrapf(err, "failed to create gen tx,err:%s", err)
		}
		genutilState.GenTxs = append(genutilState.GenTxs, genTx)
	}
	genutiltypes.SetGenesisStateInAppState(cosmosClientCtx.Codec, appGenState, genutilState)

	// marshal the app state
	appState, err := json.MarshalIndent(appGenState, "", " ")
	if err != nil {
		return types.GenesisDoc{}, errors.Wrap(err, "failed to marshal default genesis state")
	}

	consensusParams := types.DefaultConsensusParams()
	consensusParams.Block.MaxGas = 50_000_000

	return types.GenesisDoc{
		GenesisTime:     cfg.GenesisTime,
		ChainID:         string(cfg.ChainID),
		InitialHeight:   1,
		AppState:        appState,
		ConsensusParams: consensusParams,
	}, nil
}

func signedCreateValidatorTxBytes(
	clientCtx cosmosclient.Context,
	chainID string,
	validatorName string,
	validatorPubKey cryptotypes.PubKey,
	mnemonic string,
) ([]byte, error) {
	const signerKeyName = "signer"
	clientCtx = clientCtx.WithFrom(signerKeyName)
	inMemKeyring := keyring.NewInMemory(config.NewEncodingConfig(app.ModuleBasics).Codec)
	k, err := inMemKeyring.NewAccount(
		signerKeyName,
		mnemonic,
		"",
		hd.CreateHDPath(sdk.GetConfig().GetCoinType(), 0, 0).String(),
		hd.Secp256k1,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to import account with mnemonic")
	}

	stakerAddress, err := k.GetAddress()
	if err != nil {
		return nil, errors.Wrap(err, "failed get staker address from key")
	}

	stakerSelfDelegationAmount := sdk.NewCoin(constant.DenomDev, sdkmath.NewInt(10_000_000_000_000))
	commission := stakingtypes.CommissionRates{
		Rate:          sdk.MustNewDecFromStr("0.1"),
		MaxRate:       sdk.MustNewDecFromStr("0.2"),
		MaxChangeRate: sdk.MustNewDecFromStr("0.01"),
	}

	msg, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(stakerAddress),
		validatorPubKey,
		stakerSelfDelegationAmount,
		stakingtypes.Description{
			Moniker: validatorName,
		},
		commission,
		stakerSelfDelegationAmount.Amount,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed create MsgCreateValidator transaction")
	}

	txf := tx.Factory{}.
		WithChainID(chainID).
		WithKeybase(inMemKeyring).
		WithTxConfig(clientCtx.TxConfig)
	txBuilder, err := txf.BuildUnsignedTx(msg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to build MsgCreateValidator transaction")
	}
	if err := tx.Sign(txf, signerKeyName, txBuilder, true); err != nil {
		return nil, errors.Wrap(err, "failed to sign MsgCreateValidator transaction")
	}
	return clientCtx.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
}

// defaultAuthAndBankParams creates the genesis state for both auth and bank modules, since
// creating the funded accounts must modify both genesis states.
func defaultAuthAndBankParams(
	cfg GenesisInitConfig,
) (*authtypes.GenesisState, *banktypes.GenesisState, error) {
	// auth params
	authGensis := authtypes.DefaultGenesisState()
	authGensis.Params.SigVerifyCostED25519 = 1000

	// bank params
	bankGenesis := banktypes.DefaultGenesisState()
	bankGenesis.DenomMetadata = []banktypes.Metadata{{
		Description: cfg.DisplayDenom + " coin",
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    cfg.Denom,
				Exponent: 0,
			},
			{
				Denom:    cfg.DisplayDenom,
				Exponent: 6,
			},
		},
		Base:    cfg.Denom,
		Display: cfg.DisplayDenom,
		Name:    cfg.Denom,
		Symbol:  cfg.Denom,
	}}

	// configure funded accoutns
	var accounts authtypes.GenesisAccounts
	bankGenesis.Balances = cfg.BankBalances
	for _, bb := range cfg.BankBalances {
		accountAddress := sdk.MustAccAddressFromBech32(bb.Address)
		accounts = append(accounts, authtypes.NewBaseAccount(accountAddress, nil, 0, 0))
		bankGenesis.Supply = bankGenesis.Supply.Add(bb.Coins...)
	}

	packedAccounts, err := authtypes.PackAccounts(authtypes.SanitizeGenesisAccounts(accounts))
	if err != nil {
		return nil, nil, err
	}
	authGensis.Accounts = packedAccounts

	return authGensis, bankGenesis, nil
}
