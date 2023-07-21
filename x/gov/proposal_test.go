package gov_test

import (
	"reflect"
	"testing"
	_ "unsafe"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"

	"github.com/CoreumFoundation/coreum/v2/testutil/simapp"
)

// To access private variable from github.com/gogo/protobuf we link it to local variable.
// This is needed to iterate through all registered protobuf types.
//
//go:linkname revProtoTypes github.com/gogo/protobuf/proto.revProtoTypes
var revProtoTypes map[reflect.Type]string

func TestExpectedRegisteredProposals(t *testing.T) {
	knownProposals := map[string]struct{}{
		// proposals we have integration tests for

		"/cosmos.gov.v1beta1.TextProposal":                        {},
		"/cosmos.params.v1beta1.ParameterChangeProposal":          {},
		"/cosmos.distribution.v1beta1.CommunityPoolSpendProposal": {},
		"/cosmos.upgrade.v1beta1.SoftwareUpgradeProposal":         {},
		"/cosmwasm.wasm.v1.PinCodesProposal":                      {},
		"/cosmwasm.wasm.v1.UnpinCodesProposal":                    {},

		// proposals without tests

		"/cosmos.upgrade.v1beta1.CancelSoftwareUpgradeProposal": {},
		"/cosmwasm.wasm.v1.StoreAndInstantiateContractProposal": {},
		"/cosmwasm.wasm.v1.UpdateInstantiateConfigProposal":     {},
		"/cosmwasm.wasm.v1.InstantiateContractProposal":         {},
		"/cosmwasm.wasm.v1.SudoContractProposal":                {},
		"/cosmwasm.wasm.v1.MigrateContractProposal":             {},
		"/cosmwasm.wasm.v1.ClearAdminProposal":                  {},
		"/cosmwasm.wasm.v1.ExecuteContractProposal":             {},
		"/cosmwasm.wasm.v1.UpdateAdminProposal":                 {},
		"/cosmwasm.wasm.v1.StoreCodeProposal":                   {},
		"/ibc.core.client.v1.UpgradeProposal":                   {},
		"/ibc.core.client.v1.ClientUpdateProposal":              {},
	}

	// This is required to compile all the proposals used by the app
	simapp.New()

	var unknownProposals []string
	for protoType := range revProtoTypes {
		instance := reflect.New(protoType.Elem()).Interface()
		proposal, ok := instance.(govtypes.Content)
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
