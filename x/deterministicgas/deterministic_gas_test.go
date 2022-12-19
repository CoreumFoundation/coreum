package deterministicgas_test

import (
	"fmt"
	"reflect"
	"testing"
	_ "unsafe"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distributiontypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/golang/protobuf/proto"

	"github.com/CoreumFoundation/coreum/testutil/simapp"
	assettypes "github.com/CoreumFoundation/coreum/x/asset/types"
)

func TestProtoMessageName(t *testing.T) {
	msg := &distributiontypes.MsgFundCommunityPool{}
	name := proto.MessageName(msg)
	fmt.Println(name) // "cosmos.distribution.v1beta1.MsgFundCommunityPool"

	msg2 := &assettypes.MsgBurnFungibleToken{}
	name2 := proto.MessageName(msg2)
	fmt.Println(name2) // "coreum.asset.v1.MsgBurnFungibleToken"

	msg3 := &banktypes.MsgSend{}
	name3 := proto.MessageName(msg3)
	fmt.Println(name3) // cosmos.bank.v1beta1.MsgSend

	msg4 := &wasmtypes.MsgExecuteContract{}
	name4 := proto.MessageName(msg4)
	fmt.Println(name4) // cosmwasm.wasm.v1.MsgExecuteContract
}

//go:linkname revProtoTypes github.com/gogo/protobuf/proto.revProtoTypes
var revProtoTypes map[reflect.Type]string

func TestRevProtoTypes(t *testing.T) {
	simapp.New()

	fmt.Printf("total types num: %v\n\n", len(revProtoTypes))
	for tt, name := range revProtoTypes {
		reflectPtr := reflect.New(tt)
		sdkMsg, ok := reflectPtr.Elem().Interface().(sdk.Msg)
		if !ok {
			continue
		}

		fmt.Printf("%T => %v\n", sdkMsg, name)
	}
}
