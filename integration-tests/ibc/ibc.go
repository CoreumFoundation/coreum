//go:build integrationtests

package ibc

import (
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
)

// ConvertToIBCDenom returns the IBC denom based on the channelID and denom.
func ConvertToIBCDenom(channelID, denom string) string {
	//nolint:staticcheck // TODO: fix after upgrading to cosmos v0.53
	return ibctransfertypes.ParseDenomTrace(
		//nolint:staticcheck // TODO: fix after upgrading to cosmos v0.53
		ibctransfertypes.GetPrefixedDenom(ibctransfertypes.PortID, channelID, denom),
	).IBCDenom()
}
