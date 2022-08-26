package wasm

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/types"
	"github.com/CoreumFoundation/coreum/pkg/wasm"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	_ "embed"
)

var (
	//go:embed contracts/simple-state/artifacts/simple_state.wasm
	simpleStateWASM []byte
)

// TestSimpleStateContract runs a contract deployment flow and tries to modify the state after deployment.
// This is a E2E check for the WASM integration, to ensure it works for a simple state contract (Counter).
func TestSimpleStateContract(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
	var (
		adminWallet        = testing.RandomWallet()
		networkConfig      wasm.ChainConfig
		stagedContractPath string
	)

	minGasPrice := chain.Network.InitialGasPrice()
	nativeDenom := chain.Network.TokenSymbol()
	nativeTokens := func(v string) string {
		return v + nativeDenom
	}

	initTestState := func(ctx context.Context) error {
		orPanic(chain.Network.FundAccount(adminWallet.Key.PubKey(), nativeTokens("100000000000000000000000000000000000")))
		networkConfig = wasm.ChainConfig{
			ChainID: string(chain.Network.ChainID()),
			// FIXME: Take this value from Network.InitialGasPrice() once Milad integrates it into crust
			MinGasPrice: nativeTokens(minGasPrice.String()),
			RPCEndpoint: chain.RPCAddr,
		}

		// FIXME: if workdir for the test is fixed, we can avoid embedding & staging
		// the artefacts. Should be just referencing the local file.

		stagedContractsDir := filepath.Join(os.TempDir(), "crust", "wasm", "artifacts")
		if err := os.MkdirAll(stagedContractsDir, 0700); err != nil {
			err = errors.Wrap(err, "failed to init the WASM staging dig")
			return err
		}

		stagedContractPath = filepath.Join(stagedContractsDir, "simple_state.wasm")
		if err := ioutil.WriteFile(stagedContractPath, simpleStateWASM, 0600); err != nil {
			err = errors.Wrap(err, "failed to stage the WASM contract for the test")
			return err
		}

		return nil
	}

	runTestFunc := func(ctx context.Context, t testing.T) {
		testSimpleStateContract(
			adminWallet,
			networkConfig,
			stagedContractPath,
		)(ctx, t)
	}

	return initTestState, runTestFunc
}

func testSimpleStateContract(
	adminWallet types.Wallet,
	networkConfig wasm.ChainConfig,
	stagedContractPath string,
) func(context.Context, testing.T) {
	return func(ctx context.Context, t testing.T) {
		expect := require.New(t)

		// Store the contract code on the chain, instantiation will be done in later step

		deployOut, err := wasm.Deploy(ctx, wasm.DeployConfig{
			Network: networkConfig,
			From:    adminWallet,

			ArtefactPath: stagedContractPath,
		})
		expect.NoError(err)
		expect.NotEmpty(deployOut.StoreTxHash)

		// Instantiate the contract and set the initial counter state
		// This step could be done within previous step, but separated there
		// so we could chech the intermediate result of code storage.

		deployOut, err = wasm.Deploy(ctx, wasm.DeployConfig{
			Network: networkConfig,
			From:    adminWallet,

			ArtefactPath: stagedContractPath,
			InstantiationConfig: wasm.ContractInstanceConfig{
				NeedInstantiation:  true,
				InstantiatePayload: `{"count": 1337}`,
			},
		})
		expect.NoError(err)
		expect.NotEmpty(deployOut.InitTxHash)
		expect.NotEmpty(deployOut.ContractAddr)

		// Query the contract state to get the initial count

		queryOut, err := wasm.Query(ctx, deployOut.ContractAddr, wasm.QueryConfig{
			Network:      networkConfig,
			QueryPayload: `{"get_count": {}}`,
		})
		expect.NoError(err)

		response := simpleStateQueryResponse{}
		err = json.Unmarshal(queryOut.Result, &response)
		expect.NoError(err)
		expect.Equal(1337, response.Count)

		// Execute contract to increment the count

		execOut, err := wasm.Execute(ctx, deployOut.ContractAddr, wasm.ExecuteConfig{
			Network:        networkConfig,
			From:           adminWallet,
			ExecutePayload: `{"increment": {}}`,
		})
		expect.NoError(err)
		expect.NotEmpty(execOut.ExecuteTxHash)
		expect.Equal(deployOut.ContractAddr, execOut.ContractAddress)
		expect.Equal("try_increment", execOut.MethodExecuted)

		// Query the contract once again to ensure the count has been incremented

		queryOut, err = wasm.Query(ctx, deployOut.ContractAddr, wasm.QueryConfig{
			Network:      networkConfig,
			QueryPayload: `{"get_count": {}}`,
		})
		expect.NoError(err)

		response = simpleStateQueryResponse{}
		err = json.Unmarshal(queryOut.Result, &response)
		expect.NoError(err)
		expect.Equal(1338, response.Count)
	}
}

type simpleStateQueryResponse struct {
	Count int `json:"count"`
}
