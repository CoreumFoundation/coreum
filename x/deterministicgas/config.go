package deterministicgas

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/samber/lo"

	assetfttypes "github.com/CoreumFoundation/coreum/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/x/asset/nft/types"
	nfttypes "github.com/CoreumFoundation/coreum/x/nft"
)

type gasByMsgFunc = func(msg sdk.Msg) (uint64, bool)

// Config specifies gas required by all transaction types
// Crisis module is intentionally skipped here because it is already deterministic by design and fee is specified
// using `consume_fee` param in genesis.
type Config struct {
	FixedGas uint64

	freeBytes      uint64
	freeSignatures uint64

	gasByMsg map[string]gasByMsgFunc
}

// DefaultConfig returns default config for deterministic gas.
func DefaultConfig() Config {
	cfg := Config{
		FixedGas:       50000,
		freeBytes:      2048,
		freeSignatures: 1,
	}

	cfg.gasByMsg = map[string]gasByMsgFunc{
		// asset/ft
		MsgType(&assetfttypes.MsgIssue{}):               constantGasFunc(70000),
		MsgType(&assetfttypes.MsgMint{}):                constantGasFunc(11000),
		MsgType(&assetfttypes.MsgBurn{}):                constantGasFunc(23000),
		MsgType(&assetfttypes.MsgFreeze{}):              constantGasFunc(5000),
		MsgType(&assetfttypes.MsgUnfreeze{}):            constantGasFunc(2500),
		MsgType(&assetfttypes.MsgGloballyFreeze{}):      constantGasFunc(5000),
		MsgType(&assetfttypes.MsgGloballyUnfreeze{}):    constantGasFunc(2500),
		MsgType(&assetfttypes.MsgSetWhitelistedLimit{}): constantGasFunc(5000),

		// asset/nft
		MsgType(&assetnfttypes.MsgBurn{}):       constantGasFunc(16000),
		MsgType(&assetnfttypes.MsgIssueClass{}): constantGasFunc(16000),
		MsgType(&assetnfttypes.MsgMint{}):       constantGasFunc(39000),
		MsgType(&assetnfttypes.MsgFreeze{}):     constantGasFunc(7000),
		MsgType(&assetnfttypes.MsgUnfreeze{}):   constantGasFunc(5000),

		// authz
		MsgType(&authz.MsgExec{}):   cfg.authzMsgExecGasFunc(2000),
		MsgType(&authz.MsgGrant{}):  constantGasFunc(7000),
		MsgType(&authz.MsgRevoke{}): constantGasFunc(2500),

		// bank
		MsgType(&banktypes.MsgSend{}):      bankSendMsgGasFunc(24000),
		MsgType(&banktypes.MsgMultiSend{}): bankMultiSendMsgGasFunc(11000),

		// distribution
		MsgType(&distributiontypes.MsgFundCommunityPool{}):           constantGasFunc(15000),
		MsgType(&distributiontypes.MsgSetWithdrawAddress{}):          constantGasFunc(5000),
		MsgType(&distributiontypes.MsgWithdrawDelegatorReward{}):     constantGasFunc(65000),
		MsgType(&distributiontypes.MsgWithdrawValidatorCommission{}): constantGasFunc(22000),

		// feegrant
		MsgType(&feegranttypes.MsgGrantAllowance{}):  constantGasFunc(10000),
		MsgType(&feegranttypes.MsgRevokeAllowance{}): constantGasFunc(2500),

		// gov
		MsgType(&govtypes.MsgSubmitProposal{}): constantGasFunc(65000),
		MsgType(&govtypes.MsgVote{}):           constantGasFunc(7000),
		MsgType(&govtypes.MsgVoteWeighted{}):   constantGasFunc(9000),
		MsgType(&govtypes.MsgDeposit{}):        constantGasFunc(52000),

		// nft
		MsgType(&nfttypes.MsgSend{}): constantGasFunc(16000),

		// slashing
		MsgType(&slashingtypes.MsgUnjail{}): constantGasFunc(25000),

		// staking
		MsgType(&stakingtypes.MsgDelegate{}):        constantGasFunc(69000),
		MsgType(&stakingtypes.MsgUndelegate{}):      constantGasFunc(112000),
		MsgType(&stakingtypes.MsgBeginRedelegate{}): constantGasFunc(142000),
		MsgType(&stakingtypes.MsgCreateValidator{}): constantGasFunc(76000),
		MsgType(&stakingtypes.MsgEditValidator{}):   constantGasFunc(13000),

		// wasm
		MsgType(&wasmtypes.MsgUpdateAdmin{}): constantGasFunc(8000),
		MsgType(&wasmtypes.MsgClearAdmin{}):  constantGasFunc(6500),
	}

	registerUndeterministicGasFuncs(
		&cfg,
		[]sdk.Msg{
			// crisis
			// MsgVerifyInvariant is defined as undeterministic since fee
			// charged by this tx type is defined as param inside module.
			&crisistypes.MsgVerifyInvariant{},

			// evidence
			// MsgSubmitEvidence is defined as undeterministic since we do not
			// have any custom evidence type implemented, so it should fail on
			// ValidateBasic step.
			&evidencetypes.MsgSubmitEvidence{},

			// wasm
			&wasmtypes.MsgStoreCode{},
			&wasmtypes.MsgInstantiateContract{},
			&wasmtypes.MsgInstantiateContract2{},
			&wasmtypes.MsgExecuteContract{},
			&wasmtypes.MsgMigrateContract{},
			&wasmtypes.MsgIBCSend{},
			&wasmtypes.MsgIBCCloseChannel{},
		},
	)

	return cfg
}

