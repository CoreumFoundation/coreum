package config

import (
	"context"
	"encoding/json"
	"time"

	sdkmath "cosmossdk.io/math"
	cometbftcrypto "github.com/cometbft/cometbft/crypto"
	"github.com/cometbft/cometbft/types"
	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/tx"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/v4/pkg/config/constant"
	assetfttypes "github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
	customparamstypes "github.com/CoreumFoundation/coreum/v4/x/customparams/types"
)

// GenesisInitConfig is used to pass genesis creating paramers to cored.
//
//nolint:tagliatelle
type GenesisInitConfig struct {
	ChainID            constant.ChainID              `json:"chain_id"`
	Denom              string                        `json:"denom"`
	DisplayDenom       string                        `json:"display_denom"`
	AddressPrefix      string                        `json:"address_prefix"`
	GenesisTime        time.Time                     `json:"genesis_time"`
	GovConfig          GenesisInitGovConfig          `json:"gov_config"`
	CustomParamsConfig GenesisInitCustomParamsConfig `json:"custom_params_config"`
	BankBalances       []banktypes.Balance           `json:"bank_balances"`
	Validators         []GenesisInitValidator        `json:"validators"`
}

// GenesisInitGovConfig is the gov config of the GenesisInitConfig.
//
//nolint:tagliatelle
type GenesisInitGovConfig struct {
	MinDeposit            sdk.Coins     `json:"min_deposit"`
	ExpeditedMinDeposit   sdk.Coins     `json:"expedited_min_deposit"`
	VotingPeriod          time.Duration `json:"voting_period"`
	ExpeditedVotingPeriod time.Duration `json:"expedited_voting_period"`
}

// GenesisInitCustomParamsConfig is the custom params config of the GenesisInitConfig.
//
//nolint:tagliatelle
type GenesisInitCustomParamsConfig struct {
	MinSelfDelegation sdkmath.Int `json:"min_self_delegation"`
}

// GenesisInitValidator is the validator config of the GenesisInitConfig.
//
//nolint:tagliatelle
type GenesisInitValidator struct {
	DelegatorMnemonic string                `json:"delegator_mnemonic"`
	Pubkey            cometbftcrypto.PubKey `json:"pub_key"`
	ValidatorName     string                `json:"validator_name"`
}

// GenDocFromInput generates genesis doc from genesis init config.
//
//nolint:funlen
func GenDocFromInput(
	ctx context.Context,
	cfg GenesisInitConfig,
	cosmosClientCtx cosmosclient.Context,
	basicManager module.BasicManager,
) (types.GenesisDoc, error) {
	cdc := cosmosClientCtx.Codec
	appGenState := basicManager.DefaultGenesis(cdc)

	// set gov config
	govGenesis := govtypesv1.DefaultGenesisState()
	fourteenDays := 14 * 24 * time.Hour
	govGenesis.Params.MaxDepositPeriod = &fourteenDays
	govGenesis.Params.BurnVoteQuorum = true
	govGenesis.Params.ProposalCancelRatio = "1.0"
	if len(cfg.GovConfig.MinDeposit) > 0 {
		govGenesis.Params.MinDeposit = cfg.GovConfig.MinDeposit
	}
	if len(cfg.GovConfig.ExpeditedMinDeposit) > 0 {
		govGenesis.Params.ExpeditedMinDeposit = cfg.GovConfig.ExpeditedMinDeposit
	}
	if cfg.GovConfig.VotingPeriod > 0 {
		govGenesis.Params.VotingPeriod = &cfg.GovConfig.VotingPeriod
	}
	if cfg.GovConfig.ExpeditedVotingPeriod > 0 {
		govGenesis.Params.ExpeditedVotingPeriod = &cfg.GovConfig.ExpeditedVotingPeriod
	}
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
	assetftGenesis.Params.IssueFee.Amount = sdkmath.NewInt(10_000_000)
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
	crisisGenesis.ConstantFee.Amount = sdkmath.NewInt(500_000_000_000)
	appGenState[crisistypes.ModuleName] = cdc.MustMarshalJSON(crisisGenesis)

	// distribution params
	distributionGenesis := distributiontypes.DefaultGenesisState()
	distributionGenesis.Params.CommunityTax = sdkmath.LegacyMustNewDecFromStr("0.050000000000000000")
	appGenState[distributiontypes.ModuleName] = cdc.MustMarshalJSON(distributionGenesis)

	// mint params
	mintGenesis := minttypes.DefaultGenesisState()
	mintGenesis.Params.BlocksPerYear = 17900000
	mintGenesis.Params.InflationMin = sdkmath.LegacyMustNewDecFromStr("0")
	mintGenesis.Params.InflationMax = sdkmath.LegacyMustNewDecFromStr("0.20")
	mintGenesis.Params.InflationRateChange = sdkmath.LegacyMustNewDecFromStr("0.13")
	mintGenesis.Minter.Inflation = sdkmath.LegacyMustNewDecFromStr("0.1")

	appGenState[minttypes.ModuleName] = cdc.MustMarshalJSON(mintGenesis)

	// slashing params
	slashingGenesis := slashingtypes.DefaultGenesisState()
	slashingGenesis.Params.DowntimeJailDuration = 60 * time.Second
	slashingGenesis.Params.SignedBlocksWindow = 34000
	slashingGenesis.Params.SlashFractionDowntime = sdkmath.LegacyMustNewDecFromStr("0.005")

	appGenState[slashingtypes.ModuleName] = cdc.MustMarshalJSON(slashingGenesis)

	// staking params
	stakingGenesis := stakingtypes.DefaultGenesisState()
	stakingGenesis.Params.MaxValidators = 32
	stakingGenesis.Params.UnbondingTime = 168 * time.Hour

	appGenState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(stakingGenesis)

	// genutil state
	genutilState := genutiltypes.DefaultGenesisState()
	for _, validatorInfo := range cfg.Validators {
		pubKey, err := cryptocodec.FromCmtPubKeyInterface(validatorInfo.Pubkey)
		if err != nil {
			return types.GenesisDoc{}, errors.Wrapf(err, "error converting public key, err:%s", err)
		}

		genTx, err := signedCreateValidatorTxBytes(
			ctx,
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
	consensusParams.Block.MaxBytes = 6_291_456
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
	ctx context.Context,
	clientCtx cosmosclient.Context,
	chainID string,
	validatorName string,
	validatorPubKey cryptotypes.PubKey,
	mnemonic string,
) ([]byte, error) {
	const signerKeyName = "signer"
	clientCtx = clientCtx.WithFrom(signerKeyName)
	inMemKeyring := keyring.NewInMemory(NewEncodingConfig().Codec)
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
		Rate:          sdkmath.LegacyMustNewDecFromStr("0.1"),
		MaxRate:       sdkmath.LegacyMustNewDecFromStr("0.2"),
		MaxChangeRate: sdkmath.LegacyMustNewDecFromStr("0.01"),
	}

	msg, err := stakingtypes.NewMsgCreateValidator(
		sdk.ValAddress(stakerAddress).String(),
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
	if err := tx.Sign(ctx, txf, signerKeyName, txBuilder, true); err != nil {
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

	// configure funded accounts
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
