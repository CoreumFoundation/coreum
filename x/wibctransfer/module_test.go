package wibctransfer

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/gogoproto/grpc"
	"github.com/cosmos/ibc-go/v10/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v10/modules/apps/transfer/keeper"
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

// The test checks the migration registration of the original IBC transfer module.
// Since we override the "Register Services" we want to be sure that after the update of the SDK,
// The original transfer module won't have unexpected migrations.
func TestAppModuleOriginalTransfer_RegisterServices(t *testing.T) {
	transferModule := transfer.NewAppModule(ibctransferkeeper.Keeper{})
	configurator := newConfiguratorMock()
	transferModule.RegisterServices(configurator)
	require.Equal(t, []uint64{2, 3, 4, 5}, configurator.capturedMigrationVersions)
	require.Equal(t, uint64(6), transferModule.ConsensusVersion())
}
