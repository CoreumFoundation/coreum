package deterministicgas_test

import (
	"fmt"
	"reflect"
	"testing"
	_ "unsafe"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"

	"github.com/CoreumFoundation/coreum/pkg/config"
	"github.com/CoreumFoundation/coreum/testutil/simapp"
)

//go:linkname revProtoTypes github.com/gogo/protobuf/proto.revProtoTypes
var revProtoTypes map[reflect.Type]string

func TestDeterministicGasRequirements(t *testing.T) {
	simapp.New()

	fmt.Printf("total types num: %v\n\n", len(revProtoTypes))

	// the idea here is to iterate through all types convertable to sdk.Msg and verify
	// that they are defined as deterministic or non-deterministic in DeterministicGasRequirements.
	// Some internal types might be skipped e.g. tendermint.*
	for tt, name := range revProtoTypes {
		reflectPtr := reflect.New(tt)
		sdkMsg, ok := reflectPtr.Elem().Interface().(sdk.Msg)
		if !ok {
			continue
		}

		fmt.Printf("%T => %v\n", sdkMsg, name)
		_, ok = config.DefaultDeterministicGasRequirements().GasRequiredByMessage(sdkMsg)
		assert.True(t, ok)
	}
}
