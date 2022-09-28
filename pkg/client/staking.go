package client

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/pkg/tx"
	"github.com/CoreumFoundation/coreum/pkg/types"
)

// GetBondedTokens returns bonded tokens amount
func (c Client) GetBondedTokens(ctx context.Context) (sdk.Int, error) {
	resp, err := c.StakingQueryClient().Pool(ctx, &stakingtypes.QueryPoolRequest{})
	if err != nil {
		return sdk.NewInt(0), err
	}

	return resp.Pool.BondedTokens, nil
}

// GetStakingParams returns staking params
func (c Client) GetStakingParams(ctx context.Context) (*stakingtypes.Params, error) {
	resp, err := c.StakingQueryClient().Params(ctx, &stakingtypes.QueryParamsRequest{})
	if err != nil {
		return nil, err
	}

	return &resp.Params, nil
}

// GetValidators returns validators list
func (c Client) GetValidators(ctx context.Context) ([]stakingtypes.Validator, error) {
	resp, err := c.StakingQueryClient().Validators(ctx, &stakingtypes.QueryValidatorsRequest{
		Status: stakingtypes.BondStatusBonded,
	})
	if err != nil {
		return nil, err
	}

	return resp.Validators, nil
}

// GetValidator returns validator by the given address
func (c Client) GetValidator(ctx context.Context, addr sdk.ValAddress) (*stakingtypes.Validator, error) {
	resp, err := c.StakingQueryClient().Validator(ctx, &stakingtypes.QueryValidatorRequest{
		ValidatorAddr: addr.String(),
	})
	if err != nil {
		return nil, err
	}

	return &resp.Validator, nil
}

// TxSubmitDelegationInput holds input data for PrepareTxSubmitDelegation
type TxSubmitDelegationInput struct {
	Delegator types.Wallet
	Validator sdk.ValAddress
	Amount    sdk.Coin

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

	msg := stakingtypes.NewMsgDelegate(delegatorAddress, input.Validator, input.Amount)

	signedTx, err := c.Sign(ctx, input.Base, msg)
	if err != nil {
		return nil, err
	}

	return c.Encode(signedTx), nil
}
