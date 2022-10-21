package testing

import (
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	tmtypes "github.com/tendermint/tendermint/abci/types"
)

// FindTypedEvent finds the event in the list of events and returns the decoded event.
func FindTypedEvent(t T, event proto.Message, events []tmtypes.Event) interface{} {
	eventName := proto.MessageName(event)
	for i := range events {
		if events[i].Type != eventName {
			continue
		}

		msg, err := sdk.ParseTypedEvent(events[i])
		require.NoError(t, err)

		return msg
	}

	require.Failf(t, "%s event, not found in the events", eventName)
	return nil
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
