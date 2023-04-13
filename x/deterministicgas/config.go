package deterministicgas

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/armon/go-metrics"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
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

//go:generate go run spec/generate.go spec/README.md

// These constants define gas for messages which have custom calculation logic.
const (
	BankSendPerCoinGas            = 24000
	BankMultiSendPerOperationsGas = 11000
	AuthzExecOverhead             = 2000
)

type (
	gasByMsgFunc = func(msg sdk.Msg) (uint64, bool)

	MsgType = string
)

// Config specifies gas required by all transaction types
// Crisis module is intentionally skipped here because it is already deterministic by design and fee is specified
// using `consume_fee` param in genesis.
type Config struct {
	FixedGas uint64

	FreeBytes      uint64
	FreeSignatures uint64

	gasByMsg map[MsgType]gasByMsgFunc
}

// DefaultConfig returns default config for deterministic gas.
//
//nolint:funlen
func DefaultConfig() Config {
	cfg := Config{
		FixedGas:       50000,
		FreeBytes:      2048,
		FreeSignatures: 1,
	}

	cfg.gasByMsg = map[MsgType]gasByMsgFunc{
		// asset/ft
		MsgTypeURL(&assetfttypes.MsgIssue{}):               constantGasFunc(70000),
		MsgTypeURL(&assetfttypes.MsgMint{}):                constantGasFunc(11000),
		MsgTypeURL(&assetfttypes.MsgBurn{}):                constantGasFunc(23000),
		MsgTypeURL(&assetfttypes.MsgFreeze{}):              constantGasFunc(5000),
		MsgTypeURL(&assetfttypes.MsgUnfreeze{}):            constantGasFunc(2500),
		MsgTypeURL(&assetfttypes.MsgGloballyFreeze{}):      constantGasFunc(5000),
		MsgTypeURL(&assetfttypes.MsgGloballyUnfreeze{}):    constantGasFunc(2500),
		MsgTypeURL(&assetfttypes.MsgSetWhitelistedLimit{}): constantGasFunc(5000),

		// asset/nft
		MsgTypeURL(&assetnfttypes.MsgBurn{}):                constantGasFunc(16000),
		MsgTypeURL(&assetnfttypes.MsgIssueClass{}):          constantGasFunc(16000),
		MsgTypeURL(&assetnfttypes.MsgMint{}):                constantGasFunc(39000),
		MsgTypeURL(&assetnfttypes.MsgFreeze{}):              constantGasFunc(7000),
		MsgTypeURL(&assetnfttypes.MsgUnfreeze{}):            constantGasFunc(5000),
		MsgTypeURL(&assetnfttypes.MsgAddToWhitelist{}):      constantGasFunc(7000),
		MsgTypeURL(&assetnfttypes.MsgRemoveFromWhitelist{}): constantGasFunc(3500),

		// authz
		MsgTypeURL(&authz.MsgExec{}):   cfg.authzMsgExecGasFunc(AuthzExecOverhead),
		MsgTypeURL(&authz.MsgGrant{}):  constantGasFunc(7000),
		MsgTypeURL(&authz.MsgRevoke{}): constantGasFunc(2500),

		// bank
		MsgTypeURL(&banktypes.MsgSend{}):      bankSendMsgGasFunc(BankSendPerCoinGas),
		MsgTypeURL(&banktypes.MsgMultiSend{}): bankMultiSendMsgGasFunc(BankMultiSendPerOperationsGas),

		// distribution
		MsgTypeURL(&distributiontypes.MsgFundCommunityPool{}):           constantGasFunc(15000),
		MsgTypeURL(&distributiontypes.MsgSetWithdrawAddress{}):          constantGasFunc(5000),
		MsgTypeURL(&distributiontypes.MsgWithdrawDelegatorReward{}):     constantGasFunc(65000),
		MsgTypeURL(&distributiontypes.MsgWithdrawValidatorCommission{}): constantGasFunc(22000),

		// feegrant
		MsgTypeURL(&feegranttypes.MsgGrantAllowance{}):  constantGasFunc(10000),
		MsgTypeURL(&feegranttypes.MsgRevokeAllowance{}): constantGasFunc(2500),

		// gov
		MsgTypeURL(&govtypes.MsgVote{}):         constantGasFunc(7000),
		MsgTypeURL(&govtypes.MsgVoteWeighted{}): constantGasFunc(9000),
		MsgTypeURL(&govtypes.MsgDeposit{}):      constantGasFunc(52000),

		// nft
		MsgTypeURL(&nfttypes.MsgSend{}): constantGasFunc(16000),

		// slashing
		MsgTypeURL(&slashingtypes.MsgUnjail{}): constantGasFunc(25000),

		// staking
		MsgTypeURL(&stakingtypes.MsgDelegate{}):        constantGasFunc(69000),
		MsgTypeURL(&stakingtypes.MsgUndelegate{}):      constantGasFunc(112000),
		MsgTypeURL(&stakingtypes.MsgBeginRedelegate{}): constantGasFunc(142000),
		MsgTypeURL(&stakingtypes.MsgCreateValidator{}): constantGasFunc(76000),
		MsgTypeURL(&stakingtypes.MsgEditValidator{}):   constantGasFunc(13000),

		// vesting
		MsgTypeURL(&vestingtypes.MsgCreateVestingAccount{}): constantGasFunc(25000),

		// wasm
		MsgTypeURL(&wasmtypes.MsgUpdateAdmin{}): constantGasFunc(8000),
		MsgTypeURL(&wasmtypes.MsgClearAdmin{}):  constantGasFunc(6500),
	}

	registerNondeterministicGasFuncs(
		&cfg,
		[]sdk.Msg{
			// gov
			// MsgSubmitProposal is defined as nondeterministic because it runs a proposal handler function
			// specific for each proposal and those functions consume unknown amount of gas.
			&govtypes.MsgSubmitProposal{},

			// crisis
			// MsgVerifyInvariant is defined as nondeterministic since fee
			// charged by this tx type is defined as param inside module.
			&crisistypes.MsgVerifyInvariant{},

			// evidence
			// MsgSubmitEvidence is defined as nondeterministic since we do not
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
	return cfg.FreeBytes*params.TxSizeCostPerByte + cfg.FreeSignatures*params.SigVerifyCostSecp256k1
}

// GasRequiredByMessage returns gas required by message and true if message is deterministic.
// Function returns 0 and false if message is nondeterministic or unknown.
func (cfg Config) GasRequiredByMessage(msg sdk.Msg) (uint64, bool) {
	gasFunc, ok := cfg.gasByMsg[MsgTypeURL(msg)]
	if ok {
		return gasFunc(msg)
	}

	// Currently we treat unknown message types as nondeterministic.
	// In the future other approach could be to return third boolean parameter
	// identifying if message is known and report unknown messages to monitoring.
	reportUnknownMessageMetric(MsgTypeURL(msg))
	return 0, false
}

// GasByMessageMap returns copy mapping of message types and functions to calculate gas for specific type.
func (cfg Config) GasByMessageMap() map[MsgType]gasByMsgFunc {
	newGasByMsg := make(map[MsgType]gasByMsgFunc, len(cfg.gasByMsg))
	for k, v := range cfg.gasByMsg {
		newGasByMsg[k] = v
	}
	return newGasByMsg
}

// MsgTypeURL returns TypeURL of a msg in cosmos SDK style.
// Samples of values returned by the function:
// "/cosmos.distribution.v1beta1.MsgFundCommunityPool"
// "/coreum.asset.ft.v1.MsgMint".
func MsgTypeURL(msg sdk.Msg) MsgType {
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

func registerNondeterministicGasFuncs(cfg *Config, msgs []sdk.Msg) {
	for _, msg := range msgs {
		cfg.gasByMsg[MsgTypeURL(msg)] = nondeterministicGasFunc()
	}
}

func constantGasFunc(constGasVal uint64) gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		return constGasVal, true
	}
}

func nondeterministicGasFunc() gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		return 0, false
	}
}

func bankSendMsgGasFunc(bankSendPerCoinGas uint64) gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		m, ok := msg.(*banktypes.MsgSend)
		if !ok {
			return 0, false
		}
		entriesNum := len(m.Amount)

		return uint64(lo.Max([]int{entriesNum, 1})) * bankSendPerCoinGas, true
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

func reportUnknownMessageMetric(msgName MsgType) {
	metrics.IncrCounterWithLabels([]string{"deterministic_gas_unknown_message"}, 1, []metrics.Label{
		{Name: "msg_name", Value: msgName},
	})
}
