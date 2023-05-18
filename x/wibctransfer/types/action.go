package types

import "github.com/CoreumFoundation/coreum/x/wibc"

const (
	// ActionOut is used when IBC transfer to another chain is initialized by executing ibctransfertypes.MsgTransfer message.
	ActionOut wibc.Action = "ibcTransferOut"
	// ActionIn is used when incoming IBC transfer comes to the target chain.
	ActionIn wibc.Action = "ibcTransferIn"
)
