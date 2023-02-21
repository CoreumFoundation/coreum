package event

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	tmtypes "github.com/tendermint/tendermint/abci/types"
)

// FindTypedEvents finds events in the list of events, and marshals them to the event type.
func FindTypedEvents[T proto.Message](events []tmtypes.Event) ([]T, error) {
	var res []T

	event := *new(T)
	eventName := proto.MessageName(event)
	for _, e := range events {
		if e.Type != eventName {
			continue
		}

		msg, err := sdk.ParseTypedEvent(e)
		if err != nil {
			return nil, err
		}

		typedMsg, ok := msg.(T)
		if !ok {
			return nil, errors.Errorf("can't cast found event to %T", event)
		}

		res = append(res, typedMsg)
	}
	if len(res) == 0 {
		return nil, errors.Errorf("can't find event %T in events", event)
	}
	return res, nil
}

// FindUint64EventAttribute finds the first event attribute by type and attribute name and convert it to the uint64 type.
func FindUint64EventAttribute(events []tmtypes.Event, etype, attribute string) (uint64, error) {
	strAttr, err := FindStringEventAttribute(events, etype, attribute)
	if err != nil {
		return 0, errors.New("can't find the codeID in the tx events")
	}
	uintAttr, err := strconv.ParseUint(strAttr, 10, 64)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to parse types %q event attribute %q event attribute as uint64", etype, attribute)
	}

	return uintAttr, nil
}

// FindStringEventAttribute finds the first string event attribute by type and attribute name.
func FindStringEventAttribute(events []tmtypes.Event, etype, attribute string) (string, error) {
	for _, ev := range sdk.StringifyEvents(events) {
		if ev.Type == etype {
			if value, found := findAttribute(ev, attribute); found {
				return value, nil
			}
		}
	}
	return "", errors.Errorf("can't find the types %q event attribute %q of ", etype, attribute)
}

func findAttribute(ev sdk.StringEvent, attr string) (string, bool) {
	for _, attrItem := range ev.Attributes {
		if attrItem.Key == attr {
			return attrItem.Value, true
		}
	}

	return "", false
}
