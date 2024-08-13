package simulation

import (
	"math/rand"
	"strings"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/CoreumFoundation/coreum/v4/x/asset/ft/types"
)

// Message types.
var (
	TypeMsgIssue = sdk.MsgTypeURL(&types.MsgIssue{})
)

// Simulation operation weights constants.
const (
	OpWeightMsgIssue      = "op_weight_msg_issue"
	DefaultWeightMsgIssue = 100
)

// OperationFactory creates simulation messages.
type OperationFactory struct {
	appParams simtypes.AppParams
	cdc       codec.JSONCodec
	ak        types.AccountKeeper
	bk        types.BankKeeper
}

// NewOperationFactory returns new instance of the OperationFactory.
func NewOperationFactory(
	appParams simtypes.AppParams,
	cdc codec.JSONCodec,
	ak types.AccountKeeper,
	bk types.BankKeeper,
) *OperationFactory {
	return &OperationFactory{
		appParams: appParams,
		cdc:       cdc,
		ak:        ak,
		bk:        bk,
	}
}

// WeightedOperations returns all the operations from the module with their respective weights.
func (op *OperationFactory) WeightedOperations() simulation.WeightedOperations {
	// make the weights updatable by the simulation
	var weightMsgIssue int
	op.appParams.GetOrGenerate(OpWeightMsgIssue, &weightMsgIssue, nil,
		func(_ *rand.Rand) {
			weightMsgIssue = DefaultWeightMsgIssue
		},
	)
	return simulation.WeightedOperations{
		simulation.NewWeightedOperation(
			weightMsgIssue,
			op.simulateMsgIssue,
		),
	}
}

func (op *OperationFactory) simulateMsgIssue(
	r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context,
	accs []simtypes.Account, chainID string,
) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
	senderAcc, _ := simtypes.RandomAcc(r, accs)
	msg, skip := op.randomIssueMsg(ctx, r, senderAcc.Address)
	if skip {
		return simtypes.NoOpMsg(types.ModuleName, TypeMsgIssue, "skip issue"), nil, nil
	}

	err := op.sendMsg(ctx, r, chainID, []cryptotypes.PrivKey{senderAcc.PrivKey}, app, senderAcc.Address, msg)
	if err != nil {
		return simtypes.NoOpMsg(types.ModuleName, sdk.MsgTypeURL(msg), "invalid issuance"), nil, err
	}

	return simtypes.NewOperationMsg(msg, false, ""), nil, nil
}

func (op *OperationFactory) randomIssueMsg(
	ctx sdk.Context,
	r *rand.Rand,
	sender sdk.AccAddress,
	// msg, skip
) (*types.MsgIssue, bool) {
	// if the account is not stored skip
	acc := op.ak.GetAccount(ctx, sender)
	if acc == nil {
		return nil, true
	}
	msg := &types.MsgIssue{
		Issuer:        sender.String(),
		Symbol:        simtypes.RandStringOfLength(r, simtypes.RandIntBetween(r, 3, 127)),
		Subunit:       strings.ToLower(simtypes.RandStringOfLength(r, simtypes.RandIntBetween(r, 1, 50))),
		Precision:     uint32(simtypes.RandIntBetween(r, 1, 20)),
		InitialAmount: simtypes.RandomAmount(r, sdkmath.NewIntWithDecimal(1, 30)),
		Description:   simtypes.RandStringOfLength(r, simtypes.RandIntBetween(r, 1, types.MaxDescriptionLength)),
		Features:      nil,
		// TODO(dzmitryhil) fix the simulation to work with the commissions since now it is failed
		// in the distribution EndBlocker since tries to allocate all tokens for the fee_collector
		// and the fee_collector has the asset_ft_tokens with the SendCommissionRate and BurnRate
		// BurnRate: sdkmath.LegacyNewDec(int64(simtypes.RandIntBetween(r, 1, 1000))).QuoInt64(10000),
		// SendCommissionRate: sdkmath.LegacyNewDec(int64(simtypes.RandIntBetween(r, 1, 1000))).QuoInt64(10000),
		URI:     simtypes.RandStringOfLength(r, simtypes.RandIntBetween(r, 1, types.MaxURILength)),
		URIHash: simtypes.RandStringOfLength(r, simtypes.RandIntBetween(r, 1, types.MaxURIHashLength)),
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, true
	}

	return msg, false
}

func (op *OperationFactory) sendMsg(
	ctx sdk.Context,
	r *rand.Rand,
	chainID string,
	privKeys []cryptotypes.PrivKey,
	app *baseapp.BaseApp,
	sender sdk.AccAddress,
	msg *types.MsgIssue,
) error {
	account := op.ak.GetAccount(ctx, sender)
	txGen := moduletestutil.MakeTestEncodingConfig().TxConfig
	tx, err := simtestutil.GenSignedMockTx(
		r,
		txGen,
		[]sdk.Msg{msg},
		sdk.Coins{},
		simtestutil.DefaultGenTxGas,
		chainID,
		[]uint64{account.GetAccountNumber()},
		[]uint64{account.GetSequence()},
		privKeys...,
	)
	if err != nil {
		return err
	}
	_, _, err = app.SimDeliver(txGen.TxEncoder(), tx)
	if err != nil {
		return err
	}

	return nil
}
