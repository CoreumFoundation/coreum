package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type smartContractRecipientKey struct{}
type smartContractSenderKey struct{}

// WithSmartContractSender sets address of sending smart contract.
func WithSmartContractSender(ctx sdk.Context, sender string) sdk.Context {
	return add(ctx, sender, smartContractSenderKey{})
}

// WithSmartContractRecipient sets address of receiving smart contract.
func WithSmartContractRecipient(ctx sdk.Context, recipient string) sdk.Context {
	return add(ctx, recipient, smartContractRecipientKey{})
}

// IsSendingSmartContract returns true if address is the smart contract sending funds.
func IsSendingSmartContract(ctx sdk.Context, addr string) bool {
	return has(ctx, addr, smartContractSenderKey{})
}

// IsReceivingSmartContract returns true if address is the smart contract receiving funds.
func IsReceivingSmartContract(ctx sdk.Context, addr string) bool {
	return has(ctx, addr, smartContractRecipientKey{})
}

func add(ctx sdk.Context, addr string, key struct{}) sdk.Context {
	set, ok := ctx.Value(key).(map[string]struct{})
	if !ok || set == nil {
		set = map[string]struct{}{}
	}
	set[addr] = struct{}{}
	return ctx.WithValue(key, set)
}

func has(ctx sdk.Context, addr string, key struct{}) bool {
	set, ok := ctx.Value(key).(map[string]struct{})
	if !ok {
		return false
	}

	_, exists := set[addr]
	return exists
}
