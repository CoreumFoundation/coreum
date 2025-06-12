//go:build integrationtests

package ibc

import (
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
)

// ConvertToIBCDenom returns the IBC denom based on the channelID and denom.
func ConvertToIBCDenom(channelID, denom string) string {
	//nolint:staticcheck // TODO
	return ibctransfertypes.ParseDenomTrace(
		//nolint:staticcheck // TODO
		ibctransfertypes.GetPrefixedDenom(ibctransfertypes.PortID, channelID, denom),
	).IBCDenom()
}
