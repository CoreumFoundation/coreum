package testing

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
)

// FindTypedEvent finds the event in the list of events and returns the decoded event.
func FindTypedEvent(t T, event proto.Message, events []abci.Event) interface{} {
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
