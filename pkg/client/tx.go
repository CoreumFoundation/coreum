package client

// This file contains helper functions used to prepare and broadcast transactions.
// Blocking broadcast mode was reimplemented to use polling instead of subscription to eliminate the case when
// transaction execution is missed due to broken websocket connection.
// For other broadcast modes we just call original Cosmos implementation.
// For more details check BroadcastRawTx & broadcastTxBlock.

import (
	"context"
	"fmt"
	"strings"

	sdkerrors "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/mempool"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/client/tx"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmoserrors "github.com/cosmos/cosmos-sdk/types/errors"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum-tools/pkg/retry"
	feemodeltypes "github.com/CoreumFoundation/coreum/v3/x/feemodel/types"
)

// Factory is a re-export of the cosmos sdk tx.Factory type, to make usage of this package more convenient.
// It will help users by removing the need to import tx package from cosmos sdk and help avoid package name collision.
type Factory = tx.Factory

// Sign signs a given tx with a named key. The bytes signed over are canonical.
// The resulting signature will be added to the transaction builder overwriting the previous
// ones if overwrite=true (otherwise, the signature will be appended).
// Signing a transaction with mutltiple signers in the DIRECT mode is not supprted and will
// return an error.
// An error is returned upon failure.
// https://github.com/cosmos/cosmos-sdk/blob/v0.45.2/client/tx/tx.go
var Sign = tx.Sign

// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will return an error upon failure.
// NOTE: copied from the link below and made some changes.
// the main idea is to add context.Context to the signature and use it
// https://github.com/cosmos/cosmos-sdk/blob/v0.45.2/client/tx/tx.go
func BroadcastTx(ctx context.Context, clientCtx Context, txf Factory, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	txf, err := prepareFactory(ctx, clientCtx, txf)
	if err != nil {
		return nil, err
	}

	if txf.SimulateAndExecute() {
		gasPrice, err := GetGasPrice(ctx, clientCtx)
		if err != nil {
			return nil, err
		}
		gasPrice.Amount = gasPrice.Amount.Mul(clientCtx.GasPriceAdjustment())
		txf = txf.WithGasPrices(gasPrice.String())

		_, adjusted, err := CalculateGas(ctx, clientCtx, txf, msgs...)
		if err != nil {
			return nil, err
		}

		txf = txf.WithGas(adjusted)
	}

	unsignedTx, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	unsignedTx.SetFeeGranter(clientCtx.FeeGranterAddress())

	// in case the name is not provided by that address, take the name by the address
	fromName := clientCtx.FromName()
	if fromName == "" && len(clientCtx.FromAddress()) > 0 {
		key, err := clientCtx.Keyring().KeyByAddress(clientCtx.FromAddress())
		if err != nil {
			return nil, errors.Errorf("failed to get key by the address %q from the keyring", clientCtx.FromAddress().String())
		}
		fromName = key.Name
	}

	err = tx.Sign(txf, fromName, unsignedTx, true)
	if err != nil {
		return nil, err
	}

	txBytes, err := clientCtx.TxConfig().TxEncoder()(unsignedTx.GetTx())
	if err != nil {
		return nil, err
	}

	return BroadcastRawTx(ctx, clientCtx, txBytes)
}

