package client

import (
	"context"
	"encoding/hex"
	"regexp"
	"strconv"
	"strings"
	"time"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"
	"github.com/samber/lo"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
	feemodeltypes "github.com/CoreumFoundation/coreum/x/feemodel/types"
)

const (
	requestTimeout       = 10 * time.Second
	txTimeout            = time.Minute
	txStatusPollInterval = 500 * time.Millisecond
)

var expectedSequenceRegExp = regexp.MustCompile(`account sequence mismatch, expected (\d+), got \d+`)

// Client is the client for cored blockchain
type Client struct {
	clientCtx           client.Context
	authQueryClient     authtypes.QueryClient
	bankQueryClient     banktypes.QueryClient
	wasmQueryClient     wasmtypes.QueryClient
	feemodelQueryClient feemodeltypes.QueryClient
}

// New creates new client for cored
func New(chainID config.ChainID, addr string) Client {
	clientProtocols := []string{"tcp", "http", "https"}
	if !lo.ContainsBy(clientProtocols, func(protocol string) bool {
		return strings.HasPrefix(addr, protocol+"://")
	}) {
		panic(errors.Errorf("the address %q contains not supported protocol, supported are: %q", addr, clientProtocols))
	}

	rpcClient, err := client.NewClientFromNode(addr)
	must.OK(err)
	// This line takes `app.ModuleBasics` but this code will be removed anyway
	clientCtx := config.NewClientContext(app.ModuleBasics).
		WithChainID(string(chainID)).
		WithClient(rpcClient)
	return Client{
		clientCtx:           clientCtx,
		authQueryClient:     authtypes.NewQueryClient(clientCtx),
		bankQueryClient:     banktypes.NewQueryClient(clientCtx),
		wasmQueryClient:     wasmtypes.NewQueryClient(clientCtx),
		feemodelQueryClient: feemodeltypes.NewQueryClient(clientCtx),
	}
}

// GetClientCtx returns the clientCtx from the client.
// TODO (dhil): this is temp workaround to get access to the configured client context util we migrate to new tx package
func (c Client) GetClientCtx() client.Context {
	return c.clientCtx
}

// GetNumberSequence returns account number and account sequence for provided address
func (c Client) GetNumberSequence(ctx context.Context, address string) (uint64, uint64, error) {
	addr, err := sdk.AccAddressFromBech32(address)
	must.OK(err)

	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	var header metadata.MD
	res, err := c.authQueryClient.Account(requestCtx, &authtypes.QueryAccountRequest{Address: addr.String()}, grpc.Header(&header))
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}

	var acc authtypes.AccountI
	if err := c.clientCtx.InterfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
		return 0, 0, errors.WithStack(err)
	}

	return acc.GetAccountNumber(), acc.GetSequence(), nil
}

// QueryBankBalances queries for bank balances owned by wallet
func (c Client) QueryBankBalances(ctx context.Context, wallet types.Wallet) (map[string]sdk.Coin, error) {
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	// FIXME (wojtek): support pagination
	resp, err := c.bankQueryClient.AllBalances(requestCtx, &banktypes.QueryAllBalancesRequest{Address: wallet.Key.Address()})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	balances := map[string]sdk.Coin{}
	for _, b := range resp.Balances {
		balances[b.Denom] = b
	}
	return balances, nil
}

// Sign takes message, creates transaction and signs it
func (c Client) Sign(ctx context.Context, input tx.BaseInput, msgs ...sdk.Msg) (authsigning.Tx, error) {
	signer := input.Signer
	if signer.AccountNumber == 0 && signer.AccountSequence == 0 {
		var err error
		signer.AccountNumber, signer.AccountSequence, err = c.GetNumberSequence(ctx, signer.Key.Address())
		if err != nil {
			return nil, err
		}

		input.Signer = signer
	}

	return tx.Sign(c.clientCtx, input, msgs...)
}

