package integrationtests

import "github.com/CoreumFoundation/coreum/pkg/client"

const (
	GaiaAccountPrefix = "cosmos"
)

type GaiaContext struct {
	ClientContext client.Context
	ChannelID     string
}
