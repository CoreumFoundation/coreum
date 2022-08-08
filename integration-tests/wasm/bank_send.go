package wasm

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/CoreumFoundation/coreum/integration-tests/testing"
	"github.com/CoreumFoundation/coreum/pkg/types"
	"github.com/CoreumFoundation/crust/pkg/contracts"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"

	_ "embed"
)

var (
	//go:embed fixtures/bank-send/artifacts/bank_send.wasm
	bankSendWASM []byte
)

// TestBankSendContract runs a contract deployment flow and tests that the contract is able to use Bank module
// to dispurse the native coins.
func TestBankSendContract(chain testing.Chain) (testing.PrepareFunc, testing.RunFunc) {
	var (
		adminWallet        = testing.RandomWallet()
		testWallet         = testing.RandomWallet()
		networkConfig      contracts.ChainConfig
		stagedContractPath string
	)

	minGasPrice := chain.Network.InitialGasPrice()
	nativeDenom := chain.Network.TokenSymbol()
	nativeTokens := func(v string) string {
		return v + nativeDenom
	}

	initTestState := func(ctx context.Context) error {
		orPanic(chain.Network.FundAccount(adminWallet.Key.PubKey(), nativeTokens("100000000000000000000000000000000000")))
		orPanic(chain.Network.FundAccount(testWallet.Key.PubKey(), nativeTokens("0")))

		networkConfig = contracts.ChainConfig{
			ChainID:     string(chain.Network.ChainID()),
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

		stagedContractPath = filepath.Join(stagedContractsDir, "bank_send.wasm")
		if err := ioutil.WriteFile(stagedContractPath, bankSendWASM, 0600); err != nil {
			err = errors.Wrap(err, "failed to stage the WASM contract for the test")
			return err
		}

		return nil
	}

	runTestFunc := func(ctx context.Context, t testing.T) {
		testBankSendContract(
			chain,
			adminWallet,
			testWallet,
			networkConfig,
			stagedContractPath,
			nativeDenom,
			nativeTokens,
		)(ctx, t)
	}

	return initTestState, runTestFunc
}

func testBankSendContract(
	chain testing.Chain,
	adminWallet types.Wallet,
	testWallet types.Wallet,
	networkConfig contracts.ChainConfig,
	stagedContractPath string,
	nativeDenom string,
	nativeTokens func(string) string,
) func(context.Context, testing.T) {
	return func(ctx context.Context, t testing.T) {
		expect := require.New(t)

		deployOut, err := contracts.Deploy(ctx, contracts.DeployConfig{
			Network: networkConfig,
			From:    adminWallet,

			ArtefactPath: stagedContractPath,
			NeedRebuild:  false,
		})
		expect.NoError(err)
		expect.NotEmpty(deployOut.StoreTxHash)

		deployOut, err = contracts.Deploy(ctx, contracts.DeployConfig{
			Network: networkConfig,
			From:    adminWallet,

			ArtefactPath: stagedContractPath,
			InstantiationConfig: contracts.ContractInstanceConfig{
				NeedInstantiation:  true,
				InstantiatePayload: `{"count": 0}`,

				// transfer some coins during instantiation,
				// so we could withdraw them later using contract code.
				Amount: nativeTokens("10000"),
			},
		})
		expect.NoError(err)
		expect.NotEmpty(deployOut.InitTxHash)
		expect.NotEmpty(deployOut.ContractAddr)

		// check that the contract has the bank balance after instantiation

		contractBalance, err := chain.Client.BankQueryClient().Balance(ctx,
			&banktypes.QueryBalanceRequest{
				Address: deployOut.ContractAddr,
				Denom:   nativeDenom,
			})
		expect.NoError(err)
		expect.NotNil(contractBalance.Balance)
		expect.Equal(nativeDenom, contractBalance.Balance.Denom)
		expect.Equal("10000", contractBalance.Balance.Amount.String())

		// withdraw half of the coins to a test wallet, previously empty

		withdrawMsg := fmt.Sprintf(
			`{"withdraw": { "amount":"5000", "denom":"%s", "recipient":"%s" }}`,
			nativeDenom,
			testWallet.Address().String(),
		)

		execOut, err := contracts.Execute(ctx, deployOut.ContractAddr, contracts.ExecuteConfig{
			Network:        networkConfig,
			From:           adminWallet,
			ExecutePayload: withdrawMsg,
		})
		expect.NoError(err)
		expect.NotEmpty(execOut.ExecuteTxHash)
		expect.Equal(deployOut.ContractAddr, execOut.ContractAddress)
		expect.Equal("try_withdraw", execOut.MethodExecuted)

		// check that contract now has half of the coins

		contractBalance, err = chain.Client.BankQueryClient().Balance(ctx,
			&banktypes.QueryBalanceRequest{
				Address: deployOut.ContractAddr,
				Denom:   nativeDenom,
			})
		expect.NoError(err)
		expect.NotNil(contractBalance.Balance)
		expect.Equal(nativeDenom, contractBalance.Balance.Denom)
		expect.Equal("5000", contractBalance.Balance.Amount.String())

		// check that the target test wallet has another half

		testWalletBalance, err := chain.Client.BankQueryClient().Balance(ctx,
			&banktypes.QueryBalanceRequest{
				Address: testWallet.Address().String(),
				Denom:   nativeDenom,
			})
		expect.NoError(err)
		expect.NotNil(testWalletBalance.Balance)
		expect.Equal(nativeDenom, testWalletBalance.Balance.Denom)
		expect.Equal("5000", testWalletBalance.Balance.Amount.String())

		// bank send invoked by the contract code succeeded! 〠
	}
}

func orPanic(err error) {
	if err != nil {
		panic(err)
	}
}
