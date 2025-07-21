package v6

import (
	"context"
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/CoreumFoundation/coreum/v6/pkg/config/constant"
	wbankkeeper "github.com/CoreumFoundation/coreum/v6/x/wbank/keeper"
)

func migrateDenomSymbol(ctx context.Context, bankKeeper wbankkeeper.BaseKeeperWrapper) error {
	var denom string
	var newSymbol string

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	switch sdkCtx.ChainID() {
	case string(constant.ChainIDMain):
		denom = constant.DenomMain
		newSymbol = "TX"
	case string(constant.ChainIDTest):
		denom = constant.DenomTest
		newSymbol = "TESTTX"
	case string(constant.ChainIDDev):
		denom = constant.DenomDev
		newSymbol = "DEVTX"
	default:
		return fmt.Errorf("unknown chain id: %s", sdkCtx.ChainID())
	}

	meta, found := bankKeeper.GetDenomMetaData(ctx, denom)
	if !found {
		return fmt.Errorf("denom metadata not found for %s", denom)
	}

	meta.Display = strings.ToLower(newSymbol)
	meta.Symbol = newSymbol

	// Optionally adjust DenomUnits to reflect the new display name
	for i := range meta.DenomUnits {
		if meta.DenomUnits[i].Denom == strings.ToLower(meta.Display) || meta.DenomUnits[i].Exponent == 6 {
			meta.DenomUnits[i].Denom = strings.ToLower(newSymbol)
		}
	}

	bankKeeper.SetDenomMetaData(ctx, meta)

	return nil
}