// Encode encodes transaction to be broadcasted
func (c Client) Encode(signedTx authsigning.Tx) []byte {
	return must.Bytes(c.clientCtx.TxConfig.TxEncoder()(signedTx))
}

// BroadcastResult contains results of transaction broadcast
type BroadcastResult struct {
	TxHash    string
	GasUsed   uint64
	EventLogs sdk.StringEvents
}

// Broadcast broadcasts encoded transaction and returns tx hash
func (c Client) Broadcast(ctx context.Context, encodedTx []byte) (BroadcastResult, error) {
	var txHash string
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	res, err := c.clientCtx.Client.BroadcastTxSync(requestCtx, encodedTx)
	if err != nil {
		if errors.Is(err, requestCtx.Err()) {
			return BroadcastResult{}, errors.WithStack(err)
		}

		errRes := client.CheckTendermintError(err, encodedTx)
		if !isTxInMempool(errRes) {
			return BroadcastResult{}, errors.WithStack(err)
		}
		txHash = errRes.TxHash
	} else {
		txHash = res.Hash.String()
		if res.Code != 0 {
			return BroadcastResult{}, errors.Wrapf(cosmoserrors.New(res.Codespace, res.Code, res.Log),
				"transaction '%s' failed", txHash)
		}
	}

	txHashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		return BroadcastResult{}, errors.WithStack(err)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, txTimeout)
	defer cancel()

	var resultTx *coretypes.ResultTx
	err = retry.Do(timeoutCtx, txStatusPollInterval, func() error {
		requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()

		var err error
		resultTx, err = c.clientCtx.Client.Tx(requestCtx, txHashBytes, false)
		if err != nil {
			if errors.Is(err, requestCtx.Err()) {
				return retry.Retryable(errors.WithStack(err))
			}
			if errRes := client.CheckTendermintError(err, encodedTx); errRes != nil {
				if isTxInMempool(errRes) {
					return retry.Retryable(errors.WithStack(err))
				}
				return errors.WithStack(err)
			}
			return retry.Retryable(errors.WithStack(err))
		}
		if resultTx.TxResult.Code != 0 {
			res := resultTx.TxResult
			return errors.Wrapf(cosmoserrors.New(res.Codespace, res.Code, res.Log), "transaction '%s' failed", txHash)
		}
		if resultTx.Height == 0 {
			return retry.Retryable(errors.Errorf("transaction '%s' hasn't been included in a block yet", txHash))
		}
		return nil
	})
	if err != nil {
		return BroadcastResult{}, err
	}
	return BroadcastResult{
		TxHash:    txHash,
		GasUsed:   uint64(resultTx.TxResult.GasUsed),
		EventLogs: sdk.StringifyEvents(resultTx.TxResult.Events),
	}, nil
}

// EstimateGas runs the transaction cost estimation.
func (c Client) EstimateGas(ctx context.Context, input tx.BaseInput, msgs ...sdk.Msg) (uint64, error) {
	signer := input.Signer
	if signer.AccountNumber == 0 && signer.AccountSequence == 0 {
		var err error
		signer.AccountNumber, signer.AccountSequence, err = c.GetNumberSequence(ctx, signer.Key.Address())
		if err != nil {
			return 0, err
		}

		input.Signer = signer
	}

	simTxBytes, err := tx.BuildSimTx(c.clientCtx, input, msgs...)
	if err != nil {
		err = errors.Wrap(err, "failed to build sim tx bytes")
		return 0, err
	}

	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	txSvcClient := txtypes.NewServiceClient(c.clientCtx)
	simRes, err := txSvcClient.Simulate(requestCtx, &txtypes.SimulateRequest{
		TxBytes: simTxBytes,
	})
	if err != nil {
		return 0, errors.Wrap(err, "failed to simulate the transaction execution")
	}

	return simRes.GasInfo.GasUsed, nil
}

