package tx

import (
	"context"
	"encoding/hex"
	"fmt"
	"regexp"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/pkg/errors"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
)

const (
	requestTimeout       = 10 * time.Second
	txTimeout            = time.Minute
	txStatusPollInterval = 500 * time.Millisecond
)

var expectedSequenceRegExp = regexp.MustCompile(`account sequence mismatch, expected (\d+), got \d+`)

// BroadcastAsync sends transaction to chain, ensuring it passes CheckTx.
// Doesn't await for Tx being included in a block.
func BroadcastAsync(
	ctx context.Context,
	clientCtx client.Context,
	config SignInput,
	msgs ...sdk.Msg,
) (txHash string, err error) {
	encodedTx, err := prepareTx(clientCtx, config, msgs...)
	if err != nil {
		return "", err
	}

	txHash = fmt.Sprintf("%X", tmtypes.Tx(encodedTx).Hash())

	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	res, err := clientCtx.Client.BroadcastTxSync(requestCtx, encodedTx)
	if err != nil {
		if errors.Is(err, requestCtx.Err()) {
			return txHash, errors.WithStack(err)
		}

		errRes := checkTendermintError(err, txHash)
		if !isTxInMempool(errRes) {
			return txHash, errors.WithStack(err)
		}
	} else if res.Code != 0 {
		err := errors.Wrapf(sdkerrors.New(res.Codespace, res.Code, res.Log),
			"transaction '%s' failed", txHash)
		return txHash, err
	}

	return res.Hash.String(), nil
}

// AwaitTx awaits until a signed transaction is included in a block, returning the TxResult.
func AwaitTx(
	ctx context.Context,
	clientCtx client.Context,
	txHash string,
) (resultTx *coretypes.ResultTx, err error) {
	txHashBytes, err := hex.DecodeString(txHash)
	if err != nil {
		err = errors.Wrap(err, "tx hash is not a valid hex")
		return nil, err
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, txTimeout)
	defer cancel()

	if err = retry.Do(timeoutCtx, txStatusPollInterval, func() error {
		requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
		defer cancel()

		var err error
		resultTx, err = clientCtx.Client.Tx(requestCtx, txHashBytes, false)
		if err != nil {
			if errors.Is(err, requestCtx.Err()) {
				return retry.Retryable(errors.WithStack(err))
			}

			if errRes := checkTendermintError(err, txHash); errRes != nil {
				if isTxInMempool(errRes) {
					return retry.Retryable(errors.WithStack(err))
				}
				return errors.WithStack(err)
			}

			return retry.Retryable(errors.WithStack(err))
		}

		if resultTx.TxResult.Code != 0 {
			res := resultTx.TxResult
			return errors.Wrapf(sdkerrors.New(res.Codespace, res.Code, res.Log), "transaction '%s' failed", txHash)
		}

		if resultTx.Height == 0 {
			return retry.Retryable(errors.Errorf("transaction '%s' hasn't been included in a block yet", txHash))
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return resultTx, nil
}

// prepareTx encodes messages in a new transaction then signs and encodes it
func prepareTx(
	clientCtx client.Context,
	config SignInput,
	msgs ...sdk.Msg,
) ([]byte, error) {
	fromAddress := sdk.AccAddress(config.PrivateKey.PubKey().Address())
	if config.AccountInfo.Number == 0 {
		info, err := clientCtx.AccountRetriever.GetAccount(clientCtx, fromAddress)
		if err != nil {
			return nil, err
		}
		acc := AccountInfo{
			Number:   info.GetAccountNumber(),
			Sequence: info.GetSequence(),
		}

		config.AccountInfo = acc
	}

	signedTx, err := Sign(clientCtx, config, msgs...)
	if err != nil {
		return nil, err
	}

	signedBytes, err := clientCtx.TxConfig.TxEncoder()(signedTx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return signedBytes, err
}

func isTxInMempool(errRes *sdk.TxResponse) bool {
	if errRes == nil {
		return false
	}
	return isSDKErrorResult(errRes.Codespace, errRes.Code, sdkerrors.ErrTxInMempoolCache)
}
