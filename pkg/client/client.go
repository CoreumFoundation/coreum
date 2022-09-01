package client

import (
	"context"
	"regexp"
	"strconv"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

const (
	requestTimeout = 10 * time.Second
)

var expectedSequenceRegExp = regexp.MustCompile(`account sequence mismatch, expected (\d+), got \d+`)

// New creates new client for cored
func New(chainID app.ChainID, addr string) Client {
	rpcClient, err := client.NewClientFromNode("tcp://" + addr)
	must.OK(err)
	clientCtx := app.
		NewDefaultClientContext().
		WithChainID(string(chainID)).
		WithClient(rpcClient)
	return Client{
		clientCtx:       clientCtx,
		authQueryClient: authtypes.NewQueryClient(clientCtx),
		bankQueryClient: banktypes.NewQueryClient(clientCtx),
	}
}

// Client is the client for cored blockchain
type Client struct {
	clientCtx       client.Context
	authQueryClient authtypes.QueryClient
	bankQueryClient banktypes.QueryClient
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
func (c Client) QueryBankBalances(ctx context.Context, wallet types.Wallet) (map[string]types.Coin, error) {
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	// FIXME (wojtek): support pagination
	resp, err := c.bankQueryClient.AllBalances(requestCtx, &banktypes.QueryAllBalancesRequest{Address: wallet.Key.Address()})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	balances := map[string]types.Coin{}
	for _, b := range resp.Balances {
		coin, err := types.NewCoin(b.Amount.BigInt(), b.Denom)
		if err != nil {
			return nil, err
		}
		balances[b.Denom] = coin
	}
	return balances, nil
}

// Encode encodes transaction to be broadcasted
func (c Client) Encode(signedTx authsigning.Tx) []byte {
	return must.Bytes(c.clientCtx.TxConfig.TxEncoder()(signedTx))
}

// Context returns the underlying cosmos client.Context types
func (c Client) Context() client.Context {
	return c.clientCtx
}

// BroadcastResult contains results of transaction broadcast
type BroadcastResult struct {
	TxHash  string
	GasUsed int64
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
		return 0, false, errors.Errorf("cosmos sdk hasn't returned expected sequence number, log message received: %s", log)
	}
	expectedSequence, err := strconv.ParseUint(matches[1], 10, 64)
	if err != nil {
		return 0, false, errors.Wrapf(err, "can't parse expected sequence number, log message received: %s", log)
	}
	return expectedSequence, true, nil
}

// IsInsufficientFeeError returns true if error was caused by insufficient fee provided with the transaction
func IsInsufficientFeeError(err error) bool {
	return asSDKError(err, cosmoserrors.ErrInsufficientFee) != nil
}
