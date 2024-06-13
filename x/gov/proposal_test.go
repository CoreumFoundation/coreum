package gov_test

import (
	"reflect"
	"testing"
	_ "unsafe"

	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/assert"

	"github.com/CoreumFoundation/coreum/v4/testutil/simapp"
)

// To access private variable from github.com/cosmos/gogoproto we link it to local variable.
// This is needed to iterate through all registered protobuf types.
//
//go:linkname revProtoTypes github.com/cosmos/gogoproto/proto.revProtoTypes
var revProtoTypes map[reflect.Type]string

// TODO(v4): drop together with x/gov/types/v1beta1 support.
func TestExpectedRegisteredLegacyProposals(t *testing.T) {
	knownProposals := map[string]struct{}{
		"/cosmos.gov.v1beta1.TextProposal":                        {},
		"/cosmos.params.v1beta1.ParameterChangeProposal":          {},
		"/cosmos.distribution.v1beta1.CommunityPoolSpendProposal": {},
		"/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal":         {},
		"/cosmos.upgrade.v1beta1.CancelSoftwareUpgradeProposal":   {},
		"/ibc.core.client.v1.UpgradeProposal":                     {},
		"/ibc.core.client.v1.ClientUpdateProposal":                {},
		"/cosmwasm.wasm.v1.PinCodesProposal":                      {},
		"/cosmwasm.wasm.v1.UpdateAdminProposal":                   {},
		"/cosmwasm.wasm.v1.StoreAndInstantiateContractProposal":   {},
		"/cosmwasm.wasm.v1.UpdateInstantiateConfigProposal":       {},
		"/cosmwasm.wasm.v1.ExecuteContractProposal":               {},
		"/cosmwasm.wasm.v1.StoreCodeProposal":                     {},
		"/cosmwasm.wasm.v1.UnpinCodesProposal":                    {},
		"/cosmwasm.wasm.v1.InstantiateContractProposal":           {},
		"/cosmwasm.wasm.v1.InstantiateContract2Proposal":          {},
		"/cosmwasm.wasm.v1.MigrateContractProposal":               {},
		"/cosmwasm.wasm.v1.SudoContractProposal":                  {},
		"/cosmwasm.wasm.v1.ClearAdminProposal":                    {},
	}

	// This is required to compile all the proposals used by the app
	simapp.New()

	var unknownProposals []string
	for protoType := range revProtoTypes {
		instance := reflect.New(protoType.Elem()).Interface()
		proposal, ok := instance.(govtypesv1beta1.Content)
		if !ok {
			continue
		}
		proposalURI := "/" + proto.MessageName(proposal.(proto.Message))

		// Skip known proposals.
		if _, exists := knownProposals[proposalURI]; exists {
			delete(knownProposals, proposalURI)
		} else {
			unknownProposals = append(unknownProposals, proposalURI)
		}
	}

	assert.Empty(t, knownProposals)
	assert.Empty(t, unknownProposals)
}
