package v1

import (
	"encoding/base64"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"

	"github.com/CoreumFoundation/coreum/x/asset/nft/types"
)

// MigrateWasmCreatedNFTData migrates all the NFT data created by smart contracts.
// In the old binary, the encoded binary was stored in the keeper by smart contracts, and in the new
// binary it is fixed to store the original data. So we need to migrate the data
// stored by smart contracts with the old binary and store the decoded format of the old data.
func MigrateWasmCreatedNFTData(ctx sdk.Context, nftKeeper NFTKeeper, assetNFTKeeper AssetNFTKeeper, wasmKeeper WasmKeeper) error {
	return assetNFTKeeper.IterateAllClassDefinitions(ctx, func(cd types.ClassDefinition) (bool, error) {
		issuerAddress, err := sdk.AccAddressFromBech32(cd.Issuer)
		if err != nil {
			return true, err
		}

		if !wasmKeeper.HasContractInfo(ctx, issuerAddress) {
			return false, nil
		}

		class, found := nftKeeper.GetClass(ctx, cd.ID)
		if !found {
			return true, errors.Errorf("class id (%s) present in definitions but not found in nft classes", cd.ID)
		}

		if class.GetData() != nil {
			class.Data, err = convertAnyToDecodedAny(class.GetData())
			if err != nil {
				return true, err
			}

			if err := nftKeeper.UpdateClass(ctx, class); err != nil {
				return true, err
			}
		}

		nfts := nftKeeper.GetNFTsOfClass(ctx, cd.ID)
		for _, n := range nfts {
			if n.Data == nil {
				continue
			}
			n.Data, err = convertAnyToDecodedAny(n.GetData())
			if err != nil {
				return true, err
			}
			if err := nftKeeper.Update(ctx, n); err != nil {
				return true, err
			}
		}

		return false, nil
	})
}

func convertAnyToDecodedAny(input *codectypes.Any) (*codectypes.Any, error) {
	var oldDataByes types.DataBytes
	err := proto.Unmarshal(input.GetValue(), &oldDataByes)
	if err != nil {
		return nil, err
	}

	var decodedBytes []byte
	decodedBytes, err = base64.StdEncoding.DecodeString(string(oldDataByes.Data))
	if err != nil {
		return nil, err
	}

	newDataBytes := types.DataBytes{Data: decodedBytes}
	return codectypes.NewAnyWithValue(&newDataBytes)
}
