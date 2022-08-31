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
	"github.com/CoreumFoundation/coreum/pkg/types"
)

type testClient struct {
	baseInput     tx.BaseInput
	coredClient   client.Client
	gasMultiplier float64
}

func newWasmTestClient(baseInput tx.BaseInput, coredClient client.Client) *testClient {
	return &testClient{
		baseInput:     baseInput,
		coredClient:   coredClient,
		gasMultiplier: 1.5,
	}
}

// instantiateConfig contains params specific to contract instantiation.
type instantiateConfig struct {
	accessType wasmtypes.AccessType
	payload    json.RawMessage
	amount     types.Coin
	label      string
	CodeID     uint64
}

// deploys the wasm contract and returns its codeID.
func (c *testClient) deploy(ctx context.Context, wasmData []byte) (uint64, error) {
	msgStoreCode := &wasmtypes.MsgStoreCode{
		Sender:       c.baseInput.Signer.Address().String(),
		WASMByteCode: wasmData,
	}

	res, err := c.coredClient.SubmitMessage(ctx, c.baseInput, msgStoreCode, client.WithGasMultiplier(c.gasMultiplier))
	if err != nil {
		return 0, err
	}

	ok, codeIDStr := client.FindEventAttribute(res.EventLogs, wasmtypes.EventTypeStoreCode, wasmtypes.AttributeKeyCodeID)
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
func (c *testClient) instantiate(ctx context.Context, req instantiateConfig) (string, error) {
	funds := sdk.NewCoins()
	if amount := req.amount; amount.Amount != nil {
		funds = funds.Add(sdk.NewCoin(amount.Denom, sdk.NewIntFromBigInt(amount.Amount)))
	}
	msgInstantiateContract := &wasmtypes.MsgInstantiateContract{
		Sender: c.baseInput.Signer.Address().String(),
		CodeID: req.CodeID,
		Label:  req.label,
		Msg:    wasmtypes.RawContractMessage(req.payload),
		Funds:  funds,
	}

	res, err := c.coredClient.SubmitMessage(ctx, c.baseInput, msgInstantiateContract, client.WithGasMultiplier(c.gasMultiplier))
	if err != nil {
		return "", err
	}

	ok, contractAddr := client.FindEventAttribute(res.EventLogs, wasmtypes.EventTypeInstantiate, wasmtypes.AttributeKeyContractAddr)
	if !ok {
		return "", errors.New("can't find the contract address in the tx events")
	}

	return contractAddr, nil
}

// deployAndInstantiate deploys, instantiate the wasm contract and returns its address.
func (c *testClient) deployAndInstantiate(ctx context.Context, wasmData []byte, initConfig instantiateConfig) (string, error) {
	codeID, err := c.deploy(ctx, wasmData)
	if err != nil {
		return "", err
	}

	initConfig.CodeID = codeID
	contractAddr, err := c.instantiate(ctx, initConfig)
	if err != nil {
		return "", err
	}

	return contractAddr, nil
}

// executes the wasm contract with the payload and optionally funding amount.
func (c *testClient) execute(ctx context.Context, contractAddr string, payload json.RawMessage, fundAmt types.Coin) error {
	funds := sdk.NewCoins()
	if fundAmt.Amount != nil {
		funds = funds.Add(sdk.NewCoin(fundAmt.Denom, sdk.NewIntFromBigInt(fundAmt.Amount)))
	}
	msgExecuteContract := &wasmtypes.MsgExecuteContract{
		Sender:   c.baseInput.Signer.Address().String(),
		Contract: contractAddr,
		Msg:      wasmtypes.RawContractMessage(payload),
		Funds:    funds,
	}

	_, err := c.coredClient.SubmitMessage(ctx, c.baseInput, msgExecuteContract, client.WithGasMultiplier(c.gasMultiplier))
	if err != nil {
		return err
	}

	return nil
}

// queries the contract with the requested payload.
func (c *testClient) query(ctx context.Context, contractAddr string, payload json.RawMessage) (json.RawMessage, error) {
	query := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddr,
		QueryData: wasmtypes.RawContractMessage(payload),
	}

	resp, err := c.coredClient.WASMQueryClient().SmartContractState(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "WASMQueryClient returns an error after smart contract state query")
	}

	return json.RawMessage(resp.Data), nil
}
