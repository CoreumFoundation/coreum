package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/CoreumFoundation/coreum-tools/pkg/logger"
	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/client"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/pkg/errors"
	tmtypes "github.com/tendermint/tendermint/types"
	"go.uber.org/zap"
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
	Amount string

	amountParsed       sdk.Coins
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
	} else if err := config.Validate(); err != nil {
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
		config.amountParsed,
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
	amount sdk.Coins,
) (methodName, txHash string, err error) {
	log := logger.Get(ctx)
	chainClient := client.New(app.ChainID(network.ChainID), network.RPCEndpoint)

	input := tx.BaseInput{
		Signer:   from,
		GasPrice: network.minGasPriceParsed,
	}

	msgExecuteContract := &wasmtypes.MsgExecuteContract{
		Sender:   from.Address().String(),
		Contract: contractAddr,
		Msg:      wasmtypes.RawContractMessage(execMsg),
		Funds:    amount,
	}

	gasLimit, err := chainClient.EstimateGas(ctx, input, msgExecuteContract)
	if err != nil {
		return "", "", errors.Wrap(err, "failed to estimate gas for MsgExecuteContract")
	}

	log.Info("Estimated gas limit",
		zap.Int("contract_msg_size", len(execMsg)),
		zap.Uint64("gas_limit", gasLimit),
	)

	input.GasLimit = uint64(float64(gasLimit) * GasMultiplier)

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

// Validate validates the contract execution config. May load some data from disk.
// FIXME validate mustn't update the state
func (c *ExecuteConfig) Validate() error {
	if body := []byte(c.ExecutePayload); json.Valid(body) {
		c.executePayloadBody = body
	} else {
		payloadFilePath := c.ExecutePayload

		body, err := ioutil.ReadFile(payloadFilePath)
		if err != nil {
			return errors.Wrapf(err, "file specified for exec payload, but couldn't be read: %s", payloadFilePath)
		}

		if !json.Valid(body) {
			return errors.Wrapf(err, "file specified for exec payload, but doesn't contain valid JSON: %s", payloadFilePath)
		}

		c.executePayloadBody = json.RawMessage(body)
	}

	if len(c.Amount) > 0 {
		amount, err := sdk.ParseCoinsNormalized(c.Amount)
		if err != nil {
			return errors.Wrapf(err, "failed to parse exec transfer amount as sdk.Coins: %s", c.Amount)
		}

		c.amountParsed = amount
	}

	if len(c.Network.MinGasPrice) == 0 {
		c.Network.minGasPriceParsed = types.Coin{
			// FIXME take it as usual tests does
			Amount: big.NewInt(1500), // matches InitialGasPrice in cored
			Denom:  "core",
		}
	} else {
		coinValue, err := sdk.ParseCoinNormalized(c.Network.MinGasPrice)
		if err != nil {
			return errors.Wrapf(err, "failed to parse min gas price coin spec as sdk.Coin: %s", c.Network.MinGasPrice)
		}

		c.Network.minGasPriceParsed = types.Coin{
			Amount: coinValue.Amount.BigInt(),
			Denom:  coinValue.Denom,
		}
	}

	return nil
}
