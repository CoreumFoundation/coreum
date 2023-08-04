package types

import (
	sdkerrors "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	"github.com/pkg/errors"
)

var _ Router = (*router)(nil)

// Handler executes the message.
type Handler = func(ctx sdk.Context, data proto.Message) error

// Router links message type to its handler.
type Router interface {
	RegisterHandler(data codec.ProtoMarshaler, h Handler) error
	Handler(data codec.ProtoMarshaler) (Handler, error)
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
func (rtr *router) RegisterHandler(data codec.ProtoMarshaler, h Handler) error {
	name := proto.MessageName(data)
	if _, exists := rtr.routes[name]; exists {
		return errors.Errorf("route %q has already been added", name)
	}

	rtr.routes[name] = h
	return nil
}

// Handler returns a handler for the given type.
func (rtr *router) Handler(data codec.ProtoMarshaler) (Handler, error) {
	name := proto.MessageName(data)
	h, exists := rtr.routes[name]
	if !exists {
		return nil, sdkerrors.Wrapf(ErrInvalidConfiguration, "route %q does not exist", name)
	}

	return h, nil
}
