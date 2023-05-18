//go:build integrationtests

package integrationtests

import (
	ibctransfertypes "github.com/cosmos/ibc-go/v4/modules/apps/transfer/types"
)

// ConvertToIBCDenom returns the IBC denom based on the channelID and denom.
func ConvertToIBCDenom(channelID, denom string) string {
	return ibctransfertypes.ParseDenomTrace(
		ibctransfertypes.GetPrefixedDenom(ibctransfertypes.PortID, channelID, denom),
	).IBCDenom()
}
