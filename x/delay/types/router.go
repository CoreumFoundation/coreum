package types

import (
	sdkerrors "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/pkg/errors"
)

var _ Router = (*router)(nil)

// Handler executes the message.
type Handler = func(ctx sdk.Context, data proto.Message) error

// Router links message type to its handler.
type Router interface {
	RegisterHandler(data proto.Message, h Handler) error
	Handler(data proto.Message) (Handler, error)
	Has(data proto.Message) bool
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

// RegisterHandler adds a handler for the given type.
func (rtr *router) RegisterHandler(data proto.Message, h Handler) error {
	name := proto.MessageName(data)
	if _, exists := rtr.routes[name]; exists {
		return errors.Errorf("route %q has already been added", name)
	}

	rtr.routes[name] = h
	return nil
}

// Handler returns a handler for the given type.
func (rtr *router) Handler(data proto.Message) (Handler, error) {
	name := proto.MessageName(data)
	h, exists := rtr.routes[name]
	if !exists {
		return nil, sdkerrors.Wrapf(ErrInvalidConfiguration, "route %q does not exist", name)
	}

	return h, nil
}

// Has checks whether the router support the data type.
func (rtr *router) Has(data proto.Message) bool {
	name := proto.MessageName(data)
	_, exists := rtr.routes[name]
	return exists
}
