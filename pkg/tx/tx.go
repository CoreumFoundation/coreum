package tx

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/pkg/errors"
)

// Factory is a re-export of the cosmos sdk tx.Factory type, to make usage of this package more convenient.
// It will help users by removing the need to import tx package from cosmos sdk and help avoid package name collision.
type Factory = tx.Factory

// BroadcastTx attempts to generate, sign and broadcast a transaction with the
// given set of messages. It will also simulate gas requirements if necessary.
// It will return an error upon failure.
// NOTE: copied from the link below and made some changes
// https://github.com/cosmos/cosmos-sdk/blob/v0.45.2/client/tx/tx.go
// TODO: add test to check if client respects ctx.
func BroadcastTx(ctx context.Context, clientCtx client.Context, txf Factory, msgs ...sdk.Msg) (*sdk.TxResponse, error) {
	txf, err := prepareFactory(ctx, clientCtx, txf)
	if err != nil {
		return nil, err
	}

	unsignedTx, err := txf.BuildUnsignedTx(msgs...)
	if err != nil {
		return nil, err
	}

	unsignedTx.SetFeeGranter(clientCtx.GetFeeGranterAddress())
	err = tx.Sign(txf, clientCtx.GetFromName(), unsignedTx, true)
	if err != nil {
		return nil, err
	}

	txBytes, err := clientCtx.TxConfig.TxEncoder()(unsignedTx.GetTx())
	if err != nil {
		return nil, err
	}

	// broadcast to a Tendermint node
	switch clientCtx.BroadcastMode {
	case flags.BroadcastSync:
		res, err := clientCtx.Client.BroadcastTxSync(ctx, txBytes)
		if err != nil {
			return nil, err
		}
		return sdk.NewResponseFormatBroadcastTx(res), nil

	case flags.BroadcastAsync:
		res, err := clientCtx.Client.BroadcastTxAsync(ctx, txBytes)
		if err != nil {
			return nil, err
		}
		return sdk.NewResponseFormatBroadcastTx(res), nil

	case flags.BroadcastBlock:
		res, err := clientCtx.Client.BroadcastTxCommit(ctx, txBytes)
		if err != nil {
			return nil, err
		}
		return sdk.NewResponseFormatBroadcastTxCommit(res), nil

	default:
		return nil, errors.Errorf("unsupported broadcast mode %s; supported types: sync, async, block", clientCtx.BroadcastMode)
	}

}

func prepareFactory(ctx context.Context, clientCtx client.Context, txf tx.Factory) (tx.Factory, error) {
	if txf.AccountNumber() == 0 || txf.Sequence() == 0 {
		acc, err := GetAccountInfo(ctx, clientCtx, clientCtx.GetFromAddress())
		if err != nil {
			return txf, err
		}
		txf = txf.
			WithAccountNumber(acc.GetAccountNumber()).
			WithSequence(acc.GetSequence())
	}

	return txf, nil
}

// GetAccountInfo returns account number and account sequence for provided address
func GetAccountInfo(
	ctx context.Context,
	clientCtx client.Context,
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
	if err := clientCtx.InterfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
		return nil, errors.WithStack(err)
	}

	return acc, nil
}
