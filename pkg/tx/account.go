package tx

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/pkg/errors"
)

// GetAccountInfo returns account number and account sequence for provided address
func GetAccountInfo(
	ctx context.Context,
	clientCtx client.Context,
	address sdk.AccAddress,
) (AccountInfo, error) {
	requestCtx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	req := &authtypes.QueryAccountRequest{
		Address: address.String(),
	}

	authQueryClient := authtypes.NewQueryClient(clientCtx)
	res, err := authQueryClient.Account(requestCtx, req)
	if err != nil {
		return AccountInfo{}, errors.WithStack(err)
	}

	var acc authtypes.AccountI
	if err := clientCtx.InterfaceRegistry.UnpackAny(res.Account, &acc); err != nil {
		return AccountInfo{}, errors.WithStack(err)
	}

	return AccountInfo{
		Number:   acc.GetAccountNumber(),
		Sequence: acc.GetSequence(),
	}, nil
}