// CalculateGas simulates the execution of a transaction and returns the
// simulation response obtained by the query and the adjusted gas amount.
//
//	FIXME(v47-multisig-calculate-gas-test) add test to calculate
func CalculateGas(ctx context.Context, clientCtx Context, txf Factory, msgs ...sdk.Msg) (*sdktx.SimulateResponse, uint64, error) {
	txf, err := prepareFactory(ctx, clientCtx, txf)
	if err != nil {
		return nil, 0, err
	}
	if txf.GasAdjustment() == 0 {
		txf = txf.WithGasAdjustment(clientCtx.GasAdjustment())
	}

	keyInfo, err := txf.Keybase().KeyByAddress(clientCtx.FromAddress())
	if err != nil {
		return nil, 0, errors.WithStack(err)
	}

	msgsAny := make([]*codectypes.Any, 0, len(msgs))
	for _, msg := range msgs {
		msgAny, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return nil, 0, errors.WithStack(err)
		}
		msgsAny = append(msgsAny, msgAny)
	}

	txBodyBytes, err := proto.Marshal(&sdktx.TxBody{
		Messages: msgsAny,
		Memo:     txf.Memo(),
	})
	if err != nil {
		return nil, 0, errors.WithStack(err)
	}

	pubKey, err := keyInfo.GetPubKey()
	if err != nil {
		return nil, 0, errors.WithStack(err)
	}
	var signatureData signing.SignatureData
	if multisigPubKey, ok := pubKey.(*multisig.LegacyAminoPubKey); ok {
		multiSignatureData := make([]signing.SignatureData, 0, multisigPubKey.Threshold)
		for i := uint32(0); i < multisigPubKey.Threshold; i++ {
			multiSignatureData = append(multiSignatureData, &signing.SingleSignatureData{
				SignMode: txf.SignMode(),
			})
		}
		signatureData = &signing.MultiSignatureData{
			Signatures: multiSignatureData,
		}
	} else {
		signatureData = &signing.SingleSignatureData{
			SignMode: txf.SignMode(),
		}
	}

	modeInfo, signature := authtx.SignatureDataToModeInfoAndSig(signatureData)

	simAuthInfoBytes, err := proto.Marshal(&sdktx.AuthInfo{
		SignerInfos: []*sdktx.SignerInfo{
			{
				PublicKey: keyInfo.PubKey,
				ModeInfo:  modeInfo,
				Sequence:  txf.Sequence(),
			},
		},
		Fee: &sdktx.Fee{},
	})
	if err != nil {
		return nil, 0, errors.WithStack(err)
	}

	txBytes, err := proto.Marshal(&sdktx.TxRaw{
		BodyBytes:     txBodyBytes,
		AuthInfoBytes: simAuthInfoBytes,
		Signatures: [][]byte{
			signature,
		},
	})
	if err != nil {
		return nil, 0, errors.WithStack(err)
	}

	txSvcClient := sdktx.NewServiceClient(clientCtx)
	simRes, err := txSvcClient.Simulate(ctx, &sdktx.SimulateRequest{
		TxBytes: txBytes,
	})
	if err != nil {
		return nil, 0, errors.Wrap(err, "transaction estimation failed")
	}

	return simRes, uint64(txf.GasAdjustment() * float64(simRes.GasInfo.GasUsed)), nil
}

// BroadcastRawTx broadcast the txBytes using the clientCtx and set BroadcastMode.
func BroadcastRawTx(ctx context.Context, clientCtx Context, txBytes []byte) (*sdk.TxResponse, error) {
	var (
		txRes *sdk.TxResponse
		err   error
	)
	switch clientCtx.BroadcastMode() {
	case flags.BroadcastSync:
		txRes, err = broadcastTxSync(ctx, clientCtx, txBytes)

	case flags.BroadcastAsync:
		txRes, err = broadcastTxAsync(ctx, clientCtx, txBytes)

	default:
		return nil, errors.Errorf("unsupported broadcast mode %s; supported types: sync, async, block", clientCtx.BroadcastMode())
	}
	if err != nil {
		return nil, err
	}

	if clientCtx.GetAwaitTx() {
		txRes, err = AwaitTx(ctx, clientCtx, txRes.TxHash)
		if err != nil {
			return nil, err
		}
	}

	return txRes, nil
}

