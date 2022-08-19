package network

import (
	"fmt"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/network"
)

type (
	// Network defines a local in-process testing network
	Network = network.Network

	// Config defines the necessary configuration used to bootstrap and start an
	// in-process local testing network
	Config = network.Config
)

// PrintLogger defines a simple logger used for the network test.
type PrintLogger struct{}

// NewPrintLogger returns a new instance of the PrintLogger.
func NewPrintLogger() *PrintLogger {
	return &PrintLogger{}
}

// Log logs information into the output.
func (l *PrintLogger) Log(args ...interface{}) {
	fmt.Println(args...)
}

// Logf logs formatter information into the output.
func (l *PrintLogger) Logf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

// New creates instance with fully configured cosmos network.
// Accepts optional config, that will be used in place of the DefaultConfig() if provided.
func New(t *testing.T, configs ...network.Config) *network.Network {
	if len(configs) > 1 {
		panic("at most one config should be provided")
	}
	var cfg network.Config
	if len(configs) == 0 {
		cfg = network.DefaultConfig()
	} else {
		cfg = configs[0]
	}

	net, err := network.New(NewPrintLogger(), tempDir(), cfg)
	if err != nil {
		panic(fmt.Sprintf("can't create new network : %s", err))
	}
	t.Cleanup(net.Cleanup)
	return net
}

func tempDir() string {
	dir := os.TempDir()
	return dir
}
