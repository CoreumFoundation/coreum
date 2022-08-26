package wasm

import (
	"context"
	"encoding/json"
	"io/ioutil"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/client"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/pkg/errors"
)

// QueryConfig contains contract execution arguments and options.
type QueryConfig struct {
	// Network holds the chain config of the network
	Network ChainConfig

	// QueryPayload is a path to a file containing JSON-encoded contract query args, or JSON-encoded body itself.
	QueryPayload string

	queryPayloadBody json.RawMessage
}

// QueryOutput contains results of contract querying.
type QueryOutput struct {
	ContractAddress string          `json:"contractAddress"`
	Result          json.RawMessage `json:"queryResult"`
}

// Query implements logic for "contracts query" CLI command.
func Query(ctx context.Context, contractAddr string, config QueryConfig) (*QueryOutput, error) {
	log := logger.Get(ctx)

	if len(contractAddr) == 0 {
		err := errors.New("contract address cannot be empty")
		return nil, err
	} else if err := config.Validate(); err != nil {
		err = errors.Wrap(err, "failed to validate the execution config")
		return nil, err
	}

	out := &QueryOutput{
		ContractAddress: contractAddr,
	}
	log.Sugar().Infof("Querying %s on chain", contractAddr)

	result, err := runContractQuery(
		ctx,
		config.Network,
		contractAddr,
		config.queryPayloadBody,
	)
	if err != nil {
		err = errors.Wrap(err, "failed to run contract query")
		return nil, err
	}

	out.Result = result
	return out, nil
}

func runContractQuery(
	ctx context.Context,
	network ChainConfig,
	contractAddr string,
	queryMsg json.RawMessage,
) (result json.RawMessage, err error) {
	chainClient := client.New(app.ChainID(network.ChainID), network.RPCEndpoint)

	query := &wasmtypes.QuerySmartContractStateRequest{
		Address:   contractAddr,
		QueryData: wasmtypes.RawContractMessage(queryMsg),
	}

	resp, err := chainClient.WASMQueryClient().SmartContractState(ctx, query)
	if err != nil {
		err := errors.Wrap(err, "WASMQueryClient returns an error after smart contract state query")
		return nil, err
	}

	return json.RawMessage(resp.Data), nil
}

// Validate validates the contract query method config.
func (c *QueryConfig) Validate() error {
	if body := []byte(c.QueryPayload); json.Valid(body) {
		c.queryPayloadBody = body
	} else {
		payloadFilePath := c.QueryPayload

		body, err := ioutil.ReadFile(payloadFilePath)
		if err != nil {
			err = errors.Wrapf(err, "file specified for exec payload, but couldn't be read: %s", payloadFilePath)
			return err
		}

		if !json.Valid(body) {
			err = errors.Wrapf(err, "file specified for exec payload, but doesn't contain valid JSON: %s", payloadFilePath)
			return err
		}

		c.queryPayloadBody = json.RawMessage(body)
	}

	return nil
}
