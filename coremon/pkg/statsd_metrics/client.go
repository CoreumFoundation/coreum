package statsd_metrics

import (
	"sync"
	"time"

	statsd "github.com/alexcesaro/statsd"
	"github.com/pkg/errors"
)

var (
	client    Statter
	clientMux = new(sync.RWMutex)
	config    *StatterConfig
)

type StatterConfig struct {
	EnvName  string
	HostName string
}

func (m *StatterConfig) BaseTags() []string {
	var baseTags []string

	if len(config.EnvName) > 0 {
		baseTags = append(baseTags, "env", config.EnvName)
	}
	if len(config.HostName) > 0 {
		baseTags = append(baseTags, "machine", config.HostName)
	}

	return baseTags
}

func (m *StatterConfig) BaseTagsMap() map[string]string {
	baseTags := make(map[string]string, 2)

	if len(config.EnvName) > 0 {
		baseTags["env"] = config.EnvName
	}
	if len(config.HostName) > 0 {
		baseTags["machine"] = config.HostName
	}

	return baseTags
}

type Statter interface {
	Count(name string, value interface{}, tags []string) error
	Incr(name string, tags []string) error
	Decr(name string, tags []string) error
	Gauge(name string, value interface{}, tags []string) error
	Timing(name string, value time.Duration, tags []string) error
	Histogram(name string, value interface{}, tags []string) error
	Close() error
}

func Close() {
	clientMux.RLock()
	defer clientMux.RUnlock()
	if client == nil {
		return
	}
	client.Close()
}

func Disable() {
	config = checkConfig(nil)
	clientMux.Lock()
	client = newNoopStatter()
	clientMux.Unlock()
}

func Init(addr string, prefix string, cfg *StatterConfig) error {
	config = checkConfig(cfg)

	var (
		statter Statter
		err     error
	)

	statter, err = newTelegrafStatter(
		statsd.Address(addr),
		statsd.Prefix(prefix),
		statsd.ErrorHandler(errHandlerNoop),
		statsd.TagsFormat(statsd.InfluxDB),
		statsd.Tags(config.BaseTags()...),
	)
	if err != nil {
		err = errors.Wrap(err, "statsd init failed")
		return err
	}

	clientMux.Lock()
	client = statter
	clientMux.Unlock()

	return nil
}

func checkConfig(cfg *StatterConfig) *StatterConfig {
	if cfg == nil {
		cfg = &StatterConfig{}
	}

	if len(cfg.EnvName) == 0 {
		cfg.EnvName = "local"
	}

	return cfg
}

func errHandlerNoop(err error) {
}
