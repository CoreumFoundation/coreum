package integrationtests

import "github.com/CoreumFoundation/coreum/pkg/client"

// gaia constants.
const (
	GaiaAccountPrefix = "cosmos"
)

// GaiaContext contains all the information related to the gaia chain.
type GaiaContext struct {
	ClientContext client.Context
}
