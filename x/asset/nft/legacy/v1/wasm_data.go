package v1

import (
	"encoding/base64"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
	assetnfttypes "github.com/CoreumFoundation/coreum/x/asset/nft/types"
	"github.com/CoreumFoundation/coreum/x/nft/keeper"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
)

func isSmartContractAddress(address sdk.AccAddress) bool {
	return len(address) == wasmtypes.ContractAddrLen
}

func MigrateWasmCreatedNFTData(ctx sdk.Context, nftKeeper keeper.Keeper, assetNFTKeeper NFTKeeper) error {
	return assetNFTKeeper.IterateAllClassDefinitions(ctx, func(cd types.ClassDefinition) (bool, error) {
		issuerAddress, err := sdk.AccAddressFromBech32(cd.Issuer)
		if err != nil {
			return true, err
		}

		if !isSmartContractAddress(issuerAddress) {
			return false, nil
		}

		class, found := nftKeeper.GetClass(ctx, cd.ID)
		if !found {
			return true, errors.Errorf("class id (%s) present in definitions but not found in nft classes", cd.ID)
		}

		class.Data, err = convertAnyToDecodedAny(class.GetData())
		if err != nil {
			return true, err
		}

		nftKeeper.UpdateClass(ctx, class)

		nfts := nftKeeper.GetNFTsOfClass(ctx, cd.ID)
		for _, n := range nfts {
			n.Data, err = convertAnyToDecodedAny(n.GetData())
			if err != nil {
				return true, err
			}
			nftKeeper.Update(ctx, n)
		}

		return false, nil
	})
}

func convertAnyToDecodedAny(input *codectypes.Any) (*codectypes.Any, error) {
	if input == nil {
		return nil, nil
	}

	var oldDataByes assetnfttypes.DataBytes
	err := proto.Unmarshal(input.GetValue(), &oldDataByes)
	if err != nil {
		return nil, err
	}

	var decodedBytes []byte
	decodedBytes, err = base64.StdEncoding.DecodeString(string(oldDataByes.Data))
	if err != nil {
		return nil, err
	}

	newDataBytes := assetnfttypes.DataBytes{Data: decodedBytes}
	output, err := codectypes.NewAnyWithValue(&newDataBytes)
	if err != nil {
		return nil, err
	}

	return output, nil
}
