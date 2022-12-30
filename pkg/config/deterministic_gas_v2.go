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

type DeterministicGasRequirements struct {
	FixedGas       uint64
	FreeBytes      uint64
	FreeSignatures uint64

	gasByMsg map[string]gasByMsgFunc
}

func DefaultDeterministicGasRequirements() DeterministicGasRequirements {
	gr := DeterministicGasRequirements{
		FixedGas:       50000,
		FreeBytes:      2048,
		FreeSignatures: 1,
	}

	gr.gasByMsg = map[string]gasByMsgFunc{
		// authz
		msgName(&authz.MsgExec{}):   gr.autzMsgExecGasFunc(2000),
		msgName(&authz.MsgGrant{}):  constGasFunc(7000),
		msgName(&authz.MsgRevoke{}): constGasFunc(7000),

		// asset/ft
		msgName(&assetfttypes.MsgIssue{}):               constGasFunc(80000),
		msgName(&assetfttypes.MsgMint{}):                constGasFunc(35000),
		msgName(&assetfttypes.MsgBurn{}):                constGasFunc(35000),
		msgName(&assetfttypes.MsgFreeze{}):              constGasFunc(55000),
		msgName(&assetfttypes.MsgUnfreeze{}):            constGasFunc(55000),
		msgName(&assetfttypes.MsgGloballyFreeze{}):      constGasFunc(5000),
		msgName(&assetfttypes.MsgGloballyUnfreeze{}):    constGasFunc(5000),
		msgName(&assetfttypes.MsgSetWhitelistedLimit{}): constGasFunc(35000),

		// asset/nft
		msgName(&assetnfttypes.MsgIssueClass{}): constGasFunc(20000),
		msgName(&assetnfttypes.MsgMint{}):       constGasFunc(30000),

		// bank
		msgName(&banktypes.MsgSend{}):      bankSendMsgGasFunc(22000),
		msgName(&banktypes.MsgMultiSend{}): bankMultiSendMsgGasFunc(27000),

		// distribution
		msgName(&distributiontypes.MsgFundCommunityPool{}):           constGasFunc(50000),
		msgName(&distributiontypes.MsgSetWithdrawAddress{}):          constGasFunc(50000),
		msgName(&distributiontypes.MsgWithdrawDelegatorReward{}):     constGasFunc(120000),
		msgName(&distributiontypes.MsgWithdrawValidatorCommission{}): constGasFunc(50000),

		// gov
		msgName(&govtypes.MsgSubmitProposal{}): constGasFunc(95000),
		msgName(&govtypes.MsgVote{}):           constGasFunc(8000),
		msgName(&govtypes.MsgVoteWeighted{}):   constGasFunc(11000),
		msgName(&govtypes.MsgDeposit{}):        constGasFunc(11000),

		// nft
		msgName(&nfttypes.MsgSend{}): constGasFunc(20000),

		// slashing
		msgName(&slashingtypes.MsgUnjail{}): constGasFunc(25000),

		// staking
		msgName(&stakingtypes.MsgDelegate{}):        constGasFunc(51000),
		msgName(&stakingtypes.MsgUndelegate{}):      constGasFunc(51000),
		msgName(&stakingtypes.MsgBeginRedelegate{}): constGasFunc(51000),
		msgName(&stakingtypes.MsgCreateValidator{}): constGasFunc(50000),
		msgName(&stakingtypes.MsgEditValidator{}):   constGasFunc(50000),

		// wasm
		msgName(&wasmtypes.MsgExecuteContract{}): undermGasFunc(),
	}

	return gr
}

func (gr DeterministicGasRequirements) TxBaseGas(params authtypes.Params) uint64 {
	return gr.FreeBytes*params.TxSizeCostPerByte + gr.FreeSignatures*params.SigVerifyCostSecp256k1
}

// GasRequiredByMessage returns gas required by message and true if message is deterministic.
// Function returns 0 and false if message is undeterministic or unknown.
func (gr DeterministicGasRequirements) GasRequiredByMessage(msg sdk.Msg) (uint64, bool) {
	gasFunc, ok := gr.gasByMsg[msgName(msg)]
	if ok {
		return gasFunc(msg)
	}
	// Unknown message.
	return 0, false
}

func constGasFunc(constGasVal uint64) gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		return constGasVal, true
	}
}

func undermGasFunc() gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		return 0, true
	}
}

func (gr DeterministicGasRequirements) autzMsgExecGasFunc(authzMsgExecOverhead uint64) gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		m, ok := msg.(*authz.MsgExec)
		if !ok {
			return 0, false // FIX here and in other places.
		}

		totalGas := uint64(0)
		childMsgs, err := m.GetMessages()
		if err != nil {
			return 0, false
		}
		for _, childMsg := range childMsgs {
			gas, isDeterministic := gr.GasRequiredByMessage(childMsg)
			if !isDeterministic {
				return 0, false
			}
			totalGas += gas
		}
		return authzMsgExecOverhead + totalGas, true
	}
}

func bankSendMsgGasFunc(bankSendPerEntryGas uint64) gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		m := msg.(*banktypes.MsgSend)
		entriesNum := len(m.Amount)

		return uint64(lo.Max([]int{entriesNum, 1})) * bankSendPerEntryGas, true
	}
}

func bankMultiSendMsgGasFunc(bankMultiSendPerEntryGas uint64) gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		m := msg.(*banktypes.MsgMultiSend)
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

// Samples of values returned by function:
// "/cosmos.distribution.v1beta1.MsgFundCommunityPool"
// "/coreum.asset.ft.v1.MsgMint"
func msgName(msg sdk.Msg) string {
	return sdk.MsgTypeURL(msg)
}
