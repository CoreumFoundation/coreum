package config

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/samber/lo"

	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/x/asset/nft/types"
	nfttypes "github.com/CoreumFoundation/coreum/x/nft"
)

type gasByMsgFunc = func(msg sdk.Msg) (uint64, bool)

// DeterministicGasRequirements specifies gas required by all transaction types
// Crisis module is intentionally skipped here because it is already deterministic by design and fee is specified
// using `consume_fee` param in genesis.
type DeterministicGasRequirements struct {
	FixedGas uint64

	freeBytes      uint64
	freeSignatures uint64

	gasByMsg map[string]gasByMsgFunc
}

// DefaultDeterministicGasRequirements returns default config for deterministic gas.
func DefaultDeterministicGasRequirements() DeterministicGasRequirements {
	dgr := DeterministicGasRequirements{
		FixedGas:       50000,
		freeBytes:      2048,
		freeSignatures: 1,
	}

	dgr.gasByMsg = map[string]gasByMsgFunc{
		// asset/ft
		MsgName(&assetfttypes.MsgIssue{}):               constantGasFunc(80000),
		MsgName(&assetfttypes.MsgMint{}):                constantGasFunc(35000),
		MsgName(&assetfttypes.MsgBurn{}):                constantGasFunc(35000),
		MsgName(&assetfttypes.MsgFreeze{}):              constantGasFunc(55000),
		MsgName(&assetfttypes.MsgUnfreeze{}):            constantGasFunc(55000),
		MsgName(&assetfttypes.MsgGloballyFreeze{}):      constantGasFunc(5000),
		MsgName(&assetfttypes.MsgGloballyUnfreeze{}):    constantGasFunc(5000),
		MsgName(&assetfttypes.MsgSetWhitelistedLimit{}): constantGasFunc(35000),

		// asset/nft
		MsgName(&assetnfttypes.MsgIssueClass{}): constantGasFunc(20000),
		MsgName(&assetnfttypes.MsgMint{}):       constantGasFunc(30000),

		// authz
		MsgName(&authz.MsgExec{}):   authzMsgExecGasFunc(2000, &dgr),
		MsgName(&authz.MsgGrant{}):  constantGasFunc(7000),
		MsgName(&authz.MsgRevoke{}): constantGasFunc(7000),

		// bank
		MsgName(&banktypes.MsgSend{}):      bankSendMsgGasFunc(22000),
		MsgName(&banktypes.MsgMultiSend{}): bankMultiSendMsgGasFunc(27000),

		// distribution
		MsgName(&distributiontypes.MsgFundCommunityPool{}):           constantGasFunc(50000),
		MsgName(&distributiontypes.MsgSetWithdrawAddress{}):          constantGasFunc(50000),
		MsgName(&distributiontypes.MsgWithdrawDelegatorReward{}):     constantGasFunc(120000),
		MsgName(&distributiontypes.MsgWithdrawValidatorCommission{}): constantGasFunc(50000),

		// gov
		MsgName(&govtypes.MsgSubmitProposal{}): constantGasFunc(95000),
		MsgName(&govtypes.MsgVote{}):           constantGasFunc(8000),
		MsgName(&govtypes.MsgVoteWeighted{}):   constantGasFunc(11000),
		MsgName(&govtypes.MsgDeposit{}):        constantGasFunc(11000),

		// nft
		MsgName(&nfttypes.MsgSend{}): constantGasFunc(20000),

		// slashing
		MsgName(&slashingtypes.MsgUnjail{}): constantGasFunc(25000),

		// staking
		MsgName(&stakingtypes.MsgDelegate{}):        constantGasFunc(51000),
		MsgName(&stakingtypes.MsgUndelegate{}):      constantGasFunc(51000),
		MsgName(&stakingtypes.MsgBeginRedelegate{}): constantGasFunc(51000),
		MsgName(&stakingtypes.MsgCreateValidator{}): constantGasFunc(50000),
		MsgName(&stakingtypes.MsgEditValidator{}):   constantGasFunc(50000),

		// wasm
		// TODO (milad): rewise gas config for WASM msgs.
		MsgName(&wasmtypes.MsgStoreCode{}):            underministicGasFunc(),
		MsgName(&wasmtypes.MsgInstantiateContract{}):  underministicGasFunc(),
		MsgName(&wasmtypes.MsgInstantiateContract2{}): underministicGasFunc(),
		MsgName(&wasmtypes.MsgExecuteContract{}):      underministicGasFunc(),
		MsgName(&wasmtypes.MsgMigrateContract{}):      underministicGasFunc(),
		MsgName(&wasmtypes.MsgUpdateAdmin{}):          underministicGasFunc(),
		MsgName(&wasmtypes.MsgClearAdmin{}):           underministicGasFunc(),
		MsgName(&wasmtypes.MsgIBCSend{}):              underministicGasFunc(),
		MsgName(&wasmtypes.MsgIBCCloseChannel{}):      underministicGasFunc(),
	}

	return dgr
}

