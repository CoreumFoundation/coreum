package wasm

import (
	"context"
	"encoding/json"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/testutil/event"
)

var gasMultiplier = 1.5

// InstantiateConfig contains params specific to contract instantiation.
type InstantiateConfig struct {
	accessType wasmtypes.AccessType
	payload    json.RawMessage
	amount     sdk.Coin
	label      string
	CodeID     uint64
}

// DeployAndInstantiate deploys, instantiate the wasm contract and returns its address.
func DeployAndInstantiate(ctx context.Context, clientCtx tx.ClientContext, txf tx.Factory, wasmData []byte, initConfig InstantiateConfig) (string, uint64, error) {
	codeID, err := deploy(ctx, clientCtx, txf, wasmData)
	if err != nil {
		return "", 0, err
	}

	initConfig.CodeID = codeID
	contractAddr, err := instantiate(ctx, clientCtx, txf, initConfig)
	if err != nil {
		return "", 0, err
	}

	return contractAddr, codeID, nil
}

// Execute executes the wasm contract with the payload and optionally funding amount.
func Execute(ctx context.Context, clientCtx tx.ClientContext, txf tx.Factory, contractAddr string, payload json.RawMessage, fundAmt sdk.Coin) (int64, error) {
	funds := sdk.NewCoins()
	if !fundAmt.Amount.IsNil() {
		funds = funds.Add(fundAmt)
	}

	msg := &wasmtypes.MsgExecuteContract{
		Sender:   clientCtx.FromAddress().String(),
		Contract: contractAddr,
		Msg:      wasmtypes.RawContractMessage(payload),
		Funds:    funds,
	}

	txf = txf.
		WithGasAdjustment(gasMultiplier)

	res, err := tx.BroadcastTx(ctx, clientCtx, txf, msg)
	if err != nil {
		return 0, err
	}
	return res.GasUsed, nil
}

// Query queries the contract with the requested payload.
func Query(ctx context.Context, clientCtx tx.ClientContext, contractAddr string, payload json.RawMessage) (json.RawMessage, error) {
	query := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddr,
		QueryData: wasmtypes.RawContractMessage(payload),
	}

	wasmClient := wasmtypes.NewQueryClient(clientCtx)
	resp, err := wasmClient.SmartContractState(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "WASMQueryClient returned an error after smart contract state Query")
	}

	return json.RawMessage(resp.Data), nil
}

// IsPinned returns true if smart contract is pinned
func IsPinned(ctx context.Context, clientCtx tx.ClientContext, codeID uint64) (bool, error) {
	wasmClient := wasmtypes.NewQueryClient(clientCtx)
	resp, err := wasmClient.PinnedCodes(ctx, &wasmtypes.QueryPinnedCodesRequest{})
	if err != nil {
		return false, errors.Wrap(err, "WASMQueryClient returned an error after querying pinned contracts")
	}
	for _, c := range resp.CodeIDs {
		if c == codeID {
			return true, nil
		}
	}
	return false, nil
}

// deploys the wasm contract and returns its codeID.
func deploy(ctx context.Context, clientCtx tx.ClientContext, txf tx.Factory, wasmData []byte) (uint64, error) {
	msgStoreCode := &wasmtypes.MsgStoreCode{
		Sender:       clientCtx.FromAddress().String(),
		WASMByteCode: wasmData,
	}

	txf = txf.
		WithGasAdjustment(gasMultiplier)

	res, err := tx.BroadcastTx(ctx, clientCtx, txf, msgStoreCode)
	if err != nil {
		return 0, err
	}

	codeID, err := event.FindUint64EventAttribute(res.Events, wasmtypes.EventTypeStoreCode, wasmtypes.AttributeKeyCodeID)
	if err != nil {
		return 0, err
	}

	return codeID, nil
}

// instantiates the contract and returns the contract address.
func instantiate(ctx context.Context, clientCtx tx.ClientContext, txf tx.Factory, req InstantiateConfig) (string, error) {
	funds := sdk.NewCoins()
	if amount := req.amount; !amount.Amount.IsNil() {
		funds = funds.Add(amount)
	}
	msg := &wasmtypes.MsgInstantiateContract{
		Sender: clientCtx.FromAddress().String(),
		CodeID: req.CodeID,
		Label:  req.label,
		Msg:    wasmtypes.RawContractMessage(req.payload),
		Funds:  funds,
	}

	txf = txf.
		WithGasAdjustment(gasMultiplier)

	res, err := tx.BroadcastTx(ctx, clientCtx, txf, msg)
	if err != nil {
		return "", err
	}

	contractAddr, err := event.FindStringEventAttribute(res.Events, wasmtypes.EventTypeInstantiate, wasmtypes.AttributeKeyContractAddr)
	if err != nil {
		return "", err
	}

	return contractAddr, nil
}
