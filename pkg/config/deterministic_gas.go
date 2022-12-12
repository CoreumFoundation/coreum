package config

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

// DefaultDeterministicGasRequirements returns default config for deterministic gas
func DefaultDeterministicGasRequirements() DeterministicGasRequirements {
	return DeterministicGasRequirements{
		FixedGas:       50000,
		FreeBytes:      2048,
		FreeSignatures: 1,

		AssetIssueFungibleToken:               80000,
		AssetMintFungibleToken:                35000,
		AssetBurnFungibleToken:                35000,
		AssetFreezeFungibleToken:              55000,
		AssetUnfreezeFungibleToken:            55000,
		AssetGloballyFreezeFungibleToken:      5000,
		AssetGloballyUnfreezeFungibleToken:    5000,
		AssetSetWhitelistedLimitFungibleToken: 35000,

		BankSendPerEntry:      22000,
		BankMultiSendPerEntry: 27000,

		DistributionFundCommunityPool:           50000,
		DistributionSetWithdrawAddress:          50000,
		DistributionWithdrawDelegatorReward:     120000,
		DistributionWithdrawValidatorCommission: 50000,

		GovSubmitProposal: 95000,
		GovVote:           8000,
		GovVoteWeighted:   11000,
		GovDeposit:        91000,

		StakingDelegate:        51000,
		StakingUndelegate:      51000,
		StakingBeginRedelegate: 51000,
		StakingCreateValidator: 50000,
		StakingEditValidator:   50000,
	}
}

// DeterministicGasRequirements specifies gas required by some transaction types
// Crisis module is intentionally skipped here because it is already deterministic by design and fee is specified
// using `consume_fee` param in genesis.
type DeterministicGasRequirements struct {
	// FixedGas is the fixed amount of gas charged on each transaction as a payment for executing ante handler. This includes:
	// - most of the stuff done by ante decorators
	// - `FreeSignatures` secp256k1 signature verifications
	// - `FreeBytes` bytes of transaction
	FixedGas uint64

	// FreeBytes defines how many tx bytes are stored for free (included in `FixedGas` price)
	FreeBytes uint64

	// FreeSignatures defines how many secp256k1 signatures are verified for free (included in `FixedGas` price)
	FreeSignatures uint64

	// x/asset
	AssetIssueFungibleToken               uint64
	AssetMintFungibleToken                uint64
	AssetBurnFungibleToken                uint64
	AssetFreezeFungibleToken              uint64
	AssetUnfreezeFungibleToken            uint64
	AssetGloballyFreezeFungibleToken      uint64
	AssetGloballyUnfreezeFungibleToken    uint64
	AssetSetWhitelistedLimitFungibleToken uint64

	// x/bank
	BankSendPerEntry      uint64
	BankMultiSendPerEntry uint64

	// x/distribution
	DistributionFundCommunityPool           uint64
	DistributionSetWithdrawAddress          uint64
	DistributionWithdrawDelegatorReward     uint64
	DistributionWithdrawValidatorCommission uint64

	// x/gov
	GovSubmitProposal uint64
	GovVote           uint64
	GovVoteWeighted   uint64
	GovDeposit        uint64

	// x/staking
	StakingDelegate        uint64
	StakingUndelegate      uint64
	StakingBeginRedelegate uint64
	StakingCreateValidator uint64
	StakingEditValidator   uint64
}

// GasRequiredByMessage returns gas required by a sdk.Msg.
// If fixed gas is not specified for the message type it returns 0.
//
//nolint:funlen // it doesn't make sense to split entries
func (dgr DeterministicGasRequirements) GasRequiredByMessage(msg sdk.Msg) (uint64, bool) {
	// Following is the list of messages having deterministic gas amount defined.
	// To test the real gas usage return `false` and run an integration test which reports the used gas.
	// Then define a reasonable value for the message and return `true` again.

	switch m := msg.(type) {
	case *assettypes.MsgIssueFungibleToken:
		return dgr.AssetIssueFungibleToken, true
	case *assettypes.MsgFreezeFungibleToken:
		return dgr.AssetFreezeFungibleToken, true
	case *assettypes.MsgUnfreezeFungibleToken:
		return dgr.AssetUnfreezeFungibleToken, true
	case *assettypes.MsgGloballyFreezeFungibleToken:
		return dgr.AssetFreezeFungibleToken, true
	case *assettypes.MsgGloballyUnfreezeFungibleToken:
		return dgr.AssetUnfreezeFungibleToken, true
	case *assettypes.MsgMintFungibleToken:
		return dgr.AssetMintFungibleToken, true
	case *assettypes.MsgBurnFungibleToken:
		return dgr.AssetBurnFungibleToken, true
	case *assettypes.MsgSetWhitelistedLimitFungibleToken:
		return dgr.AssetSetWhitelistedLimitFungibleToken, true
	case *banktypes.MsgSend:
		entriesNum := len(m.Amount)
		if len(m.Amount) == 0 {
			entriesNum = 1
		}
		return uint64(entriesNum) * dgr.BankSendPerEntry, true
	case *banktypes.MsgMultiSend:
		entriesNum := 0
		for _, i := range m.Inputs {
			entriesNum += len(i.Coins)
		}

		var outputEntriesNum int
		for _, o := range m.Outputs {
			outputEntriesNum += len(o.Coins)
		}
		if outputEntriesNum > entriesNum {
			entriesNum = outputEntriesNum
		}

		if entriesNum == 0 {
			entriesNum = 1
		}
		return uint64(entriesNum) * dgr.BankMultiSendPerEntry, true
	case *distributiontypes.MsgFundCommunityPool:
		return dgr.DistributionFundCommunityPool, true
	case *distributiontypes.MsgSetWithdrawAddress:
		return dgr.DistributionSetWithdrawAddress, true
	case *distributiontypes.MsgWithdrawDelegatorReward:
		return dgr.DistributionWithdrawDelegatorReward, true
	case *distributiontypes.MsgWithdrawValidatorCommission:
		return dgr.DistributionWithdrawValidatorCommission, true
	case *govtypes.MsgSubmitProposal:
		return dgr.GovSubmitProposal, true
	case *govtypes.MsgVote:
		return dgr.GovVote, true
	case *govtypes.MsgVoteWeighted:
		return dgr.GovVoteWeighted, true
	case *govtypes.MsgDeposit:
		return dgr.GovDeposit, true
	case *stakingtypes.MsgDelegate:
		return dgr.StakingDelegate, true
	case *stakingtypes.MsgUndelegate:
		return dgr.StakingUndelegate, true
	case *stakingtypes.MsgBeginRedelegate:
		return dgr.StakingBeginRedelegate, true
	case *stakingtypes.MsgCreateValidator:
		return dgr.StakingCreateValidator, true
	case *stakingtypes.MsgEditValidator:
		return dgr.StakingEditValidator, true
	default:
		return 0, false
	}
}

// TxBaseGas is the free gas we give to every transaction to cover costs of tx size and signature verification
func (dgr DeterministicGasRequirements) TxBaseGas(params authtypes.Params) uint64 {
	return dgr.FreeBytes*params.TxSizeCostPerByte + dgr.FreeSignatures*params.SigVerifyCostSecp256k1
}
