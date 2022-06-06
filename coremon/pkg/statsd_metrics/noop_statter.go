package statsd_metrics

import (
	"time"
)

func newNoopStatter() Statter {
	return &noopStatter{}
}

type noopStatter struct {
}

func (s *noopStatter) Count(name string, value interface{}, tags []string) error {
	return nil
}

func (s *noopStatter) Incr(name string, tags []string) error {
	return nil
}

func (s *noopStatter) Decr(name string, tags []string) error {
	return nil
}

func (s *noopStatter) Gauge(name string, value interface{}, tags []string) error {
	return nil
}

func (s *noopStatter) Timing(name string, value time.Duration, tags []string) error {
	return nil
}

func (s *noopStatter) Histogram(name string, value interface{}, tags []string) error {
	return nil
}

func (s *noopStatter) Unique(bucket string, value string) error {
	return nil
}

func (s *noopStatter) Close() error {
	return nil
}
