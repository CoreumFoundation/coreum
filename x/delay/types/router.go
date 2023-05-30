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
	RegisterHandler(msg proto.Message, h Handler) (rtr Router)
	Handler(data proto.Message) (h Handler)
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
func (rtr *router) RegisterHandler(data proto.Message, h Handler) Router {
	name := proto.MessageName(data)
	if _, exists := rtr.routes[name]; exists {
		panic(fmt.Sprintf("route %q has already been added", name))
	}

	rtr.routes[name] = h
	return rtr
}

// Handler returns a handler for a given message.
func (rtr *router) Handler(data proto.Message) Handler {
	name := proto.MessageName(data)
	h, exists := rtr.routes[name]
	if !exists {
		panic(fmt.Sprintf("route %q does not exist", name))
	}

	return h
}
