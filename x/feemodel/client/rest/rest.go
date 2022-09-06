package rest

import (
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/rest"
	"github.com/gorilla/mux"
)

// RegisterHandlers registers all x/feemodel transaction and query HTTP REST handlers
// on the provided mux router.
func RegisterHandlers(clientCtx client.Context, rtr *mux.Router) {
	r := rest.WithHTTPDeprecationHeaders(rtr)
	r.HandleFunc("/feemodel/min_gas_price", QueryMinGasPriceRequestHandlerFn(clientCtx)).Methods("GET")
}