// TxBaseGas is the free gas we give to every transaction to cover costs of
// tx size and signature verification. TxBaseGas is covered by FixedGas.
func (cfg Config) TxBaseGas(params authtypes.Params) uint64 {
	return cfg.freeBytes*params.TxSizeCostPerByte + cfg.freeSignatures*params.SigVerifyCostSecp256k1
}

// GasRequiredByMessage returns gas required by message and true if message is deterministic.
// Function returns 0 and false if message is undeterministic or unknown.
func (cfg Config) GasRequiredByMessage(msg sdk.Msg) (uint64, bool) {
	gasFunc, ok := cfg.gasByMsg[MsgType(msg)]
	if ok {
		return gasFunc(msg)
	}

	// Currently we treat unknown message types as undeterministic.
	// In the future other approach could be to return third boolean parameter
	// identifying if message is known and report unknown messages to monitoring.
	return 0, false
}

// MsgType returns TypeURL of a msg in cosmos SDK style.
// Samples of values returned by the function:
// "/cosmos.distribution.v1beta1.MsgFundCommunityPool"
// "/coreum.asset.ft.v1.MsgMint".
func MsgType(msg sdk.Msg) string {
	return sdk.MsgTypeURL(msg)
}

// NOTE: we need to pass Config by pointer here because
// it needs to be initialized later map with all msg types inside to estimate gas recursively.
func (cfg *Config) authzMsgExecGasFunc(authzMsgExecOverhead uint64) gasByMsgFunc {
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
			gas, isDeterministic := cfg.GasRequiredByMessage(childMsg)
			if !isDeterministic {
				return 0, false
			}
			totalGas += gas
		}
		return totalGas, true
	}
}

func registerUndeterministicGasFuncs(cfg *Config, msgs []sdk.Msg) {
	for _, msg := range msgs {
		cfg.gasByMsg[MsgType(msg)] = underministicGasFunc()
	}
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

func bankMultiSendMsgGasFunc(bankMultiSendPerOperationGas uint64) gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		m, ok := msg.(*banktypes.MsgMultiSend)
		if !ok {
			return 0, false
		}
		totalOperationsNum := 0
		for _, inp := range m.Inputs {
			totalOperationsNum += len(inp.Coins)
		}

		for _, outp := range m.Outputs {
			totalOperationsNum += len(outp.Coins)
		}

		// Minimum 2 operations (1 input & 1 output) should be present inside any multi-send.
		return uint64(lo.Max([]int{totalOperationsNum, 2})) * bankMultiSendPerOperationGas, true
	}
}
