package client2

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"time"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	"github.com/cosmos/cosmos-sdk/client"
	cosmosed25519 "github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cosmossecp256k1 "github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/CoreumFoundation/coreum/app"
	"github.com/CoreumFoundation/coreum/pkg/tx2"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

const (
	requestTimeout       = 10 * time.Second
	txTimeout            = time.Minute
	txStatusPollInterval = 500 * time.Millisecond
)

var expectedSequenceRegExp = regexp.MustCompile(`account sequence mismatch, expected (\d+), got \d+`)

// New creates new client for cored
func New(chainID app.ChainID, addr string) (Client, error) {
	parsedURL, err := url.Parse(addr)
	if err != nil {
		return Client{}, errors.WithStack(err)
	}
	switch parsedURL.Scheme {
	case "tcp", "http", "https":
	default:
		return Client{}, errors.Errorf("unknown scheme '%s' in address", parsedURL.Scheme)
	}
	rpcClient, err := client.NewClientFromNode(addr)
	if err != nil {
		return Client{}, errors.WithStack(err)
	}

	clientCtx := app.
		NewDefaultClientContext().
		WithChainID(string(chainID)).
		WithClient(rpcClient)

	return Client{
		clientCtx:       clientCtx,
		txServiceClient: txtypes.NewServiceClient(clientCtx),
		authQueryClient: authtypes.NewQueryClient(clientCtx),
		bankQueryClient: banktypes.NewQueryClient(clientCtx),
	}, nil
}

// Client is the client for cored blockchain
type Client struct {
	clientCtx       client.Context
	txServiceClient txtypes.ServiceClient
	authQueryClient authtypes.QueryClient
	bankQueryClient banktypes.QueryClient
}

