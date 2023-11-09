package types

import (
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankCoinTransferrer transfers coins to the smart contract.
type BankCoinTransferrer struct {
	parentTransferrer wasmkeeper.CoinTransferrer
}

// NewBankCoinTransferrer returns new transferrer.
func NewBankCoinTransferrer(bankKeeper wasmtypes.BankKeeper) BankCoinTransferrer {
	return BankCoinTransferrer{
		parentTransferrer: wasmkeeper.NewBankCoinTransferrer(bankKeeper),
	}
}

// TransferCoins transfers coins to the smart contract.
func (c BankCoinTransferrer) TransferCoins(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amount sdk.Coins) error {
	return c.parentTransferrer.TransferCoins(WithSmartContractRecipient(ctx, toAddr.String()), fromAddr, toAddr, amount)
}
