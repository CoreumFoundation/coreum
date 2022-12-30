package config

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/x/asset/nft/types"
	nfttypes "github.com/CoreumFoundation/coreum/x/nft"
)

type gasByMsgFunc = func(msg sdk.Msg) (uint64, bool)

type GasConfig struct {
	FixedGas       uint64
	FreeBytes      uint64
	FreeSignatures uint64

	gasByMsg map[string]gasByMsgFunc
}

func DefaultGasRequirementsV2() GasConfig {
	gr := GasConfig{
		FixedGas:       50000,
		FreeBytes:      2048,
		FreeSignatures: 1,
	}

	gr.gasByMsg = map[string]gasByMsgFunc{
		// authz
		msgName(&authz.MsgExec{}):   gr.autzMsgExecFunc(2000),
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
		msgName(&banktypes.MsgSend{}):      bankSenMsgFunc(22000),
		msgName(&banktypes.MsgMultiSend{}): bankMultiSendMsgFunc(27000),

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

func (gr GasConfig) GasRequiredByMessageV2(msg sdk.Msg) (uint64, bool) {
	gas, ok := gr.gasByMsg[msgName(msg)]
	if ok {
		return gas(msg)
	}
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

func (gr GasConfig) autzMsgExecFunc(authzMsgExecOverhead uint64) gasByMsgFunc {
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
			gas, isDeterministic := gr.GasRequiredByMessageV2(childMsg)
			if !isDeterministic {
				return 0, false
			}
			totalGas += gas
		}
		return authzMsgExecOverhead + totalGas, true

	}
}

func bankSenMsgFunc(bankSendPerEntryGas uint64) gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		m := msg.(*banktypes.MsgSend)
		entriesNum := len(m.Amount)
		if len(m.Amount) == 0 {
			entriesNum = 1
		}
		return uint64(entriesNum) * bankSendPerEntryGas, true
	}
}

func bankMultiSendMsgFunc(bankMultiSendPerEntryGas uint64) gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		m := msg.(*banktypes.MsgMultiSend)
		entriesNum := 0
		for _, inp := range m.Inputs {
			entriesNum += len(inp.Coins)
		}

		outputEntriesNum := 0
		for _, outp := range m.Outputs {
			outputEntriesNum += len(outp.Coins)
		}
		if outputEntriesNum > entriesNum {
			entriesNum = outputEntriesNum
		}

		if entriesNum == 0 {
			entriesNum = 1
		}
		return uint64(entriesNum) * bankMultiSendPerEntryGas, true
	}
}

// Samples of values returned by function:
// "/cosmos.distribution.v1beta1.MsgFundCommunityPool"
// "/coreum.asset.ft.v1.MsgMint"
func msgName(msg sdk.Msg) string {
	return sdk.MsgTypeURL(msg)
}