// GetNumberSequence returns account number and account sequence for provided address
func (c Client) GetNumberSequence(ctx context.Context, address types.Address) (uint64, uint64, error) {
	addr, err := sdk.AccAddressFromBech32(string(address))
	if err != nil {
		return 0, 0, errors.WithStack(err)
	}

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
func (c Client) QueryBankBalances(ctx context.Context, address types.Address) (map[string]types.Coin, error) {
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	// FIXME (wojtek): support pagination
	resp, err := c.bankQueryClient.AllBalances(requestCtx, &banktypes.QueryAllBalancesRequest{Address: string(address)})
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

// BroadcastResult contains results of transaction broadcast
type BroadcastResult struct {
	TxHash  string
	GasUsed int64
}

// EstimateGas runs the transaction cost estimation and returns new suggested gas limit,
// in contrast with the default Cosmos SDK gas estimation logic, this method returns unadjusted gas used.
func (c Client) EstimateGas(ctx context.Context, encodedTx []byte) (int64, error) {
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	simRes, err := c.txServiceClient.Simulate(requestCtx, &txtypes.SimulateRequest{
		TxBytes: encodedTx,
	})
	if err != nil {
		return 0, errors.Wrap(err, "failed to simulate the transaction execution")
	}

	// usually gas has to be multiplied by some adjustment coefficient: e.g. *1.5
	// but in this case we return unadjusted, so every module can decide the adjustment value
	return int64(simRes.GasInfo.GasUsed), nil
}

// BroadcastSync broadcasts encoded transaction, waits until it is included in a block and returns tx hash
func (c Client) BroadcastSync(ctx context.Context, encodedTx []byte) (BroadcastResult, error) {
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
	} else if res.Code != 0 {
		txHash := fmt.Sprintf("%X", tmtypes.Tx(encodedTx).Hash())
		return BroadcastResult{}, errors.Wrapf(cosmoserrors.New(res.Codespace, res.Code, res.Log),
			"transaction '%s' failed", txHash)
	}

	return c.AwaitTx(ctx, encodedTx)
}

// BroadcastAsync broadcasts encoded transaction and returns tx hash immediately,
// it doesn't even wait for CheckTx being run for transaction for maximum throughput
func (c Client) BroadcastAsync(ctx context.Context, encodedTx []byte) (BroadcastResult, error) {
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	res, err := c.clientCtx.Client.BroadcastTxAsync(requestCtx, encodedTx)
	if err != nil {
		return BroadcastResult{}, errors.WithStack(err)
	}

	txHash := fmt.Sprintf("%X", tmtypes.Tx(encodedTx).Hash())
	if res.Code != 0 {
		return BroadcastResult{}, errors.Wrapf(cosmoserrors.New(res.Codespace, res.Code, res.Log),
			"transaction '%s' failed", txHash)
	}

	return BroadcastResult{
		TxHash: txHash,
	}, nil
}

// AwaitTx waits until transaction is included in a block
func (c Client) AwaitTx(ctx context.Context, encodedTx []byte) (BroadcastResult, error) {
	txHashBytes := tmtypes.Tx(encodedTx).Hash()
	txHash := fmt.Sprintf("%X", txHashBytes)

	timeoutCtx, cancel := context.WithTimeout(ctx, txTimeout)
	defer cancel()

	var resultTx *coretypes.ResultTx
	err := retry.Do(timeoutCtx, txStatusPollInterval, func() error {
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
		TxHash:  txHash,
		GasUsed: resultTx.TxResult.GasUsed,
	}, nil
}

// TxBankSendInput holds input data for PrepareTxBankSend
type TxBankSendInput struct {
	Sender   types.Address
	Receiver types.Address
	Amount   types.Coin

	Base tx2.BaseInput
}

// PrepareTxBankSend creates a transaction sending tokens from one wallet to another
func (c Client) PrepareTxBankSend(ctx context.Context, input TxBankSendInput) ([]byte, error) {
	fromAddress, err := sdk.AccAddressFromBech32(string(input.Sender))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	toAddress, err := sdk.AccAddressFromBech32(string(input.Receiver))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err := input.Amount.Validate(); err != nil {
		return nil, errors.Wrap(err, "amount to send is invalid")
	}

	encodedTx, err := c.prepareTx(ctx, input.Base, banktypes.NewMsgSend(fromAddress, toAddress, sdk.Coins{
		{
			Denom:  input.Amount.Denom,
			Amount: sdk.NewIntFromBigInt(input.Amount.Amount),
		},
	}))
	if err != nil {
		return nil, err
	}

	return encodedTx, err
}

// TxStakingCreateValidatorInput holds input data for PrepareTxStakingCreateValidator
type TxStakingCreateValidatorInput struct {
	ValidatorPublicKey ed25519.PublicKey
	StakedBalance      types.Coin

	Base tx2.BaseInput
}

// PrepareTxStakingCreateValidator creates a transaction adding validator
func (c Client) PrepareTxStakingCreateValidator(ctx context.Context, input TxStakingCreateValidatorInput) ([]byte, error) {
	amount, err := sdk.ParseCoinNormalized(input.StakedBalance.String())
	if err != nil {
		return nil, errors.Wrapf(err, "not able to parse stake balances %s", input.StakedBalance)
	}

	commission := stakingtypes.CommissionRates{
		Rate:          sdk.MustNewDecFromStr("0.1"),
		MaxRate:       sdk.MustNewDecFromStr("0.2"),
		MaxChangeRate: sdk.MustNewDecFromStr("0.01"),
	}

	valPubKey := &cosmosed25519.PubKey{Key: input.ValidatorPublicKey}
	stakerPubKey := &cosmossecp256k1.PubKey{Key: input.Base.Signer.PublicKey}
	stakerAddress := sdk.AccAddress(stakerPubKey.Address())

	msg, err := stakingtypes.NewMsgCreateValidator(sdk.ValAddress(stakerAddress), valPubKey, amount, stakingtypes.Description{Moniker: stakerAddress.String()}, commission, sdk.OneInt())
	if err != nil {
		return nil, errors.Wrap(err, "not able to make CreateValidatorMessage")
	}

	encodedTx, err := c.prepareTx(ctx, input.Base, msg)
	if err != nil {
		return nil, err
	}

	return encodedTx, err
}

// prepareTx includes messages in a new transaction then signs and encodes it
func (c Client) prepareTx(ctx context.Context, input tx2.BaseInput, msgs ...sdk.Msg) ([]byte, error) {
	if input.Signer.Account == nil {
		var err error
		account := &tx2.AccountInfo{}
		account.Number, account.Sequence, err = c.GetNumberSequence(ctx, input.Signer.PublicKey.Address())
		if err != nil {
			return nil, err
		}

		input.Signer.Account = account
	}

	signedTx, err := tx2.Sign(c.clientCtx, input, msgs...)
	if err != nil {
		return nil, err
	}

	signedBytes, err := c.clientCtx.TxConfig.TxEncoder()(signedTx)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return signedBytes, err
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

// IsInsufficientFeeError returns true if error was caused by insufficient fee provided with the transaction
func IsInsufficientFeeError(err error) bool {
	return asSDKError(err, cosmoserrors.ErrInsufficientFee) != nil
}
