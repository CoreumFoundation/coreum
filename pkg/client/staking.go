package client

import (
	"context"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// GetValidators returns validators list
func (c Client) GetValidators(ctx context.Context) ([]stakingtypes.Validator, error) {
	resp, err := c.stakingQueryClient.Validators(ctx, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.BondStatusBonded,
	})
	must.OK(err)

	return resp.Validators, nil
}

// TxSubmitDelegationInput holds input data for PrepareTxSubmitDelegation
type TxSubmitDelegationInput struct {
	Delegator types.Wallet
	Validator sdk.ValAddress
	Amount    types.Coin

	Base tx.BaseInput
}

// PrepareTxSubmitDelegation creates a transaction to submit a delegation
func (c Client) PrepareTxSubmitDelegation(ctx context.Context, input TxSubmitDelegationInput) ([]byte, error) {
	delegatorAddress, err := sdk.AccAddressFromBech32(input.Delegator.Key.Address())
	if err != nil {
		return nil, err
	}

	if err = input.Amount.Validate(); err != nil {
		return nil, errors.Wrap(err, "amount to delegate is invalid")
	}

	msg := stakingtypes.NewMsgDelegate(delegatorAddress, input.Validator, sdk.Coin{
		Denom:  input.Amount.Denom,
		Amount: sdk.NewIntFromBigInt(input.Amount.Amount),
	})

	signedTx, err := c.Sign(ctx, input.Base, msg)
	if err != nil {
		return nil, err
	}

	return c.Encode(signedTx), nil
}
