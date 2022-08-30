package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	tmtypes "github.com/tendermint/tendermint/types"
	"go.uber.org/zap"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// ExecuteConfig contains contract execution arguments and options.
type ExecuteConfig struct {
	// Network holds the chain config of the network
	Network ChainConfig

	// From specifies credentials for signing the execution transactions.
	From types.Wallet

	// ExecutePayload is a path to a file containing JSON-encoded contract exec args, or JSON-encoded body itself.
	ExecutePayload string

	// Amount specifies Coins to send to the contract during execution.
	Amount types.Coin

	executePayloadBody json.RawMessage
}

// ExecuteOutput contains the results of the contract method execution.
type ExecuteOutput struct {
	ContractAddress string `json:"contractAddress"`
	MethodExecuted  string `json:"methodExecuted"`
	ExecuteTxHash   string `json:"execTxHash"`
}

// Execute implements logic for "contracts exec" CLI command.
func Execute(ctx context.Context, contractAddr string, config ExecuteConfig) (*ExecuteOutput, error) {
	log := logger.Get(ctx)

	if len(contractAddr) == 0 {
		err := errors.New("contract address cannot be empty")
		return nil, err
	} else if err := config.ValidateAndLoad(); err != nil {
		return nil, errors.Wrap(err, "failed to validate the execution config")
	}

	out := &ExecuteOutput{
		ContractAddress: contractAddr,
	}
	log.Sugar().
		With(zap.String("from", config.From.Address().String())).
		Infof("Executing %s on chain", contractAddr)

	methodName, execTxHash, err := runContractExecution(
		ctx,
		config.Network,
		config.From,
		contractAddr,
		config.executePayloadBody,
		config.Amount,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed to run contract execution")
	}

	out.MethodExecuted = methodName
	out.ExecuteTxHash = execTxHash

	return out, nil
}

func runContractExecution(
	ctx context.Context,
	network ChainConfig,
	from types.Wallet,
	contractAddr string,
	execMsg json.RawMessage,
	amount types.Coin,
) (methodName, txHash string, err error) {
	log := logger.Get(ctx)
	chainClient := network.Client

	input := tx.BaseInput{
		Signer:   from,
		GasPrice: network.MinGasPrice,
	}

	funds := sdk.NewCoins()
	if amount.Amount != nil {
		funds = funds.Add(sdk.NewCoin(amount.Denom, sdk.NewIntFromBigInt(amount.Amount)))
	}
	msgExecuteContract := &wasmtypes.MsgExecuteContract{
		Sender:   from.Address().String(),
		Contract: contractAddr,
		Msg:      wasmtypes.RawContractMessage(execMsg),
		Funds:    funds,
	}

	gasLimit, err := chainClient.EstimateGas(ctx, input, msgExecuteContract)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to estimate gas for MsgExecuteContract")
	}

	log.Info("Estimated gas limit",
		zap.Int("contract_msg_size", len(execMsg)),
		zap.Uint64("gas_limit", gasLimit),
	)

	input.GasLimit = uint64(float64(gasLimit) * gasMultiplier)

	signedTx, err := chainClient.Sign(ctx, input, msgExecuteContract)
	if err != nil {
		return "", "", errors.Wrapf(err, "failed to sign transaction as %s", from.Address().String())
	}

	txBytes := chainClient.Encode(signedTx)
	txHash = fmt.Sprintf("%X", tmtypes.Tx(txBytes).Hash())
	res, err := chainClient.Broadcast(ctx, txBytes)
	if err != nil {
		return "", txHash, errors.Wrapf(err, "failed to broadcast Tx %s", txHash)
	}

	if len(res.EventLogs) > 0 {
		client.LogEventLogsInfo(log, res.EventLogs)
	}

	for _, ev := range res.EventLogs {
		if ev.Type == wasmtypes.WasmModuleEventType {
			if value, ok := attrFromEvent(ev, "method"); ok {
				methodName = value
				break
			}
		}
	}

	return methodName, txHash, nil
}

// ValidateAndLoad validates the contract execution config and loads it's initial state.
// TODO(dhil) it would be better not to sore the state in the config and not set in the validation.
func (c *ExecuteConfig) ValidateAndLoad() error {
	if body := []byte(c.ExecutePayload); json.Valid(body) {
		c.executePayloadBody = body
	} else {
		payloadFilePath := c.ExecutePayload

		body, err := os.ReadFile(payloadFilePath)
		if err != nil {
			return errors.Wrapf(err, "file specified for exec payload, but couldn't be read: %s", payloadFilePath)
		}

		if !json.Valid(body) {
			return errors.Wrapf(err, "file specified for exec payload, but doesn't contain valid JSON: %s", payloadFilePath)
		}

		c.executePayloadBody = body
	}

	if c.Amount.Amount != nil {
		if err := c.Amount.Validate(); err != nil {
			return errors.Wrapf(err, "invalid Amount: %v", c.Amount)
		}
	}

	if err := c.Network.MinGasPrice.Validate(); err != nil {
		return errors.Wrapf(err, "invalid MinGasPrice: %v", c.Network.MinGasPrice)
	}

	return nil
}
