package keeper

import (
	"sort"

	sdkerrors "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"

	"github.com/CoreumFoundation/coreum/v3/x/asset/ft/types"
	wibctransfertypes "github.com/CoreumFoundation/coreum/v3/x/wibctransfer/types"
)

// BeforeSendCoins checks that a transfer request is allowed or not.
func (k Keeper) BeforeSendCoins(ctx sdk.Context, fromAddress, toAddress sdk.AccAddress, coins sdk.Coins) error {
	return k.applyFeatures(
		ctx,
		banktypes.Input{Address: fromAddress.String(), Coins: coins},
		[]banktypes.Output{{Address: toAddress.String(), Coins: coins}},
	)
}

// BeforeInputOutputCoins extends InputOutputCoins method of the bank keeper.
func (k Keeper) BeforeInputOutputCoins(ctx sdk.Context, inputs []banktypes.Input, outputs []banktypes.Output) error {
	if len(inputs) > 1 {
		return banktypes.ErrMultipleSenders
	}
	if len(inputs) == 0 {
		return banktypes.ErrNoInputs
	}
	return k.applyFeatures(ctx, inputs[0], outputs)
}

type accountOperationMap map[string]sdkmath.Int

type groupedByDenomAccountOperations map[string]accountOperationMap

func (g groupedByDenomAccountOperations) add(address string, coins sdk.Coins) {
	for _, coin := range coins {
		accountBalances, ok := g[coin.Denom]
		if !ok {
			accountBalances = make(map[string]sdkmath.Int)
		}
		oldAmount, ok := accountBalances[address]
		if !ok {
			oldAmount = sdkmath.ZeroInt()
		}

		oldAmount = oldAmount.Add(coin.Amount)
		accountBalances[address] = oldAmount
		g[coin.Denom] = accountBalances
	}
}

func (k Keeper) applyFeatures(ctx sdk.Context, input banktypes.Input, outputs []banktypes.Output) error {
	groupOutputs := make(groupedByDenomAccountOperations)
	for _, out := range outputs {
		groupOutputs.add(out.Address, out.Coins)
	}

	return k.applyRules(ctx, input, groupOutputs)
}

func (k Keeper) applyRules(ctx sdk.Context, input banktypes.Input, outputs groupedByDenomAccountOperations) error {
	sender, err := sdk.AccAddressFromBech32(input.Address)
	if err != nil {
		return sdkerrors.Wrapf(err, "invalid address %s", input.Address)
	}

	for _, coin := range input.Coins {
		def, err := k.GetDefinition(ctx, coin.Denom)
		if types.ErrInvalidDenom.Is(err) || types.ErrTokenNotFound.Is(err) {
			continue
		}

		outOps := outputs[coin.Denom]

		issuer, err := sdk.AccAddressFromBech32(def.Issuer)
		if err != nil {
			return sdkerrors.Wrapf(err, "invalid address %s", def.Issuer)
		}

		burnAmount := k.ApplyRate(ctx, def.BurnRate, issuer, sender, outOps)
		if err := k.burnIfSpendable(ctx, sender, def, burnAmount); err != nil {
			return err
		}

		commissionAmount := k.ApplyRate(ctx, def.SendCommissionRate, issuer, sender, outOps)
		commissionCoin := sdk.NewCoins(sdk.NewCoin(def.Denom, commissionAmount))
		if err := k.bankKeeper.SendCoins(ctx, sender, issuer, commissionCoin); err != nil {
			return err
		}

		if err := k.isCoinSpendable(ctx, sender, def, coin.Amount); err != nil {
			return err
		}

		if err := iterateMapDeterministic(outOps, func(account string, amount sdkmath.Int) error {
			accountAddr, err := sdk.AccAddressFromBech32(account)
			if err != nil {
				return sdkerrors.Wrapf(err, "invalid address %s", account)
			}
			return k.isCoinReceivable(ctx, accountAddr, def, amount)
		}); err != nil {
			return err
		}
		if err != nil {
			return err
		}
	}

	return nil
}

// ApplyRate calculates how the burn or commission amount should be calculated.
func (k Keeper) ApplyRate(ctx sdk.Context, rate sdk.Dec, issuer, sender sdk.AccAddress, outOps accountOperationMap) sdkmath.Int {
	// We decided that rates should not be charged on incoming IBC transfers.
	// According to our current protocol, it cannot be done because sender pays the rates, meaning that escrow address
	// would be charged leading to breaking the IBC mechanics.
	if wibctransfertypes.IsPurposeIn(ctx) {
		return sdk.ZeroInt()
	}

	// Context is marked with ACK purpose in two cases:
	// - when IBC transfer succeeded on the receiving chain (positive ACK)
	// - when IBC transfer has been rejected by the other chain (negative ACK)
	// This function is called only in the negative case, when the IBC transfer must be rolled back and funds
	// must be sent back to the sender. In this case we should not charge the rates.
	if wibctransfertypes.IsPurposeAck(ctx) {
		return sdk.ZeroInt()
	}

	// Same thing as above just in case of IBC timeout.
	if wibctransfertypes.IsPurposeTimeout(ctx) {
		return sdk.ZeroInt()
	}
	// Since burning & send commissions are not applied when sending to/from token issuer or from any smart contract,
	// we can't simply apply original burn rate or send commission rates when bank multisend contains issuer or smart contract in
	// input or issuer in outputs. To recalculate new adjusted amount we exclude amount sent to issuers.

	// Examples
	// burn_rate: 10%

	// inputs:
	// 100

	// outputs:
	// 75
	// 25 <-- issuer

	// In this case commissioned amount is: 75
	// Expected commission: 75 * 10% = 7.5
	// which is deduces from the sender account.
	if rate.IsNil() || !rate.IsPositive() {
		return sdk.ZeroInt()
	}

	if sender.String() == issuer.String() {
		return sdk.ZeroInt()
	}

	// we do not apply burn and commission rate if sender is an smart contract address.
	if len(sender) == wasmtypes.ContractAddrLen && k.wasmKeeper.HasContractInfo(ctx, sender) {
		return sdk.ZeroInt()
	}

	taxableOutputSum := sdk.NewInt(0)
	issuerStr := issuer.String()
	for account, amount := range outOps {
		if account == issuerStr {
			continue
		}
		taxableOutputSum = taxableOutputSum.Add(amount)
	}

	return rate.MulInt(taxableOutputSum).Ceil().RoundInt()
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

func iterateMapDeterministic[V any](m map[string]V, fn func(key string, value V) error) error {
	keys := sortedKeys(m)
	for _, key := range keys {
		v := m[key]
		if err := fn(key, v); err != nil {
			return err
		}
	}

	return nil
}