// GetAccountInfo returns account number and account sequence for provided address.
func GetAccountInfo(
	ctx context.Context,
	clientCtx Context,
	address sdk.AccAddress,
) (authtypes.AccountI, error) {
	req := &authtypes.QueryAccountRequest{
		Address: address.String(),
	}
	authQueryClient := authtypes.NewQueryClient(clientCtx)
	res, err := authQueryClient.Account(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	var acc authtypes.AccountI
	if err := clientCtx.InterfaceRegistry().UnpackAny(res.Account, &acc); err != nil {
		return nil, errors.WithStack(err)
	}

	return acc, nil
}

// AwaitTx waits until a signed transaction is included in a block, returning the result.
func AwaitTx(
	ctx context.Context,
	clientCtx Context,
	txHash string,
) (txResponse *sdk.TxResponse, err error) {
	txSvcClient := sdktx.NewServiceClient(clientCtx)
	timeoutCtx, cancel := context.WithTimeout(ctx, clientCtx.config.TimeoutConfig.TxTimeout)
	defer cancel()

	if err = retry.Do(timeoutCtx, clientCtx.config.TimeoutConfig.TxStatusPollInterval, func() error {
		requestCtx, cancel := context.WithTimeout(ctx, clientCtx.config.TimeoutConfig.RequestTimeout)
		defer cancel()

		res, err := txSvcClient.GetTx(requestCtx, &sdktx.GetTxRequest{
			Hash: txHash,
		})
		if err != nil {
			return retry.Retryable(errors.WithStack(err))
		}

		txResponse = res.TxResponse
		if txResponse.Code != 0 {
			return errors.Wrapf(sdkerrors.ABCIError(txResponse.Codespace, txResponse.Code, txResponse.Logs.String()),
				"transaction '%s' failed, raw log:%s", txResponse.TxHash, txResponse.RawLog)
		}

		if txResponse.Height == 0 {
			return retry.Retryable(errors.Errorf("transaction '%s' hasn't been included in a block yet", txHash))
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return txResponse, nil
}

// AwaitNextBlocks waits for next blocks.
func AwaitNextBlocks(
	ctx context.Context,
	clientCtx Context,
	nextBlocks int64,
) error {
	tmQueryClient := tmservice.NewServiceClient(clientCtx)
	timeoutCtx, cancel := context.WithTimeout(ctx, clientCtx.config.TimeoutConfig.TxNextBlocksTimeout)
	defer cancel()

	heightToStart := int64(0)
	return retry.Do(timeoutCtx, clientCtx.config.TimeoutConfig.TxNextBlocksPollInterval, func() error {
		requestCtx, cancel := context.WithTimeout(ctx, clientCtx.config.TimeoutConfig.RequestTimeout)
		defer cancel()

		res, err := tmQueryClient.GetLatestBlock(requestCtx, &tmservice.GetLatestBlockRequest{})
		if err != nil {
			return retry.Retryable(errors.WithStack(err))
		}

		var currentHeight int64
		if res.SdkBlock != nil {
			currentHeight = res.SdkBlock.Header.Height
		} else {
			// TODO: Remove this in v4 version of cored. Now it is needed because we might still use it in integration tests together with v2 cored binary.
			currentHeight = res.Block.Header.Height //nolint:staticcheck // Yes, we know that this is deprecated
		}

		if heightToStart == 0 {
			heightToStart = currentHeight
		}

		targetHeight := heightToStart + nextBlocks
		if currentHeight < targetHeight {
			return retry.Retryable(errors.Errorf("target block: %d hasn't been reached yet, current: %d", targetHeight, currentHeight))
		}

		return nil
	})
}

// GetGasPrice returns the current gas price of the chain.
func GetGasPrice(
	ctx context.Context,
	clientCtx Context,
) (sdk.DecCoin, error) {
	feeQueryClient := feemodeltypes.NewQueryClient(clientCtx)
	res, err := feeQueryClient.MinGasPrice(ctx, &feemodeltypes.QueryMinGasPriceRequest{})
	if err != nil {
		return sdk.DecCoin{}, errors.WithStack(err)
	}

	return res.GetMinGasPrice(), nil
}

func broadcastTxAsync(ctx context.Context, clientCtx Context, txBytes []byte) (*sdk.TxResponse, error) {
	requestCtx, cancel := context.WithTimeout(ctx, clientCtx.config.TimeoutConfig.RequestTimeout)
	defer cancel()

	// rpc client
	if clientCtx.RPCClient() != nil {
		res, err := clientCtx.RPCClient().BroadcastTxAsync(requestCtx, txBytes)
		if err != nil {
			return nil, err
		}
		return sdk.NewResponseFormatBroadcastTx(res), nil
	}
	// grpc client
	txSvcClient := sdktx.NewServiceClient(clientCtx)
	res, err := txSvcClient.BroadcastTx(requestCtx, &sdktx.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    sdktx.BroadcastMode_BROADCAST_MODE_ASYNC,
	})
	if err != nil {
		return nil, err
	}

	return res.TxResponse, nil
}

func broadcastTxSync(ctx context.Context, clientCtx Context, txBytes []byte) (*sdk.TxResponse, error) {
	requestCtx, cancel := context.WithTimeout(ctx, clientCtx.config.TimeoutConfig.RequestTimeout)
	defer cancel()

	// rpc client
	txHash := fmt.Sprintf("%X", tmtypes.Tx(txBytes).Hash())
	if clientCtx.RPCClient() != nil {
		res, err := clientCtx.RPCClient().BroadcastTxSync(requestCtx, txBytes)
		if err != nil {
			if err := processTxCommitError(requestCtx, err); err != nil {
				return nil, err
			}
		} else if res.Code != 0 {
			return nil, errors.Wrapf(sdkerrors.ABCIError(res.Codespace, res.Code, res.Log),
				"transaction '%s' failed", txHash)
		}

		return sdk.NewResponseFormatBroadcastTx(res), nil
	}

	// grpc client
	txSvcClient := sdktx.NewServiceClient(clientCtx)
	res, err := txSvcClient.BroadcastTx(requestCtx, &sdktx.BroadcastTxRequest{
		TxBytes: txBytes,
		Mode:    sdktx.BroadcastMode_BROADCAST_MODE_SYNC,
	})
	if err != nil {
		if err := processTxCommitError(requestCtx, err); err != nil {
			return nil, err
		}
	} else if res.TxResponse.Code != 0 {
		return nil, errors.Wrapf(sdkerrors.ABCIError(res.TxResponse.Codespace, res.TxResponse.Code, res.TxResponse.Logs.String()),
			"transaction '%s' failed, raw log:%s", res.TxResponse.TxHash, res.TxResponse.RawLog)
	}

	return res.TxResponse, nil
}

func processTxCommitError(ctx context.Context, err error) error {
	if errors.Is(err, ctx.Err()) {
		return errors.WithStack(err)
	}

	if err := convertTendermintError(err); !cosmoserrors.ErrTxInMempoolCache.Is(err) {
		return errors.WithStack(err)
	}

	return nil
}

func prepareFactory(ctx context.Context, clientCtx Context, txf tx.Factory) (tx.Factory, error) {
	if txf.AccountNumber() == 0 && txf.Sequence() == 0 {
		acc, err := GetAccountInfo(ctx, clientCtx, clientCtx.FromAddress())
		if err != nil {
			return txf, err
		}
		txf = txf.
			WithAccountNumber(acc.GetAccountNumber()).
			WithSequence(acc.GetSequence())
	}

	return txf, nil
}

// the idea behind this function is to map it similarly to how cosmos sdk does it in the link below
// so the users can match against cosmos sdk error types.
// https://github.com/cosmos/cosmos-sdk/blob/v0.45.2/client/broadcast.go#L49
func convertTendermintError(err error) error {
	if err == nil {
		return nil
	}
	errStr := strings.ToLower(err.Error())

	switch {
	case strings.Contains(errStr, strings.ToLower(mempool.ErrTxInCache.Error())):
		return cosmoserrors.ErrTxInMempoolCache.Wrap(err.Error())
	case strings.Contains(errStr, cosmoserrors.ErrMempoolIsFull.Error()):
		return cosmoserrors.ErrMempoolIsFull.Wrap(err.Error())
	case strings.Contains(errStr, cosmoserrors.ErrTxTooLarge.Error()):
		return cosmoserrors.ErrTxTooLarge.Wrap(err.Error())
	default:
		return err
	}
}
