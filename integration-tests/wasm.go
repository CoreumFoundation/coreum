package integrationtests

import (
	"context"
	"encoding/json"

	sdkerrors "cosmossdk.io/errors"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/v2/pkg/client"
	"github.com/CoreumFoundation/coreum/v2/testutil/event"
)

var gasMultiplier = 1.5

// InstantiateConfig contains params specific to contract instantiation.
type InstantiateConfig struct {
	Admin      sdk.AccAddress
	AccessType wasmtypes.AccessType
	Payload    json.RawMessage
	Amount     sdk.Coin
	Label      string
	CodeID     uint64
}

// Wasm provides wasm client for the testing.
type Wasm struct {
	chainCtx ChainContext
}

// NewWasm returns new instance of the Wasm.
func NewWasm(chainCtx ChainContext) Wasm {
	return Wasm{
		chainCtx: chainCtx,
	}
}

// DeployAndInstantiateWASMContract deploys, instantiateWASMContract the wasm contract and returns its address.
func (w Wasm) DeployAndInstantiateWASMContract(ctx context.Context, txf client.Factory, fromAddress sdk.AccAddress, wasmData []byte, initConfig InstantiateConfig) (string, uint64, error) {
	codeID, err := w.DeployWASMContract(ctx, txf, fromAddress, wasmData)
	if err != nil {
		return "", 0, err
	}

	initConfig.CodeID = codeID
	contractAddr, err := w.InstantiateWASMContract(ctx, txf, fromAddress, initConfig)
	if err != nil {
		return "", 0, err
	}

	return contractAddr, codeID, nil
}

// ExecuteWASMContract executes the wasm contract with the payload and optionally funding amount.
func (w Wasm) ExecuteWASMContract(ctx context.Context, txf client.Factory, fromAddress sdk.AccAddress, contractAddr string, payload json.RawMessage, fundAmt sdk.Coin) (int64, error) {
	funds := sdk.NewCoins()
	if !fundAmt.Amount.IsNil() {
		funds = funds.Add(fundAmt)
	}

	msg := &wasmtypes.MsgExecuteContract{
		Sender:   w.chainCtx.MustConvertToBech32Address(fromAddress),
		Contract: contractAddr,
		Msg:      wasmtypes.RawContractMessage(payload),
		Funds:    funds,
	}

	res, err := w.chainCtx.BroadcastTxWithSigner(ctx, addGasMultiplier(txf), fromAddress, msg)
	if err != nil {
		return 0, err
	}
	return res.GasUsed, nil
}

// QueryWASMContract queries the contract with the requested payload.
func (w Wasm) QueryWASMContract(ctx context.Context, contractAddr string, payload json.RawMessage) (json.RawMessage, error) {
	query := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddr,
		QueryData: wasmtypes.RawContractMessage(payload),
	}

	wasmClient := wasmtypes.NewQueryClient(w.chainCtx.ClientContext)
	resp, err := wasmClient.SmartContractState(ctx, query)
	if err != nil {
		return nil, sdkerrors.Wrap(err, "WASMQueryClient returned an error after smart contract state queryWASMContract")
	}

	return json.RawMessage(resp.Data), nil
}

// DeployWASMContract the wasm contract and returns its codeID.
func (w Wasm) DeployWASMContract(ctx context.Context, txf client.Factory, fromAddress sdk.AccAddress, wasmData []byte) (uint64, error) {
	msg := &wasmtypes.MsgStoreCode{
		Sender:       w.chainCtx.MustConvertToBech32Address(fromAddress),
		WASMByteCode: wasmData,
	}

	res, err := w.chainCtx.BroadcastTxWithSigner(ctx, addGasMultiplier(txf), fromAddress, msg)
	if err != nil {
		return 0, err
	}

	codeID, err := event.FindUint64EventAttribute(res.Events, wasmtypes.EventTypeStoreCode, wasmtypes.AttributeKeyCodeID)
	if err != nil {
		return 0, err
	}

	return codeID, nil
}

// InstantiateWASMContract instantiates the contract and returns the contract address.
func (w Wasm) InstantiateWASMContract(ctx context.Context, txf client.Factory, fromAddress sdk.AccAddress, req InstantiateConfig) (string, error) {
	funds := sdk.NewCoins()
	if amount := req.Amount; !amount.Amount.IsNil() {
		funds = funds.Add(amount)
	}

	msg := &wasmtypes.MsgInstantiateContract{
		Sender: w.chainCtx.MustConvertToBech32Address(fromAddress),
		Admin: func() string {
			if req.Admin != nil {
				return w.chainCtx.MustConvertToBech32Address(req.Admin)
			}
			return ""
		}(),
		CodeID: req.CodeID,
		Label:  req.Label,
		Msg:    wasmtypes.RawContractMessage(req.Payload),
		Funds:  funds,
	}

	res, err := w.chainCtx.BroadcastTxWithSigner(ctx, addGasMultiplier(txf), fromAddress, msg)
	if err != nil {
		return "", err
	}

	contractAddr, err := event.FindStringEventAttribute(res.Events, wasmtypes.EventTypeInstantiate, wasmtypes.AttributeKeyContractAddr)
	if err != nil {
		return "", err
	}

	return contractAddr, nil
}

// IsWASMContractPinned returns true if smart contract is pinned.
func (w Wasm) IsWASMContractPinned(ctx context.Context, codeID uint64) (bool, error) {
	wasmClient := wasmtypes.NewQueryClient(w.chainCtx.ClientContext)
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

// MigrateWASMContract migrates the wasm contract.
func (w Wasm) MigrateWASMContract(
	ctx context.Context,
	txf client.Factory,
	fromAddress sdk.AccAddress,
	contractAddress string,
	codeID uint64,
	payload json.RawMessage,
) error {
	msg := &wasmtypes.MsgMigrateContract{
		Sender:   w.chainCtx.MustConvertToBech32Address(fromAddress),
		Contract: contractAddress,
		CodeID:   codeID,
		Msg:      wasmtypes.RawContractMessage(payload),
	}

	_, err := w.chainCtx.BroadcastTxWithSigner(ctx, addGasMultiplier(txf), fromAddress, msg)
	if err != nil {
		return err
	}

	return nil
}

func addGasMultiplier(txf client.Factory) client.Factory {
	if txf.Gas() == 0 {
		return txf.WithGasAdjustment(gasMultiplier)
	}

	return txf
}
