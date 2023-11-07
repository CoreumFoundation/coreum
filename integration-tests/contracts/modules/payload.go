package modules

import (
	"encoding/json"

	"github.com/CoreumFoundation/coreum-tools/pkg/must"
)

// EmptyPayload represents empty payload.
var EmptyPayload = must.Bytes(json.Marshal(struct{}{}))
