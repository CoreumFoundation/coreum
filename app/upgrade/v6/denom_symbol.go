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

	sdkCtx := sdk.UnwrapSDKContext(ctx)
	switch sdkCtx.ChainID() {
	case string(constant.ChainIDMain):
		denom = constant.DenomMain
	case string(constant.ChainIDTest):
		denom = constant.DenomTest
	case string(constant.ChainIDDev):
		denom = constant.DenomDev
	default:
		return fmt.Errorf("unknown chain id: %s", sdkCtx.ChainID())
	}

	meta, found := bankKeeper.GetDenomMetaData(ctx, denom)
	if !found {
		return fmt.Errorf("denom metadata not found for %s", denom)
	}

	meta.Description = strings.ReplaceAll(meta.Description, "core", "tx")
	meta.Base = strings.ReplaceAll(meta.Base, "core", "tx")
	meta.Display = strings.ReplaceAll(meta.Display, "core", "tx")
	meta.Name = strings.ReplaceAll(meta.Name, "core", "tx")
	meta.Symbol = strings.ReplaceAll(meta.Symbol, "core", "tx")

	// Optionally adjust DenomUnits to reflect the new display name
	for i := range meta.DenomUnits {
		meta.DenomUnits[i].Denom = strings.ReplaceAll(meta.DenomUnits[i].Denom, "core", "tx")
	}

	bankKeeper.SetDenomMetaData(ctx, meta)

	return nil
}
