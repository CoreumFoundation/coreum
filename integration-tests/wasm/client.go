package wasm

import (
	"context"
	"encoding/json"
	"strconv"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
)

// Client represents the wasm client with the helper functions used for the testing.
type Client struct {
	coredClient   client.Client
	gasMultiplier float64
}

// NewClient new wasm client.
func NewClient(coredClient client.Client) *Client {
	return &Client{
		coredClient:   coredClient,
		gasMultiplier: 1.5,
	}
}

// InstantiateConfig contains params specific to contract instantiation.
type InstantiateConfig struct {
	accessType wasmtypes.AccessType
	payload    json.RawMessage
	amount     sdk.Coin
	label      string
	CodeID     uint64
}

// DeployAndInstantiate deploys, instantiate the wasm contract and returns its address.
func (c *Client) DeployAndInstantiate(ctx context.Context, baseInput tx.BaseInput, wasmData []byte, initConfig InstantiateConfig) (string, error) {
	codeID, err := c.deploy(ctx, baseInput, wasmData)
	if err != nil {
		return "", err
	}

	initConfig.CodeID = codeID
	contractAddr, err := c.instantiate(ctx, baseInput, initConfig)
	if err != nil {
		return "", err
	}

	return contractAddr, nil
}

// Execute executes the wasm contract with the payload and optionally funding amount.
func (c *Client) Execute(ctx context.Context, baseInput tx.BaseInput, contractAddr string, payload json.RawMessage, fundAmt sdk.Coin) error {
	funds := sdk.NewCoins()
	if !fundAmt.Amount.IsNil() {
		funds = funds.Add(fundAmt)
	}

	if _, err := c.submitWithEstimatedGasLimit(ctx, baseInput, &wasmtypes.MsgExecuteContract{
		Sender:   baseInput.Signer.Address().String(),
		Contract: contractAddr,
		Msg:      wasmtypes.RawContractMessage(payload),
		Funds:    funds,
	}); err != nil {
		return err
	}

	return nil
}

// Query queries the contract with the requested payload.
func (c *Client) Query(ctx context.Context, contractAddr string, payload json.RawMessage) (json.RawMessage, error) {
	query := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddr,
		QueryData: wasmtypes.RawContractMessage(payload),
	}

	resp, err := c.coredClient.WASMQueryClient().SmartContractState(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "WASMQueryClient returns an error after smart contract state Query")
	}

	return json.RawMessage(resp.Data), nil
}

// deploys the wasm contract and returns its codeID.
func (c *Client) deploy(ctx context.Context, baseInput tx.BaseInput, wasmData []byte) (uint64, error) {
	msgStoreCode := &wasmtypes.MsgStoreCode{
		Sender:       baseInput.Signer.Address().String(),
		WASMByteCode: wasmData,
	}

	res, err := c.submitWithEstimatedGasLimit(ctx, baseInput, msgStoreCode)
	if err != nil {
		return 0, err
	}

	codeIDStr, ok := client.FindEventAttribute(res.EventLogs, wasmtypes.EventTypeStoreCode, wasmtypes.AttributeKeyCodeID)
	if !ok {
		return 0, errors.New("can't find the codeID in the tx events")
	}
	codeID, err := strconv.ParseUint(codeIDStr, 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse event attribute CodeID: %s as uint64", codeIDStr)
	}

	return codeID, nil
}

// instantiates the contract and returns the contract address.
func (c *Client) instantiate(ctx context.Context, baseInput tx.BaseInput, req InstantiateConfig) (string, error) {
	funds := sdk.NewCoins()
	if amount := req.amount; !amount.Amount.IsNil() {
		funds = funds.Add(amount)
	}
	msgInstantiateContract := &wasmtypes.MsgInstantiateContract{
		Sender: baseInput.Signer.Address().String(),
		CodeID: req.CodeID,
		Label:  req.label,
		Msg:    wasmtypes.RawContractMessage(req.payload),
		Funds:  funds,
	}

	res, err := c.submitWithEstimatedGasLimit(ctx, baseInput, msgInstantiateContract)
	if err != nil {
		return "", err
	}

	contractAddr, ok := client.FindEventAttribute(res.EventLogs, wasmtypes.EventTypeInstantiate, wasmtypes.AttributeKeyContractAddr)
	if !ok {
		return "", errors.New("can't find the contract address in the tx events")
	}

	return contractAddr, nil
}

// submitWithEstimatedGasLimit is a combination of EstimateGas, Sign and Broadcast methods.
func (c *Client) submitWithEstimatedGasLimit(ctx context.Context, input tx.BaseInput, msg sdk.Msg) (client.BroadcastResult, error) {
	if input.GasLimit == 0 {
		gasLimit, err := c.coredClient.EstimateGas(ctx, input, msg)
		if err != nil {
			return client.BroadcastResult{}, err
		}
		input.GasLimit = gasLimit
	}
	input.GasLimit = uint64(float64(input.GasLimit) * c.gasMultiplier)
	signedTx, err := c.coredClient.Sign(ctx, input, msg)
	if err != nil {
		return client.BroadcastResult{}, err
	}

	return c.coredClient.Broadcast(ctx, c.coredClient.Encode(signedTx))
}