// TxBankSendInput holds input data for PrepareTxBankSend
type TxBankSendInput struct {
	Sender   types.Wallet
	Receiver types.Wallet
	Amount   sdk.Coin

	Base tx.BaseInput
}

// PrepareTxBankSend creates a transaction sending tokens from one wallet to another
func (c Client) PrepareTxBankSend(ctx context.Context, input TxBankSendInput) ([]byte, error) {
	fromAddress, err := sdk.AccAddressFromBech32(input.Sender.Key.Address())
	must.OK(err)
	toAddress, err := sdk.AccAddressFromBech32(input.Receiver.Key.Address())
	must.OK(err)

	if err := input.Amount.Validate(); err != nil {
		return nil, errors.Wrap(err, "amount to send is invalid")
	}

	signedTx, err := c.Sign(ctx, input.Base, banktypes.NewMsgSend(fromAddress, toAddress, sdk.Coins{input.Amount}))
	if err != nil {
		return nil, err
	}

	return c.Encode(signedTx), nil
}

// BankQueryClient returns a Bank module querying client, initialized
// using the internal clientCtx.
func (c Client) BankQueryClient() banktypes.QueryClient {
	return c.bankQueryClient
}

// WASMQueryClient returns a WASM module querying client, initialized
// using the internal clientCtx.
func (c Client) WASMQueryClient() wasmtypes.QueryClient {
	return c.wasmQueryClient
}

// FeemodelQueryClient returns a feemodel module querying client, initialized
// using the internal clientCtx.
func (c Client) FeemodelQueryClient() feemodeltypes.QueryClient {
	return c.feemodelQueryClient
}

func isTxInMempool(errRes *sdk.TxResponse) bool {
	if errRes == nil {
		return false
	}
	return isSDKErrorResult(errRes.Codespace, errRes.Code, cosmoserrors.ErrTxInMempoolCache)
}

func isSDKErrorResult(codespace string, code uint32, expectedSDKError *cosmoserrors.Error) bool {
	return codespace == expectedSDKError.Codespace() &&
		code == expectedSDKError.ABCICode()
}

func asSDKError(err error, expectedSDKErr *cosmoserrors.Error) *cosmoserrors.Error {
	var sdkErr *cosmoserrors.Error
	if !errors.As(err, &sdkErr) || !isSDKErrorResult(sdkErr.Codespace(), sdkErr.ABCICode(), expectedSDKErr) {
		return nil
	}
	return sdkErr
}

// ExpectedSequenceFromError checks if error is related to account sequence mismatch, and returns expected account sequence
func ExpectedSequenceFromError(err error) (uint64, bool, error) {
	sdkErr := asSDKError(err, cosmoserrors.ErrWrongSequence)
	if sdkErr == nil {
		return 0, false, nil
	}

	log := sdkErr.Error()
	matches := expectedSequenceRegExp.FindStringSubmatch(log)
	if len(matches) != 2 {
		return 0, false, errors.Errorf("cosmos sdk hasn't returned expected sequence number, log mesage received: %s", log)
	}
	expectedSequence, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, false, errors.Wrapf(err, "can't parse expected sequence number, log mesage received: %s", log)
	}
	return expectedSequence, true, nil
}

// IsErr returns true if error was caused by insufficient funds provided with the transaction
func IsErr(err error, cosmosErr *cosmoserrors.Error) bool {
	return asSDKError(err, cosmosErr) != nil
}

// FindEventAttribute finds the first event attribute by type and attribute name.
func FindEventAttribute(event sdk.StringEvents, etype, attribute string) (string, bool) {
	for _, ev := range event {
		if ev.Type == etype {
			if value, found := findAttribute(ev, attribute); found {
				return value, true
			}
		}
	}
	return "", false
}

func findAttribute(ev sdk.StringEvent, attr string) (string, bool) {
	for _, attrItem := range ev.Attributes {
		if attrItem.Key == attr {
			return attrItem.Value, true
		}
	}

	return "", false
}