// TxBaseGas is the free gas we give to every transaction to cover costs of
// tx size and signature verification. TxBaseGas is covered by FixedGas.
func (dgr DeterministicGasRequirements) TxBaseGas(params authtypes.Params) uint64 {
	return dgr.freeBytes*params.TxSizeCostPerByte + dgr.freeSignatures*params.SigVerifyCostSecp256k1
}

// GasRequiredByMessage returns gas required by message and true if message is deterministic.
// Function returns 0 and false if message is undeterministic or unknown.
func (dgr DeterministicGasRequirements) GasRequiredByMessage(msg sdk.Msg) (uint64, bool) {
	gasFunc, ok := dgr.gasByMsg[MsgName(msg)]
	if ok {
		return gasFunc(msg)
	}
	// Unknown message.
	return 0, false
}

// MsgName returns TypeURL of a msg in cosmos SDK style.
// Samples of values returned by the function:
// "/cosmos.distribution.v1beta1.MsgFundCommunityPool"
// "/coreum.asset.ft.v1.MsgMint"
func MsgName(msg sdk.Msg) string {
	return sdk.MsgTypeURL(msg)
}

func constantGasFunc(constGasVal uint64) gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		return constGasVal, true
	}
}

func underministicGasFunc() gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		return 0, false
	}
}

// NOTE: we need to pass DeterministicGasRequirements by pointer here because
// it needs to be initialized later map with all msg types inside to estimate gas recursively.
func authzMsgExecGasFunc(authzMsgExecOverhead uint64, dgr *DeterministicGasRequirements) gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		m, ok := msg.(*authz.MsgExec)
		if !ok {
			return 0, false
		}

		totalGas := authzMsgExecOverhead
		childMsgs, err := m.GetMessages()
		if err != nil {
			return 0, false
		}
		for _, childMsg := range childMsgs {
			gas, isDeterministic := dgr.GasRequiredByMessage(childMsg)
			if !isDeterministic {
				return 0, false
			}
			totalGas += gas
		}
		return totalGas, true
	}
}

func bankSendMsgGasFunc(bankSendPerEntryGas uint64) gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		m, ok := msg.(*banktypes.MsgSend)
		if !ok {
			return 0, false
		}
		entriesNum := len(m.Amount)

		return uint64(lo.Max([]int{entriesNum, 1})) * bankSendPerEntryGas, true
	}
}

func bankMultiSendMsgGasFunc(bankMultiSendPerEntryGas uint64) gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		m, ok := msg.(*banktypes.MsgMultiSend)
		if !ok {
			return 0, false
		}
		inputEntriesNum := 0
		for _, inp := range m.Inputs {
			inputEntriesNum += len(inp.Coins)
		}

		outputEntriesNum := 0
		for _, outp := range m.Outputs {
			outputEntriesNum += len(outp.Coins)
		}

		// Select max of input or output entries & use 1 as a fallback.
		maxEntriesNum := lo.Max([]int{inputEntriesNum, outputEntriesNum, 1})
		return uint64(maxEntriesNum) * bankMultiSendPerEntryGas, true
	}
}
