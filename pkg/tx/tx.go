package tx

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
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
// NOTE: copied from
// https://github.com/cosmos/cosmos-sdk/blob/v0.45.2/client/tx/tx.go
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
	return clientCtx.BroadcastTx(txBytes)
}

func prepareFactory(ctx context.Context, clientCtx client.Context, txf tx.Factory) (tx.Factory, error) {
	if txf.AccountNumber() == 0 || txf.Sequence() == 0 {
		acc, err := GetAccountInfo(ctx, clientCtx, clientCtx.GetFromAddress())
		if err != nil {
			return txf, err
		}
		txf = txf.WithAccountNumber(acc.GetAccountNumber())
		txf = txf.WithSequence(acc.GetSequence())
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
