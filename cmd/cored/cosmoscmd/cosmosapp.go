// COPIED FROM https://github.com/ignite/cli/tree/e6a5efdaa2210fb72e33382d442268cdd466ae2d/ignite/pkg/cosmoscmd
// UNDER APACHE2.0 LICENSE
package cosmoscmd

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CosmosApp implements the common methods for a Cosmos SDK-based application
// specific blockchain.
type CosmosApp interface {
	// The assigned name of the app.
	Name() string

	// The application types codec.
	// NOTE: This should be sealed before being returned.
	LegacyAmino() *codec.LegacyAmino

	// Application updates every begin block.
	BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock

	// Application updates every end block.
	EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock

	// Application update at chain (i.e app) initialization.
	InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain

	// Loads the app at a given height.
	LoadHeight(height int64) error

	// Exports the state of the application for a genesis file.
	ExportAppStateAndValidators(
		forZeroHeight bool, jailAllowedAddrs []string,
	) (types.ExportedApp, error)

	// All the registered module account addreses.
	ModuleAccountAddrs() map[string]bool
}
