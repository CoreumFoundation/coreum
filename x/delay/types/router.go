package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
)

var _ Router = (*router)(nil)

// Handler executes the message.
type Handler = func(ctx sdk.Context, msg proto.Message) error

// Router links message type to its handler.
type Router interface {
	RegisterMessage(msg proto.Message, h Handler) (rtr Router)
	Handler(msg proto.Message) (h Handler)
}

type router struct {
	routes map[string]Handler
}

// NewRouter creates a new Router interface instance.
func NewRouter() Router {
	return &router{
		routes: map[string]Handler{},
	}
}

// RegisterMessage adds a handler for a message.
func (rtr *router) RegisterMessage(msg proto.Message, h Handler) Router {
	msgName := proto.MessageName(msg)
	if _, exists := rtr.routes[msgName]; exists {
		panic(fmt.Sprintf("route %q has already been added", msgName))
	}

	rtr.routes[msgName] = h
	return rtr
}

// Handler returns a handler for a given message.
func (rtr *router) Handler(msg proto.Message) Handler {
	msgName := proto.MessageName(msg)
	h, exists := rtr.routes[msgName]
	if !exists {
		panic(fmt.Sprintf("route %q does not exist", msgName))
	}

	return h
}
