package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountPruner implements wasm's account pruner in a way causing smart contract instantiation to be rejected if
// account exists.
type AccountPruner struct{}

// CleanupExistingAccount informs wasm module to reject smart contract instantiation if account exists.
func (ap AccountPruner) CleanupExistingAccount(_ sdk.Context, _ authtypes.AccountI) (bool, error) {
	return false, nil
}
