package app_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/CoreumFoundation/coreum/v6/testutil/simapp"
)

func TestAppStateExport(t *testing.T) {
	simApp := simapp.New()
	require.NoError(t, simApp.FinalizeBlock())

	exportedApp, err := simApp.ExportAppStateAndValidators(
		false, nil, nil,
	)
	require.NoError(t, err)
	require.NotNil(t, exportedApp.AppState)
}
