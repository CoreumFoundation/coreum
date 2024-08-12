package wbank

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types/module"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/gogoproto/grpc"
	"github.com/stretchr/testify/require"
	googlegrpc "google.golang.org/grpc"
)

var _ module.Configurator = &configuratorMock{}

type grpcServerMock struct{}

func (s grpcServerMock) RegisterService(sd *googlegrpc.ServiceDesc, ss interface{}) {}

type configuratorMock struct {
	msgServer                 grpcServerMock
	queryServer               grpcServerMock
	capturedMigrationVersions []uint64
}

func newConfiguratorMock() *configuratorMock {
	msgServer := grpcServerMock{}
	queryServer := grpcServerMock{}

	return &configuratorMock{
		msgServer:   msgServer,
		queryServer: queryServer,
	}
}

func (c *configuratorMock) MsgServer() grpc.Server {
	return c.msgServer
}

func (c *configuratorMock) QueryServer() grpc.Server {
	return c.queryServer
}

func (c *configuratorMock) RegisterMigration(
	moduleName string, forVersion uint64, handler module.MigrationHandler,
) error {
	c.capturedMigrationVersions = append(c.capturedMigrationVersions, forVersion)
	return nil
}

func (c *configuratorMock) RegisterService(sd *googlegrpc.ServiceDesc, ss interface{}) {
}

func (c *configuratorMock) Error() error {
	return nil
}

// The test checks the migration registration of the original bank.
// Since we override the "Register Services" we want to be sure that after the update of the SDK,
// The original bank won't have unexpected migrations.
func TestAppModuleOriginalBank_RegisterServices(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}).Codec
	bankModule := bank.NewAppModule(cdc, bankkeeper.BaseKeeper{}, keeper.AccountKeeper{}, nil)
	configurator := newConfiguratorMock()
	bankModule.RegisterServices(configurator)
	require.Equal(t, []uint64{1, 2, 3}, configurator.capturedMigrationVersions)
	require.Equal(t, uint64(4), bankModule.ConsensusVersion())
}
