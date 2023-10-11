package deterministicgas

import (
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/armon/go-metrics"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	nfttypes "github.com/cosmos/cosmos-sdk/x/nft"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcconnectiontypes "github.com/cosmos/ibc-go/v7/modules/core/03-connection/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/samber/lo"

	assetfttypes "github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/v3/x/asset/nft/types"
	cnfttypes "github.com/CoreumFoundation/coreum/v3/x/nft"
)

// These constants define gas for messages which have custom calculation logic.
const (
	BankSendPerCoinGas            = 50000
	BankMultiSendPerOperationsGas = 35000
	AuthzExecOverhead             = 2000
)

type (
	// MsgURL is a type used to uniquely identify msg in URL-like format. E.g "/coreum.asset.ft.v1.MsgMint".
	MsgURL string

	gasByMsgFunc = func(msg sdk.Msg) (uint64, bool)
)

// Config specifies gas required by all transaction types
// Crisis module is intentionally skipped here because it is already deterministic by design and fee is specified
// using `consume_fee` param in genesis.
type Config struct {
	FixedGas uint64

	FreeBytes      uint64
	FreeSignatures uint64

	gasByMsg map[MsgURL]gasByMsgFunc
}

// DefaultConfig returns default config for deterministic gas.
//
//nolint:funlen
func DefaultConfig() Config {
	cfg := Config{
		FixedGas:       65000,
		FreeBytes:      2048,
		FreeSignatures: 1,
	}

	cfg.gasByMsg = map[MsgURL]gasByMsgFunc{
		// asset/ft
		MsgToMsgURL(&assetfttypes.MsgIssue{}):               constantGasFunc(70000),
		MsgToMsgURL(&assetfttypes.MsgMint{}):                constantGasFunc(31000),
		MsgToMsgURL(&assetfttypes.MsgBurn{}):                constantGasFunc(35000),
		MsgToMsgURL(&assetfttypes.MsgFreeze{}):              constantGasFunc(8500),
		MsgToMsgURL(&assetfttypes.MsgUnfreeze{}):            constantGasFunc(4000),
		MsgToMsgURL(&assetfttypes.MsgGloballyFreeze{}):      constantGasFunc(5000),
		MsgToMsgURL(&assetfttypes.MsgGloballyUnfreeze{}):    constantGasFunc(2500),
		MsgToMsgURL(&assetfttypes.MsgSetWhitelistedLimit{}): constantGasFunc(9000),
		// TODO: Reestimate when next token upgrade is prepared
		MsgToMsgURL(&assetfttypes.MsgUpgradeTokenV1{}): constantGasFunc(25000),

		// asset/nft
		MsgToMsgURL(&assetnfttypes.MsgBurn{}):                constantGasFunc(26000),
		MsgToMsgURL(&assetnfttypes.MsgIssueClass{}):          constantGasFunc(16000),
		MsgToMsgURL(&assetnfttypes.MsgMint{}):                constantGasFunc(39000),
		MsgToMsgURL(&assetnfttypes.MsgFreeze{}):              constantGasFunc(8000),
		MsgToMsgURL(&assetnfttypes.MsgUnfreeze{}):            constantGasFunc(5000),
		MsgToMsgURL(&assetnfttypes.MsgAddToWhitelist{}):      constantGasFunc(7000),
		MsgToMsgURL(&assetnfttypes.MsgRemoveFromWhitelist{}): constantGasFunc(3500),

		// authz
		// FIXME (v47-deterministic): We need a procedure to estimate the overhead of the authz. Proposal:
		// 1. Estimate normal message
		// 2. Estimate the same message executed using authz
		// 3. Subtract one from the other
		// We should have an integration test doing this.
		MsgToMsgURL(&authz.MsgExec{}):   cfg.authzMsgExecGasFunc(AuthzExecOverhead),
		MsgToMsgURL(&authz.MsgGrant{}):  constantGasFunc(28000),
		MsgToMsgURL(&authz.MsgRevoke{}): constantGasFunc(8000),

		// bank
		MsgToMsgURL(&banktypes.MsgSend{}):      bankSendMsgGasFunc(BankSendPerCoinGas),
		MsgToMsgURL(&banktypes.MsgMultiSend{}): bankMultiSendMsgGasFunc(BankMultiSendPerOperationsGas),

		// distribution
		MsgToMsgURL(&distributiontypes.MsgFundCommunityPool{}):           constantGasFunc(17000),
		MsgToMsgURL(&distributiontypes.MsgSetWithdrawAddress{}):          constantGasFunc(5000),
		MsgToMsgURL(&distributiontypes.MsgWithdrawDelegatorReward{}):     constantGasFunc(79000),
		MsgToMsgURL(&distributiontypes.MsgWithdrawValidatorCommission{}): constantGasFunc(22000),

		// feegrant
		MsgToMsgURL(&feegranttypes.MsgGrantAllowance{}):  constantGasFunc(11000),
		MsgToMsgURL(&feegranttypes.MsgRevokeAllowance{}): constantGasFunc(2500),

		// gov
		// FIXME(v47-deterministic): check that if we want to support both go types
		MsgToMsgURL(&govtypesv1beta1.MsgVote{}):         constantGasFunc(6000),
		MsgToMsgURL(&govtypesv1beta1.MsgVoteWeighted{}): constantGasFunc(9000),
		MsgToMsgURL(&govtypesv1beta1.MsgDeposit{}):      constantGasFunc(85000),

		MsgToMsgURL(&govtypesv1.MsgVote{}):         constantGasFunc(6000),
		MsgToMsgURL(&govtypesv1.MsgVoteWeighted{}): constantGasFunc(6500),
		// FIXME (v47-deterministic): We must add integration test executing this message to have data to analyze
		MsgToMsgURL(&govtypesv1.MsgDeposit{}): constantGasFunc(52000),

		// nft
		MsgToMsgURL(&nfttypes.MsgSend{}): constantGasFunc(25000),

		// cnft
		// Deprecated: this will be removed in the next release alongside the cnft types.
		MsgToMsgURL(&cnfttypes.MsgSend{}): constantGasFunc(25000),

		// slashing
		// FIXME (v47-deterministic): We must add integration test executing this message to have data to analyze
		MsgToMsgURL(&slashingtypes.MsgUnjail{}): constantGasFunc(25000),

		// staking
		MsgToMsgURL(&stakingtypes.MsgDelegate{}):        constantGasFunc(83000),
		MsgToMsgURL(&stakingtypes.MsgUndelegate{}):      constantGasFunc(112000),
		MsgToMsgURL(&stakingtypes.MsgBeginRedelegate{}): constantGasFunc(157000),
		MsgToMsgURL(&stakingtypes.MsgCreateValidator{}): constantGasFunc(117000),
		MsgToMsgURL(&stakingtypes.MsgEditValidator{}):   constantGasFunc(13000),

		// vesting
		MsgToMsgURL(&vestingtypes.MsgCreateVestingAccount{}): constantGasFunc(30000),

		// wasm
		MsgToMsgURL(&wasmtypes.MsgUpdateAdmin{}): constantGasFunc(8000),
		MsgToMsgURL(&wasmtypes.MsgClearAdmin{}):  constantGasFunc(6500),

		// ibc transfer
		MsgToMsgURL(&ibctransfertypes.MsgTransfer{}): constantGasFunc(54000),
	}

	registerNondeterministicGasFuncs(
		&cfg,
		[]sdk.Msg{
			// auth
			&authtypes.MsgUpdateParams{}, // This is non-deterministic because all the gov proposals are non-deterministic anyway

			// bank
			&banktypes.MsgSetSendEnabled{}, // This is non-deterministic because all the gov proposals are non-deterministic anyway
			&banktypes.MsgUpdateParams{},   // This is non-deterministic because all the gov proposals are non-deterministic anyway

			// consensus
			&consensustypes.MsgUpdateParams{}, // This is non-deterministic because all the gov proposals are non-deterministic anyway

			// crisis
			&crisistypes.MsgUpdateParams{}, // This is non-deterministic because all the gov proposals are non-deterministic anyway

			// distribution
			&distributiontypes.MsgUpdateParams{},       // This is non-deterministic because all the gov proposals are non-deterministic anyway
			&distributiontypes.MsgCommunityPoolSpend{}, // This is non-deterministic because all the gov proposals are non-deterministic anyway

			// gov
			// MsgSubmitProposal is defined as nondeterministic because it runs a proposal handler function
			// specific for each proposal and those functions consume unknown amount of gas.
			&govtypesv1beta1.MsgSubmitProposal{},

			// FIXME(v47-deterministic): check that if we want to support both go types
			&govtypesv1.MsgSubmitProposal{},
			&govtypesv1.MsgExecLegacyContent{},
			&govtypesv1.MsgUpdateParams{}, // This is non-deterministic because all the gov proposals are non-deterministic anyway

			// crisis
			// MsgVerifyInvariant is defined as nondeterministic since fee
			// charged by this tx type is defined as param inside module.
			&crisistypes.MsgVerifyInvariant{},

			// evidence
			// MsgSubmitEvidence is defined as nondeterministic since we do not
			// have any custom evidence type implemented, so it should fail on
			// ValidateBasic step.
			&evidencetypes.MsgSubmitEvidence{},

			// mint
			&minttypes.MsgUpdateParams{}, // This is non-deterministic because all the gov proposals are non-deterministic anyway

			// staking
			&stakingtypes.MsgUpdateParams{}, // This is non-deterministic because all the gov proposals are non-deterministic anyway
			// FIXME(v47-deterministic): add message to deterministic (we have separate task for it)
			&stakingtypes.MsgCancelUnbondingDelegation{},

			// slashing
			&slashingtypes.MsgUpdateParams{}, // This is non-deterministic because all the gov proposals are non-deterministic anyway

			// slashing
			&slashingtypes.MsgUpdateParams{}, // This is non-deterministic because all the gov proposals are non-deterministic anyway

			// upgrade
			&upgradetypes.MsgCancelUpgrade{},   // This is non-deterministic because all the gov proposals are non-deterministic anyway
			&upgradetypes.MsgSoftwareUpgrade{}, // This is non-deterministic because all the gov proposals are non-deterministic anyway

			// vesting
			// FIXME(v47-deterministic): add message to deterministic (we have separate task for it)
			&vestingtypes.MsgCreatePeriodicVestingAccount{},
			// FIXME(v47-deterministic): add message to deterministic (we have separate task for it)
			&vestingtypes.MsgCreatePermanentLockedAccount{},

			// wasm
			&wasmtypes.MsgStoreCode{},
			&wasmtypes.MsgInstantiateContract{},
			&wasmtypes.MsgInstantiateContract2{},
			&wasmtypes.MsgExecuteContract{},
			&wasmtypes.MsgMigrateContract{},
			&wasmtypes.MsgIBCSend{},
			&wasmtypes.MsgIBCCloseChannel{},
			&wasmtypes.MsgUpdateInstantiateConfig{},
			&wasmtypes.MsgUpdateParams{}, // This is non-deterministic because all the gov proposals are non-deterministic anyway
			&wasmtypes.MsgUnpinCodes{},
			&wasmtypes.MsgPinCodes{},
			&wasmtypes.MsgSudoContract{},
			&wasmtypes.MsgStoreAndInstantiateContract{},

			// ibc/core/client
			&ibcclienttypes.MsgCreateClient{},
			&ibcclienttypes.MsgCreateClient{},
			&ibcclienttypes.MsgUpdateClient{},
			&ibcclienttypes.MsgUpgradeClient{},
			&ibcclienttypes.MsgSubmitMisbehaviour{},

			// ibc/core/connection
			&ibcconnectiontypes.MsgConnectionOpenInit{},
			&ibcconnectiontypes.MsgConnectionOpenTry{},
			&ibcconnectiontypes.MsgConnectionOpenAck{},
			&ibcconnectiontypes.MsgConnectionOpenConfirm{},

			// ibc/core/channel
			&ibcchanneltypes.MsgChannelOpenInit{},
			&ibcchanneltypes.MsgChannelOpenTry{},
			&ibcchanneltypes.MsgChannelOpenAck{},
			&ibcchanneltypes.MsgChannelOpenConfirm{},
			&ibcchanneltypes.MsgChannelCloseInit{},
			&ibcchanneltypes.MsgChannelCloseConfirm{},
			&ibcchanneltypes.MsgRecvPacket{},
			&ibcchanneltypes.MsgTimeout{},
			&ibcchanneltypes.MsgTimeoutOnClose{},
			&ibcchanneltypes.MsgAcknowledgement{},
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
	gasFunc, ok := cfg.gasByMsg[MsgToMsgURL(msg)]
	if ok {
		return gasFunc(msg)
	}

	// Currently we treat unknown message types as nondeterministic.
	// In the future other approach could be to return third boolean parameter
	// identifying if message is known and report unknown messages to monitoring.
	reportUnknownMessageMetric(MsgToMsgURL(msg))
	return 0, false
}

// GasByMessageMap returns copy mapping of message types and functions to calculate gas for specific type.
func (cfg Config) GasByMessageMap() map[MsgURL]gasByMsgFunc {
	newGasByMsg := make(map[MsgURL]gasByMsgFunc, len(cfg.gasByMsg))
	for k, v := range cfg.gasByMsg {
		newGasByMsg[k] = v
	}
	return newGasByMsg
}

// MsgToMsgURL returns TypeURL of a msg in cosmos SDK style.
// Samples of values returned by the function:
// "/cosmos.distribution.v1beta1.MsgFundCommunityPool"
// "/coreum.asset.ft.v1.MsgMint".
func MsgToMsgURL(msg sdk.Msg) MsgURL {
	return MsgURL(sdk.MsgTypeURL(msg))
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
		cfg.gasByMsg[MsgToMsgURL(msg)] = nondeterministicGasFunc
	}
}

func constantGasFunc(constGasVal uint64) gasByMsgFunc {
	return func(msg sdk.Msg) (uint64, bool) {
		return constGasVal, true
	}
}

func nondeterministicGasFunc(_ sdk.Msg) (uint64, bool) {
	return 0, false
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

func reportUnknownMessageMetric(msgURL MsgURL) {
	metrics.IncrCounterWithLabels([]string{"deterministic_gas_unknown_message"}, 1, []metrics.Label{
		{Name: "msg_name", Value: string(msgURL)},
	})
}
